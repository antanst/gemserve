package server

import (
	"crypto/tls"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gemserve/common"
	"gemserve/config"
	"gemserve/errors"
	"gemserve/logging"
	"github.com/gabriel-vasile/mimetype"
)

type ServerConfig interface {
	DirIndexingEnabled() bool
	RootPath() string
}

func GenerateResponse(conn *tls.Conn, connId string, input string) ([]byte, error) {
	trimmedInput := strings.TrimSpace(input)
	// url will have a cleaned and normalized path after this
	url, err := common.ParseURL(trimmedInput, "", true)
	if err != nil {
		return nil, errors.NewConnectionError(fmt.Errorf("failed to parse URL: %w", err))
	}
	logging.LogDebug("%s %s normalized URL path: %s", connId, conn.RemoteAddr(), url.Path)
	serverRootPath := config.CONFIG.RootPath
	localPath, err := calculateLocalPath(url.Path, serverRootPath)
	if err != nil {
		return nil, errors.NewConnectionError(err)
	}
	logging.LogDebug("%s %s request file path: %s", connId, conn.RemoteAddr(), localPath)

	// Get file/directory information
	info, err := os.Stat(localPath)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return []byte("51 not found\r\n"), nil
	} else if err != nil {
		return nil, errors.NewConnectionError(fmt.Errorf("%s %s failed to access path: %w", connId, conn.RemoteAddr(), err))
	}

	// Handle directory.
	if info.IsDir() {
		return generateResponseDir(conn, connId, url, localPath)
	}
	return generateResponseFile(conn, connId, url, localPath)
}

func generateResponseFile(conn *tls.Conn, connId string, url *common.URL, localPath string) ([]byte, error) {
	data, err := os.ReadFile(localPath)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return []byte("51 not found\r\n"), nil
	} else if err != nil {
		return nil, errors.NewConnectionError(fmt.Errorf("%s %s failed to read file: %w", connId, conn.RemoteAddr(), err))
	}

	var mimeType string
	if path.Ext(localPath) == ".gmi" {
		mimeType = "text/gemini"
	} else {
		mimeType = mimetype.Detect(data).String()
	}
	headerBytes := []byte(fmt.Sprintf("20 %s\r\n", mimeType))
	response := append(headerBytes, data...)
	return response, nil
}

func generateResponseDir(conn *tls.Conn, connId string, url *common.URL, localPath string) (output []byte, err error) {
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return nil, errors.NewConnectionError(fmt.Errorf("%s %s failed to read directory: %w", connId, conn.RemoteAddr(), err))
	}

	if config.CONFIG.DirIndexingEnabled {
		var contents []string
		contents = append(contents, "Directory index:\n\n")
		contents = append(contents, "=> ../\n")
		for _, entry := range entries {
			if entry.IsDir() {
				contents = append(contents, fmt.Sprintf("=> %s/\n", entry.Name()))
			} else {
				contents = append(contents, fmt.Sprintf("=> %s\n", entry.Name()))
			}
		}
		data := []byte(strings.Join(contents, ""))
		headerBytes := []byte("20 text/gemini;\r\n")
		response := append(headerBytes, data...)
		return response, nil
	} else {
		filePath := path.Join(localPath, "index.gmi")
		return generateResponseFile(conn, connId, url, filePath)

	}
}

func calculateLocalPath(input string, basePath string) (string, error) {
	// Check for invalid characters early
	if strings.ContainsAny(input, "\\") {
		return "", errors.NewError(fmt.Errorf("invalid characters in path: %s", input))
	}

	// If IsLocal(path) returns true, then Join(base, path)
	// will always produce a path contained within base and
	// Clean(path) will always produce an unrooted path with
	// no ".." path elements.
	filePath := input
	filePath = strings.TrimPrefix(filePath, "/")
	if filePath == "" {
		filePath = "."
	}
	filePath = strings.TrimSuffix(filePath, "/")

	localPath, err := filepath.Localize(filePath)
	if err != nil || !filepath.IsLocal(localPath) {
		return "", errors.NewError(fmt.Errorf("could not construct local path from %s: %s", input, err))
	}

	filePath = path.Join(basePath, localPath)
	return filePath, nil
}

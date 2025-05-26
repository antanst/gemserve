package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gemserve/common"
	"gemserve/config"
	logging "git.antanst.com/antanst/logging"
	"git.antanst.com/antanst/xerrors"
	"github.com/gabriel-vasile/mimetype"
)

type ServerConfig interface {
	DirIndexingEnabled() bool
	RootPath() string
}

func checkRequestURL(url *common.URL) error {
	if url.Protocol != "gemini" {
		return xerrors.NewError(fmt.Errorf("invalid URL"), 53, "URL Protocol not Gemini, proxying refused", false)
	}

	_, portStr, err := net.SplitHostPort(config.CONFIG.Listen)
	if err != nil {
		return xerrors.NewError(fmt.Errorf("failed to parse listen address: %w", err), 50, "Server configuration error", false)
	}
	listenPort, err := strconv.Atoi(portStr)
	if err != nil {
		return xerrors.NewError(fmt.Errorf("failed to parse listen port: %w", err), 50, "Server configuration error", false)
	}
	if url.Port != listenPort {
		return xerrors.NewError(fmt.Errorf("failed to parse URL: %w", err), 53, "invalid URL port, proxying refused", false)
	}
	return nil
}

func GenerateResponse(conn *tls.Conn, connId string, input string) ([]byte, error) {
	trimmedInput := strings.TrimSpace(input)
	// url will have a cleaned and normalized path after this
	url, err := common.ParseURL(trimmedInput, "", true)
	if err != nil {
		return nil, xerrors.NewError(fmt.Errorf("failed to parse URL: %w", err), 59, "Invalid URL", false)
	}
	logging.LogDebug("%s %s normalized URL path: %s", connId, conn.RemoteAddr(), url.Path)

	err = checkRequestURL(url)
	if err != nil {
		return nil, err
	}

	serverRootPath := config.CONFIG.RootPath
	localPath, err := calculateLocalPath(url.Path, serverRootPath)
	if err != nil {
		return nil, xerrors.NewError(err, 59, "Invalid path", false)
	}
	logging.LogDebug("%s %s request file path: %s", connId, conn.RemoteAddr(), localPath)

	// Get file/directory information
	info, err := os.Stat(localPath)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		return []byte("51 not found\r\n"), nil
	} else if err != nil {
		return nil, xerrors.NewError(fmt.Errorf("%s %s failed to access path: %w", connId, conn.RemoteAddr(), err), 0, "Path access failed", false)
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
		return nil, xerrors.NewError(fmt.Errorf("%s %s failed to read file: %w", connId, conn.RemoteAddr(), err), 0, "File read failed", false)
	}

	var mimeType string
	if path.Ext(localPath) == ".gmi" {
		mimeType = "text/gemini"
	} else {
		mimeType = mimetype.Detect(data).String()
	}
	headerBytes := []byte(fmt.Sprintf("20 %s; lang=en\r\n", mimeType))
	response := append(headerBytes, data...)
	return response, nil
}

func generateResponseDir(conn *tls.Conn, connId string, url *common.URL, localPath string) (output []byte, err error) {
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return nil, xerrors.NewError(fmt.Errorf("%s %s failed to read directory: %w", connId, conn.RemoteAddr(), err), 0, "Directory read failed", false)
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
		headerBytes := []byte("20 text/gemini; lang=en\r\n")
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
		return "", xerrors.NewError(fmt.Errorf("invalid characters in path: %s", input), 0, "Invalid path characters", false)
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
		return "", xerrors.NewError(fmt.Errorf("could not construct local path from %s: %s", input, err), 0, "Invalid local path", false)
	}

	filePath = path.Join(basePath, localPath)
	return filePath, nil
}

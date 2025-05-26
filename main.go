package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gemserve/config"
	"gemserve/server"
	"gemserve/uid"
	logging "git.antanst.com/antanst/logging"
	"git.antanst.com/antanst/xerrors"
)

var fatalErrors chan error

func main() {
	config.CONFIG = *config.GetConfig()

	logging.InitSlogger(config.CONFIG.LogLevel)

	err := runApp()
	if err != nil {
		panic(fmt.Sprintf("Fatal Error: %v", err))
	}
	os.Exit(0)
}

func runApp() error {
	logging.LogInfo("Starting up. Press Ctrl+C to exit")

	listenHost := config.CONFIG.Listen

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	fatalErrors = make(chan error)

	go func() {
		err := startServer(listenHost)
		if err != nil {
			fatalErrors <- xerrors.NewError(err, 0, "Server startup failed", true)
		}
	}()

	for {
		select {
		case <-signals:
			logging.LogWarn("Received SIGINT or SIGTERM signal, exiting")
			return nil
		case fatalError := <-fatalErrors:
			return xerrors.NewError(fatalError, 0, "Server error", true)
		}
	}
}

func startServer(listenHost string) (err error) {
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		return xerrors.NewError(fmt.Errorf("failed to load certificate: %w", err), 0, "Certificate loading failed", true)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", listenHost, tlsConfig)
	if err != nil {
		return xerrors.NewError(fmt.Errorf("failed to create listener: %w", err), 0, "Listener creation failed", true)
	}
	defer func(listener net.Listener) {
		// If we've got an error closing the
		// listener, make sure we don't override
		// the original error (if not nil)
		errClose := listener.Close()
		if errClose != nil && err == nil {
			err = xerrors.NewError(err, 0, "Listener close failed", true)
		}
	}(listener)

	logging.LogInfo("Server listening on %s", listenHost)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logging.LogInfo("Failed to accept connection: %v", err)
			continue
		}

		go func() {
			remoteAddr := conn.RemoteAddr().String()
			connId := uid.UID()
			err := handleConnection(conn.(*tls.Conn), connId, remoteAddr)
			if err != nil {
				var asErr *xerrors.XError
				if errors.As(err, &asErr) && asErr.IsFatal {
					fatalErrors <- asErr
					return
				} else {
					logging.LogWarn("%s %s Connection failed: %d %s (%v)", connId, remoteAddr, asErr.Code, asErr.UserMsg, err)
				}
			}
		}()
	}
}

func closeConnection(conn *tls.Conn) error {
	err := conn.CloseWrite()
	if err != nil {
		return xerrors.NewError(fmt.Errorf("failed to close TLS connection: %w", err), 50, "Connection close failed", false)
	}
	err = conn.Close()
	if err != nil {
		return xerrors.NewError(fmt.Errorf("failed to close connection: %w", err), 50, "Connection close failed", false)
	}
	return nil
}

func handleConnection(conn *tls.Conn, connId string, remoteAddr string) (err error) {
	start := time.Now()
	var outputBytes []byte

	defer func(conn *tls.Conn) {
		end := time.Now()
		tookMs := end.Sub(start).Milliseconds()
		var responseHeader string

		// On non-errors, just log response and close connection.
		if err == nil {
			// Log non-erroneous responses
			if i := bytes.Index(outputBytes, []byte{'\r'}); i >= 0 {
				responseHeader = string(outputBytes[:i])
			}
			logging.LogInfo("%s %s response %s (%dms)", connId, remoteAddr, responseHeader, tookMs)
			_ = closeConnection(conn)
			return
		}

		var code int
		var responseMsg string
		var xErr *xerrors.XError
		if errors.As(err, &xErr) {
			// On fatal errors, immediatelly return the error.
			if xErr.IsFatal {
				_ = closeConnection(conn)
				return
			}
			code = xErr.Code
			responseMsg = xErr.UserMsg
		} else {
			code = 50
			responseMsg = "server error"
		}
		responseHeader = fmt.Sprintf("%d %s", code, responseMsg)
		_, _ = conn.Write([]byte(responseHeader))
		_ = closeConnection(conn)
	}(conn)

	// Gemini is supposed to have a 1kb limit
	// on input requests.
	buffer := make([]byte, 1025)

	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		return xerrors.NewError(fmt.Errorf("failed to read connection data: %w", err), 59, "Connection read failed", false)
	}
	if n == 0 {
		return xerrors.NewError(fmt.Errorf("client did not send data"), 59, "No data received", false)
	}
	if n > 1024 {
		return xerrors.NewError(fmt.Errorf("client request size %d > 1024 bytes", n), 59, "Request too large", false)
	}

	dataBytes := buffer[:n]
	dataString := string(dataBytes)

	logging.LogInfo("%s %s request %s (%d bytes)", connId, remoteAddr, strings.TrimSpace(dataString), len(dataBytes))
	outputBytes, err = server.GenerateResponse(conn, connId, dataString)
	if err != nil {
		return err
	}
	_, err = conn.Write(outputBytes)
	if err != nil {
		return err
	}
	return nil
}

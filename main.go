package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gemserve/config"
	"gemserve/errors"
	"gemserve/logging"
	"gemserve/server"
	"gemserve/uid"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func main() {
	config.CONFIG = *config.GetConfig()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(config.CONFIG.LogLevel)
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "[2006-01-02 15:04:05]"})
	err := runApp()
	if err != nil {
		fmt.Printf("%v\n", err)
		logging.LogError("%v", err)
		os.Exit(1)
	}
}

func runApp() error {
	logging.LogInfo("Starting up. Press Ctrl+C to exit")

	var listenHost string
	if len(os.Args) != 2 {
		listenHost = "0.0.0.0:1965"
	} else {
		listenHost = os.Args[1]
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error)

	go func() {
		err := startServer(listenHost)
		if err != nil {
			serverErrors <- errors.NewFatalError(err)
		}
	}()

	for {
		select {
		case <-signals:
			logging.LogWarn("Received SIGINT or SIGTERM signal, exiting")
			return nil
		case serverError := <-serverErrors:
			return errors.NewFatalError(serverError)
		}
	}
}

func startServer(listenHost string) (err error) {
	cert, err := tls.LoadX509KeyPair("/certs/cert", "/certs/key")
	if err != nil {
		return errors.NewFatalError(fmt.Errorf("failed to load certificate: %w", err))
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", listenHost, tlsConfig)
	if err != nil {
		return errors.NewFatalError(fmt.Errorf("failed to create listener: %w", err))
	}
	defer func(listener net.Listener) {
		// If we've got an error closing the
		// listener, make sure we don't override
		// the original error (if not nil)
		errClose := listener.Close()
		if errClose != nil && err == nil {
			err = errors.NewFatalError(err)
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
			err := handleConnection(conn.(*tls.Conn))
			if err != nil {
				var asErr *errors.Error
				if errors.As(err, &asErr) {
					logging.LogError("Unexpected error: %v %v", err, err.(*errors.Error).ErrorWithStack())
				} else {
					logging.LogError("Unexpected error: %v", err)
				}
				if config.CONFIG.PanicOnUnexpectedError {
					panic("Encountered unexpected error")
				}
			}
		}()
	}
}

func closeConnection(conn *tls.Conn) error {
	err := conn.CloseWrite()
	if err != nil {
		return errors.NewConnectionError(fmt.Errorf("failed to close TLS connection: %w", err))
	}
	err = conn.Close()
	if err != nil {
		return errors.NewConnectionError(fmt.Errorf("failed to close connection: %w", err))
	}
	return nil
}

func handleConnection(conn *tls.Conn) (err error) {
	remoteAddr := conn.RemoteAddr().String()
	connId := uid.UID()
	start := time.Now()
	var outputBytes []byte

	defer func(conn *tls.Conn) {
		// Three possible cases here:
		// - We don't have an error
		// - We have a ConnectionError, which we don't propagate up
		// - We have an unexpected error.
		end := time.Now()
		tookMs := end.Sub(start).Milliseconds()
		var responseHeader string
		if err != nil {
			_, _ = conn.Write([]byte("50 server error"))
			responseHeader = "50 server error"
			// We don't propagate connection errors up.
			if errors.Is(err, errors.ConnectionError) {
				logging.LogInfo("%s %s %v", connId, remoteAddr, err)
				err = nil
			}
		} else {
			if i := bytes.Index(outputBytes, []byte{'\r'}); i >= 0 {
				responseHeader = string(outputBytes[:i])
			}
		}
		logging.LogInfo("%s %s response %s (%dms)", connId, remoteAddr, responseHeader, tookMs)
		_ = closeConnection(conn)
	}(conn)

	// Gemini is supposed to have a 1kb limit
	// on input requests.
	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		return errors.NewConnectionError(fmt.Errorf("failed to read connection data: %w", err))
	}
	if n == 0 {
		return errors.NewConnectionError(fmt.Errorf("client did not send data"))
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

package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	StatusInput            = 10
	StatusSuccess          = 20
	StatusRedirectTemp     = 30
	StatusTemporaryFailure = 40
	StatusPermanentFailure = 50
)

type GeminiHandler func(path string, conn io.ReadWriteCloser) error

type GeminiServer struct {
	server net.Listener
}

func (srv *GeminiServer) Close() error {
	return srv.server.Close()
}

func (srv *GeminiServer) Loop(handler GeminiHandler) error {
	for {
		conn, err := srv.server.Accept()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				fmt.Fprintf(os.Stderr, "accept error: %s\n", err)
				continue
			}
		}
		go handleConnection(conn, handler)
	}
}

func CreateGeminiServer(hostname string, bindAddr string, cert tls.Certificate) (GeminiServer, error) {
	var srv GeminiServer
	var err error
	// TLS session
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}, ServerName: hostname, MinVersion: tls.VersionTLS12}
	srv.server, err = tls.Listen("tcp", bindAddr, cfg)
	return srv, err
}

func SendResponse(conn io.WriteCloser, statusCode int, meta string) error {
	header := fmt.Sprintf("%d %s\r\n", statusCode, meta)
	_, err := conn.Write([]byte(header))
	return err
}

func SendContent(conn io.WriteCloser, content []byte, meta string) error {
	if err := SendResponse(conn, StatusSuccess, meta); err != nil {
		return err
	}
	_, err := conn.Write(content)
	return err
}

// sanitize an input path, ignores all characters that are not alphanumeric or a path separator /
func sanitizePath(path string) (string, error) {
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("traverse path not allowed")
	}
	return path, nil
}

func handleConnection(conn io.ReadWriteCloser, handler GeminiHandler) error {
	defer conn.Close()

	// 1500 matches the typical MTU size
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if n > 1024 {
		return SendResponse(conn, StatusPermanentFailure, "Request exceeds maximum permitted length")
	}
	// Parse incoming request URL.
	surl := strings.TrimSpace(string(buf[:n]))
	reqURL, err := url.Parse(surl)
	if err != nil {
		return SendResponse(conn, StatusPermanentFailure, "URL incorrectly formatted")
	}

	// If the URL ends with a '/', serve the index.gmi
	var reqPath string
	if strings.HasSuffix(reqURL.Path, "/") || reqURL.Path == "" {
		reqPath = filepath.Join(reqURL.Path, "index.gmi")
	} else {
		reqPath = reqURL.Path
	}

	// Note: filepath.Clean prevents path traversal attacks
	cleanPath, err := sanitizePath(filepath.Clean(reqPath))
	if err != nil {
		return SendResponse(conn, StatusPermanentFailure, "Invalid path")
	}

	return handler(cleanPath, conn)
}

package main

import (
	"crypto/tls"
	"fmt"
	"io"
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

func ServerGemini(hostname string, bindAddr string, cert tls.Certificate, handler GeminiHandler) error {
	// TLS session
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}, ServerName: hostname, MinVersion: tls.VersionTLS12}
	listener, err := tls.Listen("tcp", bindAddr, cfg)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Fprintf(os.Stderr, "accept error: %s\n", err)
				continue
			}
		}
		go handleConnection(conn, handler)
	}
	return nil
}

func sendResponse(conn io.ReadWriteCloser, statusCode int, meta string) error {
	header := fmt.Sprintf("%d %s\r\n", statusCode, meta)
	_, err := conn.Write([]byte(header))
	return err
}

func handleConnection(conn io.ReadWriteCloser, handler GeminiHandler) error {
	defer conn.Close()

	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if n > 1024 {
		return sendResponse(conn, statusPermanentFailure, "Request exceeds maximum permitted length")
	}
	// Parse incoming request URL.
	reqURL, err := url.Parse(string(buf))
	if err != nil {
		return sendResponse(conn, statusPermanentFailure, "URL incorrectly formatted")
	}

	// If the URL ends with a '/' character, assume that the user wants the index.gmi
	// file in the corresponding directory.
	var reqPath string
	if strings.HasSuffix(reqURL.Path, "/") || reqURL.Path == "" {
		reqPath = filepath.Join(reqURL.Path, "index.gmi")
	} else {
		reqPath = reqURL.Path
	}

	cleanPath := filepath.Clean(reqPath)
	return handler(cleanPath, conn)
}

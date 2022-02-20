package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"strings"
	"testing"
	"time"
)

func generateCerts(size int) tls.Certificate {
	key, err := rsa.GenerateKey(rand.Reader, size)

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	serial, err := rand.Int(rand.Reader, big.NewInt(65535))
	if err != nil {
		panic(err)
	}
	tml := x509.Certificate{
		// you can add any attr that you need
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 0, 0),
		// you have to generate a different serial number each execution
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"Internet Widgets Corporation"},
		},
		BasicConstraintsValid: true,
	}
	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		panic(err)
	}
	return tlsCert
}

func TestServer(t *testing.T) {
	t.Logf("Generating TLS certificate ... ")
	cert := generateCerts(2048)

	server, err := CreateGeminiServer("localhost", ":1965", cert)
	if err != nil {
		t.Errorf("Error creating gemini server: %s", err)
		return
	}
	defer server.Close()

	go func() {
		if err := server.Loop(testHandler); err != nil {
			t.Errorf("server loop error: %s\n", err)
		}
	}()

	// Do server request
	req, err := Gemini("localhost:1965", "/")
	if err != nil {
		t.Errorf("Error initializing gemini request: %s", err)
		return
	}
	if req.Do() != nil {
		t.Errorf("Error performing gemini request: %s", err)
		return
	}
	if req.Status != StatusSuccess {
		t.Errorf("gemini request returned status %d", req.Status)
		return
	}
	buf := make([]byte, 1500)
	n, err := req.Read(buf)
	if err != nil {
		t.Errorf("Error reading from gemini server: %s", err)
		return
	} else {
		// Check response
		resp := strings.TrimSpace(string(buf[:n]))
		if resp != "test OK" {
			fmt.Errorf("Invalid response from gemini server")
			return
		}
		fmt.Println("Response OK")
	}
	defer req.Close()
}

func testHandler(path string, conn io.ReadWriteCloser) error {
	SendResponse(conn, StatusSuccess, "text/txt")
	conn.Write([]byte("test OK"))
	return nil
}

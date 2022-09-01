package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var config Config

func DirectoryExists(dirpath string) bool {
	stat, err := os.Stat(dirpath)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func FileExists(dirpath string) bool {
	// Quick and dirty check
	_, err := os.Stat(dirpath)
	return err == nil
}

// try to load the given config file. Ignores the file if not present. Exits the program on failure
func tryLoadConfig(filename string) {
	if FileExists(filename) {
		if err := config.LoadConfigFile(filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %s\n", filename, err)
			os.Exit(1)
		}
	}
}

func chroot(dir string) error {
	runtime.LockOSThread()
	if err := syscall.Chroot(dir); err != nil {
		return err
	}
	if err := os.Chdir("/"); err != nil {
		return err
	}
	return nil
}

func setuid(uid int) error {
	if err := syscall.Setuid(uid); err != nil {
		return err
	}
	if syscall.Getuid() != uid {
		return fmt.Errorf("getuid != %d", uid)
	}
	return nil
}

func setgid(gid int) error {
	if err := syscall.Setgid(gid); err != nil {
		return err
	}
	if syscall.Getgid() != gid {
		return fmt.Errorf("getgid != %d", gid)
	}
	return nil
}

func generateCertificate() (tls.Certificate, error) {
	// Always use crypto/rand!
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	tml := x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0), // 5 years
		SerialNumber: big.NewInt(123456),
		Subject: pkix.Name{
			CommonName:   config.Hostname,
			Organization: []string{"orion Inc."},
		},
		BasicConstraintsValid: true,
	}
	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	return tlsCert, err
}

func main() {
	config.SetDefaults()
	config.LoadEnv()

	customConfigFile := flag.String("config", "", "Configuration file")
	genCert := flag.Bool("gen-cert", false, "Generate certificate on startup")
	flag.Parse()

	if *customConfigFile != "" {
		if err := config.LoadConfigFile(*customConfigFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %s\n", *customConfigFile, err)
			os.Exit(1)
		}
	} else {
		// Load default configuration files, if existing
		tryLoadConfig("/etc/orion.conf")
		tryLoadConfig("./orion.conf")
	}

	var cert tls.Certificate
	var err error
	if *genCert {
		fmt.Fprintf(os.Stderr, "Generate server certificate ... \n")
		cert, err = generateCertificate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating tls certificate: %s\n", err)
			os.Exit(2)
		}
	} else {
		// Load keys before chroot
		if !FileExists(config.Keyfile) {
			fmt.Fprintf(os.Stderr, "Server key file not found: %s\n", config.Keyfile)
			os.Exit(1)
		}
		if !FileExists(config.CertFile) {
			fmt.Fprintf(os.Stderr, "Certificate file not found: %s\n", config.CertFile)
			os.Exit(1)
		}
		cert, err = tls.LoadX509KeyPair(config.CertFile, config.Keyfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "certificate error: %s\n", err)
			os.Exit(1)
		}
	}

	/* Drop privileges here*/

	// Chroot, if configured to do so
	if config.Chroot != "" {
		if err := chroot(config.Chroot); err != nil {
			fmt.Fprintf(os.Stderr, "chroot failed: %s\n", err)
			os.Exit(1)
		}
	}
	// Setuid and setgid if configured
	if config.Gid > 0 {
		if err := setgid(config.Gid); err != nil {
			fmt.Fprintf(os.Stderr, "setgid failed: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("gid = %d\n", syscall.Getgid())
	}
	if config.Uid > 0 {
		if err := setuid(config.Uid); err != nil {
			fmt.Fprintf(os.Stderr, "setuid failed: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("uid = %d\n", syscall.Getuid())
	}

	// Make the content dir absolute
	if !strings.HasPrefix(config.ContentDir, "/") {
		workDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting the work directory: %s\n", err)
			os.Exit(1)
		}
		config.ContentDir = workDir + "/" + config.ContentDir
	}
	// Terminate content directory with a '/'
	if !strings.HasSuffix(config.ContentDir, "/") {
		config.ContentDir += "/"
	}

	// Content warnings should point user at wrong configuration early in the program
	if !DirectoryExists(config.ContentDir) {
		fmt.Fprintf(os.Stderr, "WARNING: Content directory does not exist: %s\n", config.ContentDir)
	} else {
		if !FileExists(config.ContentDir + "/index.gmi") {
			fmt.Fprintf(os.Stderr, "WARNING: index.gmi does not exists in content directory: %s/index.gmi\n", config.ContentDir)
		}
	}

	// Setup gemini server
	server, err := CreateGeminiServer(config.Hostname, config.BindAddr, cert)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server error: %s\n", err)
		os.Exit(1)
	}

	// Termination signal handling (SIGINT or SIGTERM)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		os.Exit(2)
	}()

	log.Printf("Serving %s on %s\n", config.Hostname, config.BindAddr)
	if err := server.Loop(geminiHandle); err != nil {
		fmt.Fprintf(os.Stderr, "server loop error: %s\n", err)
		os.Exit(5)
	}
}

func geminiHandle(path string, conn io.ReadWriteCloser) error {
	log.Printf("GET %s", path)
	if f, err := os.OpenFile(config.ContentDir+"/"+path, os.O_RDONLY, 0400); err != nil {
		if err == os.ErrNotExist {
			log.Printf("ERR: File not found: %s", path)
			return SendResponse(conn, StatusPermanentFailure, "Resource not found")
		} else {
			log.Printf("ERR: File error: %s", err)
			return SendResponse(conn, StatusPermanentFailure, "Resource error")
		}
	} else {
		defer f.Close()
		// Send file segment by segment (4k segments)
		buf := make([]byte, 4096)

		// MIME handling on first segment if required
		if strings.HasSuffix(path, ".gmi") {
			mime := "text/gemini; lang=en; charset=utf-8"
			if err := SendResponse(conn, StatusSuccess, mime); err != nil {
				return err
			}
		} else {
			n, err := f.Read(buf)
			if err != nil {
				log.Printf("ERR: File read error: %s", err)
				return SendResponse(conn, StatusPermanentFailure, "Resource error")
			}
			if n == 0 {
				return SendResponse(conn, StatusSuccess, "")
			}
			mime := http.DetectContentType(buf[:n])
			if err := SendResponse(conn, StatusSuccess, mime); err != nil {
				return err
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return err
			}
		}
		// Send remaining file content
		for {
			n, err := f.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return err
			}
		}
	}
	return nil
}

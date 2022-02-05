package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

func main() {
	config.SetDefaults()

	customConfigFile := flag.String("config", "", "Configuration file")
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

	// Check settings
	if !FileExists(config.Keyfile) {
		fmt.Fprintf(os.Stderr, "Server key file not found: %s\n", config.Keyfile)
		os.Exit(1)
	}
	if !FileExists(config.CertFile) {
		fmt.Fprintf(os.Stderr, "Certificate file not found: %s\n", config.CertFile)
		os.Exit(1)
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
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.Keyfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "certificate error: %s\n", err)
		os.Exit(1)
	}
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

		// TODO: Replace by proper stream handling. For now it's good because I only serve small files
		content, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("ERR: File read error: %s", err)
			return SendResponse(conn, StatusPermanentFailure, "Resource error")
		}

		// Get MIME
		var mime string
		if strings.HasSuffix(path, ".gmi") {
			mime = "text/gemini; lang=en; charset=utf-8"
		} else {
			mime = http.DetectContentType(content)
		}
		return SendContent(conn, content, mime)
	}
}

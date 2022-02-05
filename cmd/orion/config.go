package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Hostname   string // Server hostname
	CertFile   string // Certificate filename
	Keyfile    string // Key file
	BindAddr   string // Optional binding address
	ContentDir string // Gemini content directory to serve
}

func (cf *Config) SetDefaults() {
	cf.Hostname = "localhost"
	cf.CertFile = "orion.crt"
	cf.Keyfile = "orion.key"
	cf.BindAddr = ":1967"
	cf.ContentDir = "gemini/"
}

func (cf *Config) LoadConfigFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		i := strings.Index(line, "=")
		if i < 0 {
			return fmt.Errorf("Syntax error in line %d", lineCount)
		}
		name, value := strings.ToLower(strings.TrimSpace(line[:i])), strings.TrimSpace(line[i+1:])
		if name == "hostname" {
			cf.Hostname = value
		} else if name == "certfile" {
			cf.CertFile = value
		} else if name == "keyfile" {
			cf.Keyfile = value
		} else if name == "bind" {
			cf.BindAddr = value
		} else if name == "contentdir" {
			cf.ContentDir = value
		} else {
			return fmt.Errorf("Unknown setting in line %d", lineCount)
		}

	}
	return file.Close()
}

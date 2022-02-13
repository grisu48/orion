package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Hostname   string // Server hostname
	CertFile   string // Certificate filename
	Keyfile    string // Key file
	BindAddr   string // Optional binding address
	ContentDir string // Gemini content directory to serve
	Chroot     string // chroot directory, if configured
	Uid        int    // If not 0 the program will switch to this user id after initialization
	Gid        int    // If not 0 the program will switch to this group id after initialization
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
		} else if name == "chroot" {
			cf.Chroot = value
		} else if name == "uid" {
			cf.Uid, err = strconv.Atoi(value)
			if err != nil || cf.Uid < 0 || cf.Uid > 65536 {
				return fmt.Errorf("Invalid uid in line %d", lineCount)
			}
		} else if name == "gid" {
			cf.Gid, err = strconv.Atoi(value)
			if err != nil || cf.Gid < 0 || cf.Gid > 65536 {
				return fmt.Errorf("Invalid gid in line %d", lineCount)
			}
		} else {
			return fmt.Errorf("Unknown setting in line %d", lineCount)
		}

	}
	return file.Close()
}

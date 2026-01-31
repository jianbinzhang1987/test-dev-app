package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"deploymaster-pro-wails/internal/ssh"
)

const version = "1.0.0"

type payload struct {
	Version    string   `json:"version"`
	SourcePath string   `json:"sourcePath"`
	RemotePath string   `json:"remotePath"`
	Slaves     []target `json:"slaves"`
}

type target struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	RemotePath string `json:"remotePath"`
}

func main() {
	showVersion := flag.Bool("version", false, "print version")
	payloadBase64 := flag.String("payload", "", "base64 encoded payload")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if strings.TrimSpace(*payloadBase64) == "" {
		fmt.Fprintln(os.Stderr, "missing --payload")
		os.Exit(2)
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(*payloadBase64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode payload failed: %v\n", err)
		os.Exit(2)
	}

	var req payload
	if err := json.Unmarshal(payloadBytes, &req); err != nil {
		fmt.Fprintf(os.Stderr, "parse payload failed: %v\n", err)
		os.Exit(2)
	}

	if req.SourcePath == "" {
		fmt.Fprintln(os.Stderr, "missing sourcePath")
		os.Exit(2)
	}

	if len(req.Slaves) == 0 {
		fmt.Fprintln(os.Stderr, "no slaves provided")
		os.Exit(2)
	}

	for _, slave := range req.Slaves {
		user := slave.User
		if strings.TrimSpace(user) == "" {
			user = "root"
		}
		if strings.TrimSpace(slave.Password) == "" {
			fmt.Fprintf(os.Stderr, "missing password for slave %s\n", slave.Name)
			os.Exit(3)
		}

		targetPath := slave.RemotePath
		if strings.TrimSpace(targetPath) == "" {
			targetPath = req.RemotePath
		}

		client := ssh.NewClient(user, slave.Password)
		if err := client.Connect(slave.Host, slave.Port); err != nil {
			fmt.Fprintf(os.Stderr, "connect slave %s failed: %v\n", slave.Name, err)
			os.Exit(4)
		}

		sftpClient, err := client.NewSFTPClient()
		if err != nil {
			_ = client.Close()
			fmt.Fprintf(os.Stderr, "create sftp for %s failed: %v\n", slave.Name, err)
			os.Exit(4)
		}

		if err := ssh.UploadPath(sftpClient, req.SourcePath, targetPath); err != nil {
			_ = sftpClient.Close()
			_ = client.Close()
			fmt.Fprintf(os.Stderr, "upload to slave %s failed: %v\n", slave.Name, err)
			os.Exit(5)
		}

		_ = sftpClient.Close()
		_ = client.Close()
	}
}

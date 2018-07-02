package main

import (
	"io"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Connect to SSH, exiting process if cannot connect.
func connectSSHOrCrash() *ssh.Client {
	sshLog := log.WithFields(log.Fields{
		"module": "ssh",
	})
	sshLog.Info("Connecting to SSH server...")

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("root"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "192.168.2.34:32768", config)
	if err != nil {
		sshLog.Fatal("Failed to dial: ", err)
	}

	sshLog.Info("Dial succeeded!")
	return client
}

func createTunnelOrCrash(sshClient *ssh.Client, path, target string) {
	listener, err := sshClient.ListenUnix(path)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		myLog := log.WithFields(log.Fields{
			"module": "ssh-tunnel",
		})
		for {
			conn, err := listener.Accept()
			if err != nil {
				myLog.Error("Cannot accept connection", err)
				return
			}
			go func() {
				targetConn, err := net.Dial("tcp", target)
				if err != nil {
					myLog.Error("Cannot connect", err)
					return
				}
				copyConn := func(writer, reader net.Conn) {
					_, err := io.Copy(writer, reader)
					if err != nil {
						myLog.Printf("io.Copy error: %s", err)
					}
				}
				go copyConn(conn, targetConn)
				go copyConn(targetConn, conn)
			}()
		}
	}()
}

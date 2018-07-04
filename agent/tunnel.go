package main

import (
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var myLog = log.WithFields(log.Fields{
	"module": "tunnel",
})

// Connect to SSH, exiting process if cannot connect.
func connectSSHOrCrash(remote, username, password string) *ssh.Client {
	myLog.Info("Connecting to SSH server...")

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(5 * time.Second),
	}

	client, err := ssh.Dial("tcp", remote, config)
	if err != nil {
		myLog.Fatal("Cannot connect to SSH server: ", err)
	}

	myLog.Info("Dial succeeded!")
	return client
}

func createTunnelOrCrash(sshClient *ssh.Client, path, target string) {
	listener, err := sshClient.ListenUnix(path)
	if err != nil {
		myLog.Fatal("Cannot set up Unix domain socket for listening: ", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				myLog.Error("Cannot accept connection: ", err)
				return
			}
			go func() {
				targetConn, err := net.Dial("tcp", target)
				if err != nil {
					myLog.Error("Cannot connect to", target, ": ", err)
					return
				}
				transfer := func(writer, reader net.Conn) {
					_, err := io.Copy(writer, reader)
					if err != nil {
						myLog.Error("Cannot forward traffic: ", err)
					}
				}
				go transfer(conn, targetConn)
				go transfer(targetConn, conn)
			}()
		}
	}()
}

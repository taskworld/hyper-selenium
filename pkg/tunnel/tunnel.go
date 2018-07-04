package tunnel

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

type Tunnel struct {
	client *ssh.Client
}

// ConnectOrCrash connect to the SSH server, exiting process if cannot connect.
func ConnectOrCrash(remote, username, password string) *Tunnel {
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
	return &Tunnel{client}
}

// Close the underlying SSH connection.
func (t *Tunnel) Close() error {
	return t.client.Close()
}

// CreateTunnelOrCrash creates a Unix domain socket at `path`
// in the remote-side, proxying all connection to `target`
// on local side.
func (t *Tunnel) CreateTunnelOrCrash(path, target string) {
	sshClient := t.client
	tunnelLog := myLog.WithFields(log.Fields{
		"path":   path,
		"target": target,
	})
	listener, err := sshClient.ListenUnix(path)
	if err != nil {
		tunnelLog.Fatal("Cannot set up Unix domain socket for listening: ", err)
	}
	tunnelLog.Info("Unix domain socket created")

	for {
		conn, err := listener.Accept()
		if err != nil {
			tunnelLog.Fatal("Cannot accept connection: ", err)
		}
		tunnelLog.Info("Accepting connection!")
		go func() {
			targetConn, err := net.Dial("tcp", target)
			if err != nil {
				tunnelLog.Error("Cannot connect to", target, ": ", err)
				return
			}
			transfer := func(writer, reader net.Conn) {
				_, err := io.Copy(writer, reader)
				if err != nil {
					tunnelLog.Error("Cannot forward traffic: ", err)
				}
			}
			go transfer(conn, targetConn)
			go transfer(targetConn, conn)
		}()
	}
}

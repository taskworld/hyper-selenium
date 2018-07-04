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

// CreateRemoteTunnelOrCrash creates a Unix domain socket at `path`
// in the remote-side, proxying all connection to `target`
// on local side.
func (t *Tunnel) CreateRemoteTunnelOrCrash(path, target string) {
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

// CreateLocalTunnelOrCrash creates a tunnel on local machine
// that dials to the target path. Returns the local listening address.
func (t *Tunnel) CreateLocalTunnelOrCrash(path string) string {
	sshClient := t.client
	tunnelLog := myLog.WithFields(log.Fields{
		"path": path,
	})
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		tunnelLog.Fatal("Cannot set up local tunnel: ", err)
	}
	address := listener.Addr().String()
	tunnelLog = tunnelLog.WithField("address", address)
	tunnelLog.Info("Local tunnel created")

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				tunnelLog.Fatal("Cannot accept connection: ", err)
			}
			go func() {
				targetConn, err := sshClient.Dial("unix", path)
				if err != nil {
					tunnelLog.Error("Cannot connect to", path, ": ", err)
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
	}()

	return address
}

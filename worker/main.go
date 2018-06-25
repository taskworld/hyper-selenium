package main

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func main() {
	cmd := startSeleniumServerOrCrash()
	sshClient := connectSSHOrCrash()
	defer sshClient.Close()
	waitForSeleniumServerToBecomeAvailableOrCrash()
	prefix := "/tmp/meow"
	createTunnelOrCrash(sshClient, prefix+".selenium", "localhost:4444")
	createTunnelOrCrash(sshClient, prefix+".vnc", "localhost:5900")

	err := cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

// Start selenium process, exit if cannot start.
func startSeleniumServerOrCrash() *exec.Cmd {
	seleniumLog := log.WithFields(log.Fields{
		"module": "selenium",
	})

	seleniumLog.Info("Starting Selenium server...")
	cmd := exec.Command("/opt/bin/entry_point.sh")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			seleniumLog.Error(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			seleniumLog.Fatal(err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			seleniumLog.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			seleniumLog.Fatal(err)
		}
	}()

	err = cmd.Start()
	if err != nil {
		seleniumLog.Fatal("Cannot start Selenium process:", err)
	}

	return cmd
}

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

// Check for selenium server to respond positively.
func waitForSeleniumServerToBecomeAvailableOrCrash() {
	myLog := log.WithFields(log.Fields{
		"module": "checkSeleniumServer",
	})
	start := time.Now()
	maxTimeout := 30.0
	for {
		if time.Now().Sub(start).Seconds() > maxTimeout {
			myLog.Fatal("Give up checking Selenium server")
			return
		}
		timeout := time.Duration(5 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get("http://localhost:4444/wd/hub/status/")
		if err != nil {
			myLog.Error("Cannot connect to Selenium", err)
			time.Sleep(time.Duration(500) * time.Millisecond)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			myLog.WithFields(log.Fields{
				"statusCode": resp.StatusCode,
			}).Error("Status code is not 200")
			time.Sleep(time.Duration(500) * time.Millisecond)
			continue
		}
		myLog.Info("Selenium server is ready.")
		return
	}
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

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/taskworld/hyper-selenium/pkg/tunnel"
)

var sessionID string
var sshRemote string
var sshUsername string
var sshPassword string
var utility []string

func init() {
	flag.StringVar(&sessionID, "id", "", "session id -- must be unique")
	flag.StringVar(&sshRemote, "ssh-remote", "localhost:22", "ssh server address")
	flag.StringVar(&sshUsername, "ssh-username", "root", "ssh server username")
	flag.StringVar(&sshPassword, "ssh-password", "root", "ssh server password")
	flag.Parse()

	if sessionID == "" {
		fmt.Println("id is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	utility = flag.Args()
	if len(utility) < 1 {
		fmt.Println("utility is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	myLog := log.WithFields(log.Fields{
		"sessionID": sessionID,
	})
	myLog.Info("Initializing...")
	prefix := "/tmp/hyper-selenium-" + sessionID

	tunnel := tunnel.ConnectOrCrash(sshRemote, sshUsername, sshPassword)
	childEnv := []string{
		"HS_SELENIUM_ADDRESS=" + tunnel.CreateLocalTunnelOrCrash(prefix+"-selenium"),
		"HS_VNC_ADDRESS=" + tunnel.CreateLocalTunnelOrCrash(prefix+"-vnc"),
		"HS_INFO_ADDRESS=" + tunnel.CreateLocalTunnelOrCrash(prefix+"-info"),
	}
	defer tunnel.Close()

	cmd := exec.Command(utility[0], utility[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), childEnv...)
	if err := cmd.Run(); err != nil {
		myLog.Fatal("Utility run failed: ", err)
	}
}

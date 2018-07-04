package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/taskworld/hyper-selenium/pkg/tunnel"
)

var sessionID string
var sshRemote string
var sshUsername string
var sshPassword string

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
}

func main() {
	myLog := log.WithFields(log.Fields{
		"sessionID": sessionID,
	})
	myLog.Info("Initializing...")
	tunnel := tunnel.ConnectOrCrash(sshRemote, sshUsername, sshPassword)
	defer tunnel.Close()
}

package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/taskworld/hyper-selenium/agent/infoserver"
)

var sessionId string
var sshRemote string
var sshUsername string
var sshPassword string

func init() {
	flag.StringVar(&sessionId, "id", "", "session id -- must be unique")
	flag.StringVar(&sshRemote, "ssh-remote", "localhost:22", "ssh server address")
	flag.StringVar(&sshUsername, "ssh-username", "root", "ssh server username")
	flag.StringVar(&sshPassword, "ssh-password", "root", "ssh server password")
	flag.Parse()

	if sessionId == "" {
		fmt.Println("id is required")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	infoserver.StartInfoServer()

	cmd := startSeleniumServerOrCrash()
	sshClient := connectSSHOrCrash(sshRemote, sshUsername, sshPassword)
	defer sshClient.Close()
	waitForSeleniumServerToBecomeAvailableOrCrash()
	prefix := "/tmp/meow02"
	createTunnelOrCrash(sshClient, prefix+".selenium", "localhost:4444")
	createTunnelOrCrash(sshClient, prefix+".vnc", "localhost:5900")
	createTunnelOrCrash(sshClient, prefix+".info", "localhost:8080")
	startRecordingVideo()

	err := cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

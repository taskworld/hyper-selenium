package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/taskworld/hyper-selenium/agent/infoserver"
	"github.com/taskworld/hyper-selenium/agent/selenium"
	"github.com/taskworld/hyper-selenium/agent/tunnel"
	"github.com/taskworld/hyper-selenium/agent/vtr"
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

	selenium := selenium.StartOrCrash()
	defer selenium.Wait()

	tunnel := tunnel.ConnectOrCrash(sshRemote, sshUsername, sshPassword)
	defer tunnel.Close()

	selenium.WaitForServerToBecomeAvailableOrCrash()

	prefix := "/tmp/hyper-selenium-" + sessionId
	go tunnel.CreateTunnelOrCrash(prefix+"-selenium", "localhost:4444")
	go tunnel.CreateTunnelOrCrash(prefix+"-vnc", "localhost:5900")
	go tunnel.CreateTunnelOrCrash(prefix+"-info", "localhost:8080")

	go func() {
		selenium.WaitForSession()
		vtr.StartRecordingVideo()
	}()

	selenium.Wait()
}

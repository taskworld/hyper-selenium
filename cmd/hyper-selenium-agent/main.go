package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/taskworld/hyper-selenium/pkg/infoserver"
	"github.com/taskworld/hyper-selenium/pkg/selenium"
	"github.com/taskworld/hyper-selenium/pkg/tunnel"
	"github.com/taskworld/hyper-selenium/pkg/vtr"
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
	infoserver := infoserver.StartInfoServer()
	infoserver.SetStatus("initializing")

	infoserver.SetMessage("Connecting to tunnel server...")
	tunnel := tunnel.ConnectOrCrash(sshRemote, sshUsername, sshPassword)
	defer tunnel.Close()

	infoserver.SetMessage("Setting up tunnels...")
	pathPrefix := "/tmp/hyper-selenium-" + sessionID
	go tunnel.CreateRemoteTunnelOrCrash(pathPrefix+"-info", "localhost:8080")
	go tunnel.CreateRemoteTunnelOrCrash(pathPrefix+"-selenium", "localhost:4444")
	go tunnel.CreateRemoteTunnelOrCrash(pathPrefix+"-vnc", "localhost:5900")

	infoserver.SetMessage("Starting Selenium server...")
	selenium := selenium.StartOrCrash()
	defer selenium.Wait()

	infoserver.SetMessage("Waiting for Selenium server to become available...")
	selenium.WaitForServerToBecomeAvailableOrCrash()

	go func() {
		selenium.WaitForSession()
		vtr := vtr.StartRecordingVideo(nil)
		infoserver.VTR = vtr
	}()

	infoserver.SetStatus("ready")
	infoserver.SetMessage("Server is up!")
	selenium.Wait()
}

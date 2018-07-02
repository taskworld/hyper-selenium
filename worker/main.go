package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/taskworld/hyper-selenium/worker/infoserver"
)

func main() {
	infoserver.StartInfoServer()

	cmd := startSeleniumServerOrCrash()
	sshClient := connectSSHOrCrash()
	defer sshClient.Close()
	waitForSeleniumServerToBecomeAvailableOrCrash()
	prefix := "/tmp/meow"
	createTunnelOrCrash(sshClient, prefix+".selenium", "localhost:4444")
	createTunnelOrCrash(sshClient, prefix+".vnc", "localhost:5900")
	createTunnelOrCrash(sshClient, prefix+".info", "localhost:8080")

	err := cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

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
	infoServerAddress := tunnel.CreateLocalTunnelOrCrash(prefix + "-info")
	childEnv := []string{
		"HS_SELENIUM_ADDRESS=" + tunnel.CreateLocalTunnelOrCrash(prefix+"-selenium"),
		"HS_VNC_ADDRESS=" + tunnel.CreateLocalTunnelOrCrash(prefix+"-vnc"),
		"HS_INFO_ADDRESS=" + infoServerAddress,
	}
	defer tunnel.Close()

	start := time.Now()
	maxTimeout := 30.0
	lastStatusText := ""
	for {
		if time.Now().Sub(start).Seconds() > maxTimeout {
			myLog.Fatal("Give up waiting for server to be ready...")
			return
		}
		timeout := time.Duration(5 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get("http://" + infoServerAddress + "/status")
		if err != nil {
			myLog.Error("Cannot connect to server: ", err)
			time.Sleep(time.Duration(500) * time.Millisecond)
			continue
		}
		if resp.StatusCode != 200 {
			myLog.WithFields(log.Fields{
				"statusCode": resp.StatusCode,
			}).Error("Status code is not 200")
			time.Sleep(time.Duration(500) * time.Millisecond)
			continue
		}
		var data struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		// respBody, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	myLog.Error("Cannot read JSON: ", err)
		// 	continue
		// }
		// myLog.Info(string(respBody))
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			myLog.Error("Cannot parse JSON: ", err)
			continue
		}
		statusText := data.Status + ": " + data.Message
		if statusText != lastStatusText {
			lastStatusText = statusText
			myLog.Info(statusText)
		}
		if data.Status != "ready" {
			continue
		}
		myLog.Info("Server is ready now!")
		break
	}

	cmd := exec.Command(utility[0], utility[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), childEnv...)
	if err := cmd.Run(); err != nil {
		myLog.Fatal("Utility run failed: ", err)
	}
}

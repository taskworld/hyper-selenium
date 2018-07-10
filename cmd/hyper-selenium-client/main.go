package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/taskworld/hyper-selenium/pkg/tunnel"
)

var sessionID string
var sshRemote string
var sshUsername string
var sshPassword string
var videoOutPath string
var utility []string

var myLog *log.Entry

func init() {
	flag.StringVar(&sessionID, "id", "", "session id -- must be unique")
	flag.StringVar(&sshRemote, "ssh-remote", "localhost:22", "ssh server address")
	flag.StringVar(&sshUsername, "ssh-username", "root", "ssh server username")
	flag.StringVar(&sshPassword, "ssh-password", "root", "ssh server password")
	flag.StringVar(&videoOutPath, "video-out", "", "download video file")
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

	myLog = log.WithFields(log.Fields{
		"sessionID": sessionID,
	})
}

func main() {
	os.Exit(run())
}

func run() int {
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
			return 1
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
	defer fetchVideo(infoServerAddress)

	cmd := exec.Command(utility[0], utility[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), childEnv...)
	if err := cmd.Run(); err != nil {
		myLog.Error("Utility run failed: ", err)
		// Copy-and-pasted from https://stackoverflow.com/questions/10385551/get-exit-code-go
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		return 1
	}

	return 0
}

func fetchVideo(infoServerAddress string) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	myLog.Info("Finishing the video recording...")
	resp, err := client.Get("http://" + infoServerAddress + "/vtr/finish")
	if err != nil {
		myLog.Error("Cannot finish recording video: Cannot connect to server: ", err)
		return
	}
	if resp.StatusCode != 200 {
		myLog.WithFields(log.Fields{
			"statusCode": resp.StatusCode,
		}).Error("Cannot finish recording video: Status code is not 200")
		return
	}

	if videoOutPath == "" {
		myLog.Info("Finished recording video but not downloading because --video-out flag is not set.")
		return
	}

	myLog.Info("Downloading the recorded video...")
	resp, err = client.Get("http://" + infoServerAddress + "/videos/video.mp4")
	if err != nil {
		myLog.Error("Cannot download video: Cannot connect to server: ", err)
		return
	}
	if resp.StatusCode != 200 {
		myLog.WithFields(log.Fields{
			"statusCode": resp.StatusCode,
		}).Error("Cannot download video: Status code is not 200")
		return
	}
	out, err := os.Create(videoOutPath)
	if err != nil {
		myLog.Error("Cannot download video: Cannot create file: ", err)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		myLog.Error("Cannot download video: Cannot download: ", err)
		return
	}
}

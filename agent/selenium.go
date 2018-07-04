package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var myLog = log.WithFields(log.Fields{
	"module": "selenium",
})

// Start selenium process, exit if cannot start.
func startSeleniumServerOrCrash() *exec.Cmd {
	myLog.Info("Starting Selenium server...")
	cmd := exec.Command("/opt/bin/entry_point.sh")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		myLog.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		myLog.Fatal(err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			myLog.Error(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			myLog.Fatal(err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			myLog.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			myLog.Fatal(err)
		}
	}()

	err = cmd.Start()
	if err != nil {
		myLog.Fatal("Cannot start Selenium process:", err)
	}

	return cmd
}

// Check for selenium server to respond positively.
func waitForSeleniumServerToBecomeAvailableOrCrash() {
	start := time.Now()
	maxTimeout := 30.0
	myLog.Info("Waiting for Selenium server to become available...")
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
			myLog.Error("Cannot connect to Selenium: ", err)
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
		myLog.Info("Selenium server is ready. ^_^")
		return
	}
}

// Wait for a session to be created.
func waitForSeleniumSession() error {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	lastSessionCount := -1
	for {
		resp, err := client.Get("http://localhost:4444/wd/hub/sessions")
		if err != nil {
			return errors.Wrap(err, "Cannot reach Selenium server")
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("Status code (%d) is not 200", resp.StatusCode)
		}
		var data struct {
			value []interface{}
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return errors.Wrap(err, "Cannot parse resulting JSON")
		}
		sessionsCount := len(data.value)
		if lastSessionCount != sessionsCount {
			lastSessionCount = sessionsCount
			myLog.Infof("There are %d active session(s)", sessionsCount)
		}
		if sessionsCount > 0 {
			return
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
}

package main

import (
	"bufio"
	"net/http"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

// Start selenium process, exit if cannot start.
func startSeleniumServerOrCrash() *exec.Cmd {
	seleniumLog := log.WithFields(log.Fields{
		"module": "selenium",
	})

	seleniumLog.Info("Starting Selenium server...")
	cmd := exec.Command("/opt/bin/entry_point.sh")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		seleniumLog.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		seleniumLog.Fatal(err)
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

// Check for selenium server to respond positively.
func waitForSeleniumServerToBecomeAvailableOrCrash() {
	myLog := log.WithFields(log.Fields{
		"module": "checkSeleniumServer",
	})
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

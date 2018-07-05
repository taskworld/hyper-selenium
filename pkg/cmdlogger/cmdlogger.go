package cmdlogger

import (
	"bufio"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func LogCommandOutput(myLog *log.Entry, cmd *exec.Cmd) {
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
			myLog.Error("Reading from STDERR failed: ", err)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			myLog.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			myLog.Error("Reading from STDOUT failed: ", err)
		}
	}()
}

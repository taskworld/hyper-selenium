package main

import (
	"bufio"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// VTR represents a running instance of a video recorder process.
type VTR struct {
	cmd *exec.Cmd
}

var myLog := log.WithFields(log.Fields{
	"module": "vtr",
})

func startRecordingVideo() *VTR {
	myLog.Info("Recording video...")
	cmd := exec.Command(
		"ffmpeg",
		"-video_size", "1280x1024",
		"-framerate", "15",
		"-f", "x11grab",
		"-i", ":99.0",
		"/videos/video.mp4",
	)

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
		myLog.Fatal("Cannot start video recording process:", err)
	}

	myLog.Info("Video recording started...")

	return &VTR{
		cmd: cmd,
	}
}

// StopRecordingVideo tells the video recording process to stop recording
// video and blocks until video is converted successfully.
func (v *VTR) StopRecordingVideo() error {
	if err := v.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return err
	}
	if err := v.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

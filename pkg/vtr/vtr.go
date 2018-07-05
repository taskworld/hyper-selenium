package vtr

import (
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/taskworld/hyper-selenium/pkg/cmdlogger"
)

// VTR represents a running instance of a video recorder process.
type VTR struct {
	cmd      *exec.Cmd
	status   string
	listener ProgressListener
}

type ProgressListener interface {
	SetStatus(text string)
}

var myLog = log.WithFields(log.Fields{
	"module": "vtr",
})

// StartRecordingVideo starts the video capture process.
func StartRecordingVideo(listener ProgressListener) *VTR {
	myLog.Info("Recording video...")
	if listener != nil {
		listener.SetStatus("launching")
	}
	cmd := exec.Command(
		"ffmpeg",
		"-video_size", "1280x1024",
		"-framerate", "15",
		"-f", "x11grab",
		"-i", ":99.0",
		"/videos/video.mp4",
	)
	cmdlogger.LogCommandOutput(myLog.WithField("cmd", "ffmpeg"), cmd)

	err := cmd.Start()
	if err != nil {
		myLog.Fatal("Cannot start video recording process:", err)
	}

	myLog.Info("Video recording started...")
	if listener != nil {
		listener.SetStatus("recording")
	}
	return &VTR{
		cmd:      cmd,
		listener: listener,
	}
}

// StopRecordingVideo tells the video recording process to stop recording
// video and blocks until video is converted successfully.
func (v *VTR) StopRecordingVideo() error {
	myLog.Info("Stopping video recording.")
	listener := v.listener
	if listener != nil {
		listener.SetStatus("stopping")
	}
	if err := v.cmd.Process.Signal(syscall.SIGINT); err != nil {
		if listener != nil {
			listener.SetStatus("error")
		}
		myLog.Error("Cannot send SIGINT to ffmpeg process: ", err)
		return err
	}
	if err := v.cmd.Wait(); err != nil {
		myLog.Info("ffmpeg exited: ", err)
	}
	myLog.Info("Finalizing the video.")
	if listener != nil {
		listener.SetStatus("finalizing")
	}
	cmd := exec.Command(
		"MP4Box",
		"-isma",
		"-inter", "500",
		"/videos/video.mp4",
	)
	cmdlogger.LogCommandOutput(myLog.WithField("cmd", "MP4Box"), cmd)
	if err := cmd.Run(); err != nil {
		myLog.Error("Finalization error: ", err)
		listener.SetStatus("error")
		return err
	}
	if listener != nil {
		listener.SetStatus("finished")
	}
	return nil
}

func (v *VTR) setStatus(text string) {
	v.status = text
}

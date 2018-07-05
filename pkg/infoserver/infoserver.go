package infoserver

import (
	"encoding/json"
	"net/http"

	"github.com/taskworld/hyper-selenium/pkg/vtr"

	log "github.com/sirupsen/logrus"
)

type InfoServer struct {
	status  string
	message string

	VTR *vtr.VTR
}

func StartInfoServer() *InfoServer {
	infoserver := &InfoServer{}
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  infoserver.status,
			"message": infoserver.message,
		})
	})
	http.HandleFunc("/vtr/finish", func(w http.ResponseWriter, r *http.Request) {
		vtr := infoserver.VTR
		infoserver.VTR = nil
		if vtr != nil {
			vtr.StopRecordingVideo()
		}
	})
	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir("/videos"))))
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()
	return infoserver
}

func (s *InfoServer) SetStatus(text string) {
	s.status = text
}

func (s *InfoServer) SetMessage(text string) {
	s.message = text
}

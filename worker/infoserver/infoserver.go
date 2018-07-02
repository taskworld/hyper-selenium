package infoserver

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func StartInfoServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		message := r.URL.Path
		message = strings.TrimPrefix(message, "/")
		message = "Hello " + message
		w.Write([]byte(message))
	})
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()
}

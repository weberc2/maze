package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gorilla/mux"
)

func handler(hf HandlerFunc) http.HandlerFunc {
	return HTTPHandlerFunc(os.Stderr, hf)
}

func fileHandler(path string) http.HandlerFunc {
	return handler(func(
		w http.ResponseWriter,
		r *http.Request,
		logger *Logger,
	) {
		file, err := os.Open(path)
		if err != nil {
			logger.Logf("Error opening file '%s': %v", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()
		if _, err := io.Copy(w, file); err != nil {
			logger.Logf("Error writing to HTTP response writer: %v", err)
			return
		}
	})
}

func main() {
	r := mux.NewRouter()
	server := Server{}
	r.Path("/stats-socket/").HandlerFunc(handler(server.Stats))
	r.Path("/user-socket/").HandlerFunc(handler(server.User))
	r.Path("/stats/").HandlerFunc(fileHandler("./stats.html"))
	r.Path("/").HandlerFunc(fileHandler("./index.html"))

	log.Println("Listening!")
	if false { // TODO: Re-enable this once TLS is working
		if err := http.Serve(
			autocert.NewListener("localhost", "weberc2.com", "www.weberc2.com"),
			r,
		); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(":8080", r); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

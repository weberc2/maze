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

func fileHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open(path)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()
		if _, err := io.Copy(w, file); err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	r := mux.NewRouter()
	server := Server{}
	r.Path("/stats-socket/").HandlerFunc(server.Stats)
	r.Path("/user-socket/").HandlerFunc(server.User)
	r.Path("/stats/").HandlerFunc(fileHandler("./stats.html"))
	r.Path("/").HandlerFunc(fileHandler("./index.html"))

	log.Println("Listening!")
	if err := http.Serve(autocert.NewListener("weberc2.com"), r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

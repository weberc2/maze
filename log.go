package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/pborman/uuid"
)

type Logger struct {
	data []interface{}
}

func (l *Logger) Log(v interface{}) {
	l.data = append(l.data, v)
}

func (l *Logger) Logf(format string, v ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, v...))
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request, l *Logger)

type responseWriterWrapper struct {
	w          http.ResponseWriter
	statusCode int
}

func (rww *responseWriterWrapper) Header() http.Header {
	return rww.w.Header()
}

func (rww *responseWriterWrapper) Write(p []byte) (int, error) {
	return rww.w.Write(p)
}

func (rww *responseWriterWrapper) WriteHeader(s int) {
	rww.statusCode = s
	rww.w.WriteHeader(s)
}

func HTTPHandlerFunc(out io.Writer, hf HandlerFunc) http.HandlerFunc {
	// Spin up a separate goroutine for serialization so as to not stall
	// the requests. The input channel holds 1024 objects before blocking, so
	// this should give us some runway.
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "    ")
	serializer := make(chan interface{}, 1024)
	go func() {
		for v := range serializer {
			if err := encoder.Encode(v); err != nil {
				log.Println("Failed to write to logfile:", err)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		logger := Logger{}
		rww := responseWriterWrapper{w: w, statusCode: http.StatusOK}

		start := time.Now()
		hf(&rww, r, &logger)
		logger.Log(map[string]map[string]string{
			"request": map[string]string{
				"id":          uuid.New(),
				"path":        r.URL.Path,
				"referer":     r.Referer(),
				"remote_addr": r.RemoteAddr,
				"http_proto":  r.Proto,
				"user_agent":  r.UserAgent(),
				"timestamp":   start.Format(time.RFC3339Nano),
				"status_code": strconv.Itoa(rww.statusCode),
				"duration":    time.Since(start).String(),
			},
		})
		serializer <- logger.data
	}
}

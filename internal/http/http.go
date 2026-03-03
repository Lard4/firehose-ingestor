package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	mu sync.Mutex
	ch chan []byte
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/feed", s.HandleFeed)

	srv := &http.Server{
		Addr:    ":5005",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	fmt.Println("HTTP server started on :5005")
}

func (s *Server) HandleFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// register this client
	ch := make(chan []byte, 100)
	s.mu.Lock()
	s.ch = ch
	s.mu.Unlock()

	// wait for disconnect or new messages
	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			s.mu.Lock()
			s.ch = nil
			s.mu.Unlock()
			return
		}
	}
}

func (s *Server) Send(data []byte) {
	fmt.Println("Sending data:", string(data))
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ch != nil {
		s.ch <- data
	}
}

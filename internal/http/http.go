package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	mu      sync.Mutex
	// essentially just a Set of channels (one for each client that connects)
	clients map[chan []byte]struct{}
}

func NewServer() *Server {
	return &Server{
		clients: make(map[chan []byte]struct{}),
	}
}

func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/feed", s.HandleFeed)

	srv := &http.Server{
		Addr:    ":3005",
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

	fmt.Println("HTTP server started on :3005")
}

func (s *Server) HandleFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "ngrok-skip-browser-warning")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// register this client
	ch := make(chan []byte, 100)
	s.mu.Lock()
	// struct{}{} is a weird go idiom that takes up zero bytes 
	s.clients[ch] = struct{}{}
	fmt.Printf("Adding new channel to map %#v\n", s.clients)
	s.mu.Unlock()

	defer func() {
        s.mu.Lock()
        delete(s.clients, ch)
        s.mu.Unlock()
    }()

	// wait for disconnect or new messages
	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) Send(data []byte) {
    s.mu.Lock()
    defer s.mu.Unlock()
    for ch := range s.clients {
        select {
        case ch <- data:
        default:
            // channel full, drop for this client
        }
    }
}

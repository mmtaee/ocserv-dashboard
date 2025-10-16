package web

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	clients map[chan string]string // IP
	mu      sync.Mutex
}

func NewSSEServer() *Server {
	return &Server{
		clients: make(map[chan string]string),
	}
}

func (s *Server) AddClient(client chan string, ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client] = ip
	log.Printf("Client %v (%s) connected", client, ip)
}

// RemoveClient removes a client connection
func (s *Server) RemoveClient(client chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ip, ok := s.clients[client]; ok {
		log.Printf("Client %v (%s) disconnected", client, ip)
		delete(s.clients, client)
		close(client)
	}
}

func (s *Server) StartBroadcast(broadcaster <-chan string) {
	go func() {
		for msg := range broadcaster {
			s.mu.Lock()
			for ch := range s.clients {
				select {
				case ch <- msg:
				default:
					log.Println("Dropped message for client", s.clients[ch])
					continue
				}
			}
			s.mu.Unlock()
		}
	}()
}

func (s *Server) SSEHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(s.clients) > 3 {
			http.Error(w, "Too many clients connected", http.StatusTooManyRequests)
			return
		}

		// Setup SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		clientChan := make(chan string, 10)
		s.AddClient(clientChan, r.RemoteAddr)

		for {
			select {
			case <-r.Context().Done():
				s.RemoveClient(clientChan)
				return
			case message := <-clientChan:
				if message == "" {
					return
				}
				_, err := fmt.Fprintf(w, "data: %s\n\n", message)
				if err != nil {
					log.Println("Error writing to client:", err)
					return
				}
				flusher.Flush()
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

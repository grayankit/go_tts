package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	clients      = make(map[chan string]bool)
	clientsMu    sync.Mutex
	messageQueue []string
	queueMu      sync.Mutex
	isPaused     bool
)

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SSE client connected")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming Unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	msgChan := make(chan string)
	clientsMu.Lock()
	clients[msgChan] = true
	clientsMu.Unlock()

	queueMu.Lock()
	if !isPaused {
		for _, msg := range messageQueue {
			msgChan <- msg
		}
		messageQueue = nil
	}
	queueMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, msgChan)
		clientsMu.Unlock()
		close(msgChan)
		fmt.Println("SSE client disconnected")
	}()

	for {
		select {
		case msg := <-msgChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func broadcastToClients(text string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for ch := range clients {
		select {
		case ch <- text:
		default:
			log.Println("Dropped slow client")
		}
	}
}

func Broadcast(text string) {
	queueMu.Lock()
	defer queueMu.Unlock()

	clientsMu.Lock()
	numClients := len(clients)
	clientsMu.Unlock()

	if numClients == 0 || isPaused {
		log.Printf("Queueing message: %s (isPaused: %v, numClients: %d)", text, isPaused, numClients)
		messageQueue = append(messageQueue, text)
		return
	}

	log.Printf("Broadcasting message immediately: %s", text)
	broadcastToClients(text)
}

func SpeakHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Text  string `json:"text"`
		Voice string `json:"voice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if data.Text == "" {
		http.Error(w, "Missing text", http.StatusBadRequest)
		return
	}
	Broadcast(data.Text)
	w.WriteHeader(http.StatusNoContent)
}

func GetQueueHandler(w http.ResponseWriter, r *http.Request) {
	queueMu.Lock()
	defer queueMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Queue has %d items", len(messageQueue))
	if messageQueue == nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}
	json.NewEncoder(w).Encode(messageQueue)
}

type PauseRequest struct {
	Paused bool `json:"paused"`
}

func PauseHandler(w http.ResponseWriter, r *http.Request) {
	var req PauseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	queueMu.Lock()
	isPaused = req.Paused
	queueMu.Unlock()

	log.Printf("Pause state changed to: %v", isPaused)

	if !isPaused {
		go func() {
			for {
				queueMu.Lock()
				if isPaused || len(messageQueue) == 0 {
					queueMu.Unlock()
					break
				}
				msg := messageQueue[0]
				messageQueue = messageQueue[1:]
				queueMu.Unlock()

				log.Printf("Broadcasting from queue: %s", msg)
				broadcastToClients(msg)

				// Adjust sleep to match your audio length roughly
				time.Sleep(3 * time.Second)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Paused bool `json:"paused"`
	}{Paused: isPaused})
}


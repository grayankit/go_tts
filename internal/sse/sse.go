package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	clients      = make(map[chan string]bool)
	clientsMu    sync.Mutex
	messageQueue []string
	queueMu      sync.Mutex
	frontedReady bool
	isPaused     bool
)

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request called")
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
	frontedReady = true

	for _, msg := range messageQueue {
		msgChan <- msg
	}
	messageQueue = nil
	queueMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, msgChan)
		clientsMu.Unlock()
		queueMu.Lock()
		frontedReady = false
		queueMu.Unlock()
		close(msgChan)
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
func Broadcast(text string) {
	queueMu.Lock()
	defer queueMu.Unlock()
	if !frontedReady || isPaused {
		messageQueue = append(messageQueue, text)
		return
	}
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for ch := range clients {
		select {
		case ch <- text:
		default:
			log.Println("Dropped slow clients")
		}
	}
}
func SpeakHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Text  string `json:"text"`
		Voice string `json:"voice"`
	}
	json.NewDecoder(r.Body).Decode(&data)
	if data.Text == "" {
		http.Error(w, "Missing text", http.StatusBadRequest)
		return
	}
	Broadcast(data.Text)
	w.WriteHeader(204)
}

func GetQueueHandler(w http.ResponseWriter, r *http.Request) {
	queueMu.Lock()
	defer queueMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messageQueue)
}

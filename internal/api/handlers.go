package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"sync"

	"github.com/grayankit/go_tts/internal/sse"
	"github.com/grayankit/go_tts/internal/tts"
)

var (
	mu      sync.Mutex
	history []string
)

func TtsHandler(w http.ResponseWriter, r *http.Request) {
	var req tts.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}
	if req.Voice == "" {
		req.Voice = "en-us"
	}

	audio, err := tts.SynthesizeSpeech(req.Text, req.Voice)
	if err != nil {
		http.Error(w, "TTS Error"+err.Error(), http.StatusInternalServerError)
		return
	}
	mu.Lock()
	if len(history) >= 10 {
		history = history[1:]
	}
	history = append(history, req.Text)
	mu.Unlock()

	w.Header().Set("Content-Type", "audio/wav")
	w.Write(audio)
}
func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(history)
}
func VoicesHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("espeak-ng", "--voices")
	out, err := cmd.Output()
	if err != nil {
		http.Error(w, "Could not list voices", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, bytes.NewReader(out))
}
func PreviewHandler(w http.ResponseWriter, r *http.Request) {
	voice := r.URL.Query().Get("voice")
	if voice == "" {
		voice = "en-us"
	}
	text := "Hi! I am going to sound like this"
	audio, err := tts.SynthesizeSpeech(text, voice)
	if err != nil {
		http.Error(w, "TTS error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "audio/wav")
	w.Write(audio)
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/tts", TtsHandler)
	mux.HandleFunc("/api/history", HistoryHandler)
	mux.HandleFunc("/api/voices", VoicesHandler)
	mux.HandleFunc("api/preview", PreviewHandler)
	mux.HandleFunc("/events", sse.EventsHandler)
	mux.HandleFunc("/api/speak", sse.SpeakHandler)
	mux.HandleFunc("/api/queue", sse.GetQueueHandler)
}

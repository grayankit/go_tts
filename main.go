package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type TTSRequest struct {
	Text  string `json:"text"`
	Voice string `json:"voice"`
}

var (
	mu           sync.Mutex
	history      []string
	clients      = make(map[chan string]bool)
	clientsMu    sync.Mutex
	messageQueue []string
	queueMu      sync.Mutex
	frontedReady bool
	isPaused     bool
)

func synthesizeSpeech(text, voice string) ([]byte, error) {
	if after, ok := strings.CutPrefix(voice, "eleven:"); ok {
		voiceID := after
		return elevenLabsTTS(text, voiceID)
	}
	cmd := exec.Command("espeak-ng", "-v", voice, "--stdout", text)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.Bytes(), err
}
func ttsHandler(w http.ResponseWriter, r *http.Request) {
	var req TTSRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}
	if req.Voice == "" {
		req.Voice = "en-us"
	}

	audio, err := synthesizeSpeech(req.Text, req.Voice)
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
func historyHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(history)
}
func voicesHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("espeak-ng", "--voices")
	out, err := cmd.Output()
	if err != nil {
		http.Error(w, "Could not list voices", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, bytes.NewReader(out))
}
func previewHandler(w http.ResponseWriter, r *http.Request) {
	voice := r.URL.Query().Get("voice")
	if voice == "" {
		voice = "en-us"
	}
	text := "Hi! I am going to sound like this"
	audio, err := synthesizeSpeech(text, voice)
	if err != nil {
		http.Error(w, "TTS error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "audio/wav")
	w.Write(audio)
}
func elevenLabsTTS(text, voiceID string) ([]byte, error) {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("Error Loading .env file")
	}

	apiKey := os.Getenv("ELEVEN_API_KEY")

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	payload := map[string]any{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]any{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
func eventsHandler(w http.ResponseWriter, r *http.Request) {
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
func broadcast(text string) {
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
func speakHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Text  string `json:"text"`
		Voice string `json:"voice"`
	}
	json.NewDecoder(r.Body).Decode(&data)
	if data.Text == "" {
		http.Error(w, "Missing text", http.StatusBadRequest)
		return
	}
	broadcast(data.Text)
	w.WriteHeader(204)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/api/tts", ttsHandler)
	http.HandleFunc("/api/history", historyHandler)
	http.HandleFunc("/api/voices", voicesHandler)
	http.HandleFunc("/api/preview", previewHandler)
	http.HandleFunc("/events", eventsHandler)
	http.HandleFunc("/api/speak", speakHandler)

	fmt.Println("Listening on port 3001 for tts")
	http.ListenAndServe(":3001", nil)
}

package tts

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

	"github.com/joho/godotenv"
)

type Request struct {
	Text  string `json:"text"`
	Voice string `json:"voice"`
}

func SynthesizeSpeech(text, voice string) ([]byte, error) {
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

func elevenLabsTTS(text, voiceID string) ([]byte, error) {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Println("Error loading .env file, will use environment variables")
	}

	apiKey := os.Getenv("ELEVEN_API_KEY")

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	payload := map[string]any{
		"text":     text,
		"model_id": "eleven_multilingual_v2",
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

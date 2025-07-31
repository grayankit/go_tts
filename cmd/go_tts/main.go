package main

import (
	"fmt"
	"net/http"

	"github.com/grayankit/go_tts/internal/api"
)

func main() {
	mux := http.NewServeMux()

	api.RegisterRoutes(mux)

	mux.Handle("/", http.FileServer(http.Dir("static")))

	fmt.Println("Listening on port 3001 for tts")
	http.ListenAndServe(":3001", mux)
}

# Go TTS

Go TTS is a simple web application that converts text to speech. It is built with Go for the backend and plain HTML, CSS, and JavaScript for the frontend.

## About

This application provides a simple interface to enter text and get an audio output of the spoken text. It uses a simple SSE to stream the audio data to the client.

## Features

- Text-to-speech conversion
- Simple and intuitive UI
- Real-time audio streaming

## Getting Started

To get a local copy up and running follow these simple example steps.

### Prerequisites

- Go (version 1.24.5 or later)
- Node.js and npm (for Tailwind CSS)

### Installation

1. **Clone the repo**
   ```sh
   git clone https://github.com/grayankit/go_tts.git
   ```
2. **Install NPM packages**
   ```sh
   npm install
   ```
3. **Build the CSS**
   ```sh
   npm run css:build
   ```
4. **Run the application**
   ```sh
   go run cmd/go_tts/main.go
   ```

## Docker

This project includes a Docker setup for both development and production environments.

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Development

The development environment is configured with hot-reloading for both the Go backend and the frontend assets. To start the development environment, run:

```sh
docker-compose up dev
```

The application will be available at `http://localhost:3001`.

### Production

The production environment is an optimized, multi-stage build. To build and run the production environment, use:

```sh
docker-compose up prod
```

The application will be available at `http://localhost:3001`.

## Usage

1. Open your browser and navigate to `http://localhost:3001`.
2. Enter the text you want to convert in the text area.
3. Click the "Speak" button to listen to the audio.

## API Endpoints

- `GET /` - Serves the main HTML page.
- `POST /api/tts` - Accepts a JSON payload with a "text" field and returns an audio stream.
- `GET /api/history` - Returns the history of synthesized text.
- `GET /api/voices` - Returns a list of available voices.
- `GET /api/preview?voice={voice}` - Returns a preview of the selected voice.
- `GET /events` - SSE endpoint for real-time events.
- `POST /api/speak` - Accepts a JSON payload with "text" and "voice" to speak.
- `GET /api/queue` - Returns the current queue of text to be spoken.
- `POST /api/pause` - Pauses the current speech synthesis.

## Built With

- [Go](https://golang.org/)
- [Tailwind CSS](https://tailwindcss.com/)


Project Link: [https://github.com/grayankit/go_tts](https://github.com/grayankit/go_tts)

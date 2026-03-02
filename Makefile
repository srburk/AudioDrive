-include .env

DATABASE_URL  ?= postgres://localhost/audiodrive?sslmode=disable
BASE_URL      ?= http://localhost:8080
PORT          ?= 8080
TTS_ENDPOINT  ?= https://api.openai.com/v1/audio/speech
TTS_API_KEY   ?=
AUDIO_DIR     ?= $(CURDIR)/audio

export DATABASE_URL BASE_URL PORT TTS_ENDPOINT TTS_API_KEY AUDIO_DIR

.PHONY: all api caddy funnel stop logs clean test build

## Start everything (API + caddy) in the background
all: build audio
	@$(MAKE) -j2 api caddy

## Build the API binary
build:
	go build -o bin/api .

## Run the HTTP API (foreground — used via make -j)
api: bin/api
	./bin/api

## Run Caddy (foreground — used via make -j)
caddy:
	caddy run --config Caddyfile

## Expose port 80 via Tailscale Funnel (Tailscale handles TLS)
funnel:
	tailscale funnel 80

## Create the audio output directory
audio:
	mkdir -p $(AUDIO_DIR)

## Run unit tests
test:
	go test ./...

## Remove built binaries
clean:
	rm -rf bin/

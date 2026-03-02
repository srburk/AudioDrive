# TODO

## Remove audio GET handler from the API server

The `GET /audio/{id}` route and `AudioStore` exist to stream audio files through the Go app. This is redundant once Caddy serves `web/` statically. The intent is for audio to live in `web/audio/` and be served directly by Caddy, not proxied through the API server.

Steps:
- Set `AUDIO_DIR=web/audio` so the worker writes files into the Caddy-served directory
- Update `feed/builder.go` to derive enclosure URLs from the actual filename (`filepath.Base(*u.AudioPath)`) rather than the Go route pattern (`/audio/{id}`)
- Remove `handler.GetAudio`, the `AudioStore` interface, `store.DiskAudioStore`, and related tests
- Remove the `GET /audio/{id}` route from `internal/server/server.go`
- Remove the `audioStore` wiring from `main.go`

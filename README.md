# proxy-replay
A lightweight HTTP proxy that records and replays API responses using request headers. Currently intended for local development and test.

## How to run
Start up:
```
go run ./cmd/app/main.go --dir mocks --port :8081
```
This will:
* save mocked response to a directory called "mocks"
* start up a server on port 8081 that you can send requests to for either recording or playback.

## Usage
Record:
```
curl http://localhost:8081 \
	-H "X-Proxy-Replay-Mode: record" \
  	-H "X-Proxy-Replay-Key: test2" \
  	-H "X-Proxy-Target-URL: https://httpbin.org/uuid"
```
This will:
* forward the request to the target URL
* store the response as "test2.json"
* return the real response to the client

Replay:
```
curl http://localhost:8081 \
  	-H "X-Proxy-Replay-Mode: replay" \
  	-H "X-Proxy-Replay-Key: test2"
```
This will:
* return the previously recorded response
* skip calling any external service, the response comes from proxy-replay


## Notes
* Response bodies are stored as base64
* Headers are preserved as-is
* keys must be managed by the user, using the same key twice will overwrite previous saves.
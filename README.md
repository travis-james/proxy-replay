# proxy-replay
A lightweight HTTP proxy that can record and replay API responses using headers.

## How to run
Start up:
```
go run ./cmd/app/main.go --dir mocks --port :8081
```
## How to use it
Record:
```
curl http://localhost:8081 \
	-H "X-Proxy-Replay-Mode: record" \
  	-H "X-Proxy-Replay-Key: test2" \
  	-H "X-Proxy-Target-URL: https://httpbin.org/uuid"
```

Replay:
```
curl http://localhost:8081 \
  	-H "X-Proxy-Replay-Mode: replay" \
  	-H "X-Proxy-Replay-Key: test2"
```
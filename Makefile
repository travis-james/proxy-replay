run:
	go run ./cmd/app/main.go --dir mocks --port :8081
record:
	curl http://localhost:8081 \
		-H "X-Proxy-Replay-Mode: record" \
  		-H "X-Proxy-Replay-Key: test2" \
  		-H "X-Proxy-Target-URL: https://httpbin.org/uuid"
replay:
	curl http://localhost:8081 \
  		-H "X-Proxy-Replay-Mode: replay" \
  		-H "X-Proxy-Replay-Key: test2"
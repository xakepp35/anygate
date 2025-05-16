.PHONY: test cover benchmark

test:
	go test -count=10 -race ./...

cover:
	go test -short -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out


benchmark:
	go test -run=^$$ -bench=. -benchmem ./...
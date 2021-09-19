build:
	go build -o=app .

run: build
	./app
	
test:
	go test -race -timeout=5s ./...

generate:
	go generate ./...

lint:
	golangci-lint run

fmt:
	gofumpt -l -w .

checks: test lint fmt

build:
	go build -o=app .

run: build
	./app
	
test:
	go test ./...

generate:
	go generate ./...

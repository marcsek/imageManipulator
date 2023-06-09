build:
	@go build -o bin/imageManipulator

run: build
	@./bin/imageManipulator

test:
	@go test -v ./...
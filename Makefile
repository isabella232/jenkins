all:
	go fmt
	go vet
	godep go test
	godep go build

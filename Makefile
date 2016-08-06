all:
	go fmt
	go vet
	go test
	go install

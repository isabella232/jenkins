all:
	go fmt
	go vet
	glide install
	go test
	go install

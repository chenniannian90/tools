test: tidy
	go test -v -race ./...

cover:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

tidy:
	go mod tidy
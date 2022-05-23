APP_NAME = pocmon
GOBIN = $(GOPATH)/bin

go-build:
	GOOS=linux GOARCH=amd64 go build -o ./build/$(APP_NAME)_linux_64 ./
	GOOS=darwin GOARCH=amd64 go build -o ./build/$(APP_NAME)_darwin_64 ./

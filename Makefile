GOFILES := $(shell find . -name "*.go")
BINARY_NAME := garagedoor-service

.PHONY: armv6
armv6: .dist/$(BINARY_NAME).armv6

.dist/g$(BINARY_NAME).armv6: $(GOFILES)
	GOARCH=arm GOARM=6 go build -o .dist/$(BINARY_NAME).armv6 main.go

.PHONY: armv7
armv7: .dist/$(BINARY_NAME).armv7

.dist/$(BINARY_NAME).armv7: $(GOFILES)
	GOARCH=arm GOARM=7 go build -o .dist/$(BINARY_NAME).armv7 main.go

.PHONY: arm64
arm64: .dist/$(BINARY_NAME).arm64

.dist/$(BINARY_NAME).arm64: $(GOFILES)
	GOARCH=arm64 go build -o .dist/$(BINARY_NAME).arm64 main.go

.PHONY: amd64
clean:
	rm -rf .dist

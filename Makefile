APP_NAME=kbot
VERSION=1.0.2
IMAGE_TAG=quay.io/paranoidlookup/$(APP_NAME):latest


.PHONY: linux arm64 macos windows image clean

linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X github.com/mexxo-dvp/kbot/cmd.appVersion=$(VERSION)" -o bin/$(APP_NAME)-linux-amd64 main.go

arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-X github.com/mexxo-dvp/kbot/cmd.appVersion=$(VERSION)" -o bin/$(APP_NAME)-linux-arm64 main.go

macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X github.com/mexxo-dvp/kbot/cmd.appVersion=$(VERSION)" -o bin/$(APP_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X github.com/mexxo-dvp/kbot/cmd.appVersion=$(VERSION)" -o bin/$(APP_NAME)-darwin-arm64 main.go

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-X github.com/mexxo-dvp/kbot/cmd.appVersion=$(VERSION)" -o bin/$(APP_NAME)-windows-amd64.exe main.go

image:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE_TAG) .

clean:
	rm -rf bin/
	docker rmi $(IMAGE_TAG) || true

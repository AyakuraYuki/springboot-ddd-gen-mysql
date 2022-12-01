#!/usr/bin/env sh
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o phoenix-ddd-gen-mysql-macos-arm64
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o phoenix-ddd-gen-mysql-macos-x64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o phoenix-ddd-gen-mysql-win-x64.exe
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o phoenix-ddd-gen-mysql-linux-amd64

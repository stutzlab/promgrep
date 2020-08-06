#!/bin/sh

echo "Removing previous dist"
rm /dist/promgrep-linux-amd64
rm /dist/promgrep-linux-raspberry
rm /dist/promgrep-darwin-amd64
rm /dist/promgrep-windows-amd64.exe

set -e

echo "Compile for linux-amd64"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/promgrep-linux-amd64
chmod +x /dist/promgrep-linux-amd64
echo "Saved to /dist/promgrep-linux-amd64"

echo "Compile for linux-arm (works on Raspberry)"
GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/promgrep-linux-raspberry
chmod +x /dist/promgrep-linux-amd64
echo "Saved to /dist/promgrep-linux-amd64"

echo "Compile for darwin-amd64"
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/promgrep-darwin-amd64
chmod +x /dist/promgrep-darwin-amd64
echo "Saved to /dist/promgrep-darwin-amd64"

echo "Compile for windows-amd64"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/promgrep-windows-amd64.exe
echo "Saved to /dist/promgrep-windows-amd64"


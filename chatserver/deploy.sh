#!/bin/bash

#NOTE: this "deploy" script is temporary until the deploy pipeline and containerization are final. This is for testing purposes to save time

GOOS=linux GOARCH=arm64 go build -o chatserver

scp chatserver jwa_h@raspberrypi.local:gochat
scp homepage.txt jwa_h@raspberrypi.local:gochat

cd ../gochat

GOOS=linux GOARCH=arm64 go build -o ../chatserver/bin/clients/linux/gochat
GOOS=darwin GOARCH=arm64 go build -o ../chatserver/bin/clients/mac/gochat
GOOS=windows GOARCH=amd64 go build -o ../chatserver/bin/clients/windows/gochat

cd ../chatserver

scp bin/clients/linux/gochat jwa_h@raspberrypi.local:gochat/bin/clients/linux
scp bin/clients/mac/gochat jwa_h@raspberrypi.local:gochat/bin/clients/mac
scp bin/clients/windows/gochat jwa_h@raspberrypi.local:gochat/bin/clients/windows

#ssh jwa_h@raspberrypi.local ./gochat/chatserver
#cd gochat
#./chatserver




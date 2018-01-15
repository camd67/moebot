#!/usr/bin/env bash

echo "Downloading required dev tools..."
# required stuff for setup (tools and stuff)
go get -u github.com/jteeuwen/go-bindata github.com/golang/dep/cmd/dep

echo "Downloading required packages..."
# required stuff for programming
cd ./moebot_bot/
dep ensure
cd ../

echo "Compiling base static assets for webserver..."
go-bindata -o ./www/bindata.go ./www/static/...

echo "Done!"
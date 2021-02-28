#!/bin/bash
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

rm -rf ./build && \
go get && \
go build -o build/main main.go && \
zip build/main.zip build/main #&& \
#aws lambda update-function-code --region eu-central-1 --function-name tzbot --zip-file fileb://./build/main.zip --publish


#GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w' -a -installsuffix cgo  -o main main.go &&  zip build/main.zip main
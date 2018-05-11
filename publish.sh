#!/bin/bash
set GOOS=linux

rm -rf ./build && \
go get && \
go build -o build/main main.go && \
zip build/main.zip build/main #&& \
#aws lambda update-function-code --region eu-central-1 --function-name tzbot --zip-file fileb://./build/main.zip --publish

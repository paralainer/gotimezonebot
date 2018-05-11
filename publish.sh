rm -rf ./build && \
go get && GOOS=linux go build -o build/main && \
zip build/main.zip build/main && \
aws lambda update-function-code --region eu-central-1 --function-name tzbot --zip-file fileb://./build/main.zip

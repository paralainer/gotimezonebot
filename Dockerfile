FROM golang:1.8.1-alpine

WORKDIR /go/src/gotimezonebot
COPY . .

RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."

CMD ["go-wrapper", "run"]
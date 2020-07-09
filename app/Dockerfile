FROM golang:1.14-alpine

WORKDIR /root/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/app

CMD ["./out/app"]

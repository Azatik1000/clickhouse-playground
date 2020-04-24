FROM golang:1.14

WORKDIR /root/project
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["server"]

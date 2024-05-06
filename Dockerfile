FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY config/ ./
COPY *.go ./

RUN go mod tidy

RUN go build -o /device-collector

EXPOSE 8080

CMD [ "/device-collector" ]
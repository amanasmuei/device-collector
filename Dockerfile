FROM golang:1.16-alpine

WORKDIR /app

COPY /device-collector/go.mod ./
RUN go mod download

COPY /device-collector/config/ ./
COPY *.go ./

RUN go build -o /device-collector

EXPOSE 8080

CMD [ "/device-collector" ]
FROM golang:latest as go-builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=go-builder /app/faucet /app/

EXPOSE 3000

ENTRYPOINT ["/app/faucet"]

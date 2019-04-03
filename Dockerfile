FROM golang:latest as builder

WORKDIR /app
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch
COPY --from=builder /app/faucet /app/

EXPOSE 3000
CMD ["/app/faucet"]

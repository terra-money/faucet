# Build faucet API
FROM golang:latest as go-builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# Build front end
FROM node:lts-alpine as node-builder

WORKDIR /app

COPY frontend/package.json .
COPY frontend/package-lock.json .

RUN npm i

COPY frontend .

RUN npm run build

# Copy essential files from build images
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=go-builder /app/faucet /app/
COPY --from=node-builder /app/build /app/frontend/build

EXPOSE 3000

ENTRYPOINT ["/app/faucet"]

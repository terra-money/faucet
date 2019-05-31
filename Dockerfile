# node build
FROM node:lts-alpine as node-builder
WORKDIR /app/frontend
COPY /frontend/package*.json ./
RUN yarn install
COPY /frontend/ .
RUN yarn run build

# go build
FROM golang:latest as go-builder
WORKDIR /app
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# staging
FROM alpine:latest
RUN apk add --update ca-certificates
WORKDIR /app
COPY --from=node-builder /app/frontend/build /app/frontend/build/
COPY --from=go-builder /app/faucet /app/

EXPOSE 3000
CMD ["/app/faucet"]

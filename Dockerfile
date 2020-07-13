FROM golang:latest as go-builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build


FROM node:lts-alpine as node-builder

WORKDIR /app

COPY frontend/package.json .
COPY frontend/yarn.lock .

RUN yarn install

COPY frontend .

RUN yarn run build


FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=go-builder /app/faucet /app/
COPY --from=node-builder /app/build /app/frontent/build

EXPOSE 3000

ENTRYPOINT ["/app/faucet"]

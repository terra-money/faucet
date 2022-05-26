FROM golang:1.18.2-alpine3.15 as go-builder

ARG arch=x86_64
ENV LIBWASMVM_VERSION=v1.0.0

ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 7d2239e9f25e96d0d4daba982ce92367aacf0cbd95d2facb8442268f2b1cc1fc
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep f6282df732a13dec836cda1f399dd874b1e3163504dbd9607c6af915b2740479

RUN cp /lib/libwasmvm_muslc.${arch}.a /lib/libwasmvm_muslc.a

RUN set -eux; apk add --no-cache ca-certificates build-base;

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go ./

#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 LEDGER_ENABLED=false BUILD_TAGS=muslc go build
RUN GOOS=linux GOARCH=amd64 go build -tags muslc

FROM alpine:3.15.4

RUN apk add --no-cache ca-certificates build-base;

WORKDIR /app

COPY --from=go-builder /app/faucet /app/

EXPOSE 3000

ENTRYPOINT ["/app/faucet"]

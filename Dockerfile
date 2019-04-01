FROM golang:latest

RUN mkdir -p "$GOPATH/src/github.com/cosmos/faucet"
WORKDIR $GOPATH/src/github.com/cosmos/faucet
COPY faucet.go .

RUN go install -v ./...

CMD ["faucet"]

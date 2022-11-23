# Terra Testnet Faucet

![banner](./terra-faucet.png)

Terra Testnet Faucet is a client tool that allows anyone to easily request a nominal amount of Terra or Luna assets for testing purposes. This app needs to be deployed on a Terra testnet full node, because it relies on using the `terracli` command to send tokens.

**WARNING**: Tokens recieved over the faucet are not real assets and have no market value.

This faucet implementation is a fork of the [Cosmos Faucet](https://github.com/cosmos/faucet).

## Get tokens on Terra testnets

Using the testnets is really easy. Simply go to https://faucet.terra.money, chose your network and input your testnet address. 

## Usage

Build the docker image.

```bash
docker build -t faucet .
```

Run it with the mnemonic and recaptcha key as env vars.

```bash
# Mnemonic must be splitted by _
export MNEMONIC=
export PORT=4501

docker run -p $PORT:$PORT \
    -e MNEMONIC=$MNEMONIC \
    -e PORT=$PORT \
    faucet
```

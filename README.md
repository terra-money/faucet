# Mars Hub Testnet Faucet

![banner](./banner.png)

Mars Hub testnet Faucet is a client tool that allows anyone to easily request a nominal amount of Terra or Luna assets for testing purposes.

**WARNING**: Tokens recieved over the faucet are not real assets and have no market value.

This faucet implementation is a fork of the [Terra Faucet](https://github.com/terra-money/faucet).

## Get tokens on Mars testnet

Using the testnets is really easy. Simply go to https://faucet.marsprotocol.io and connect your (Keplr) wallet.

## Usage

Build the docker image.

```bash
docker build -t faucet .
```

Run it with the mnemonic and recaptcha key as env vars.

```bash
docker run -p 3000:3000 \
    -e MNEMONIC=$MY_MNEMONIC \
    -e RECAPTCHA_KEY=$RECAPTCHA_KEY \
    -e PORT=8080 \  # default to 3000
    faucet
```

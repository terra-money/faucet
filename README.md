# Terra Testnet Faucet

This faucet app allows anyone to easily request 10 faucetToken and 1 steak. This app needs to be deployed on a Cosmos testnet full node, because it relies on using the `terracli` command to send tokens.

## Get reCAPTCHA Key

Go to the [Google reCAPTCHA Admin](https://www.google.com/recaptcha/admin) and create a new reCAPTCHA site. For the version of captcha, choose `reCAPTCHA v2`.

In the file `./frontend/src/views/Faucet.vue` on line 60, change the `sitekey` to your new reCAPTCHA client side integration site key.

```
sitekey: "6LdqyV0UAAAAAEqgBxvSsDpL2aeTEgkz_VTz1Vi1"
```

## Set ENV Variables

The faucet requires 4 different enviroment variables to set in order to function. They are: 

1. `KEY`, the name of your faucet account.
2. `NODE`, the address of your `terrad` node (probably don't have to change)
3. `CHAIN`, the chain id of the testnet.
4. `PASS`, the password of your faucet account.

And here are the app's defaults if you don't set any environment variables:

```
key = os.Getenv("KEY")
if key == "" {
  key = "faucet"
}

node = os.Getenv("NODE")
if node == "" {
  node = "http://localhost:46657"
}

chain = os.Getenv("CHAIN")
if chain == "" {
  chain = "soju-0005"
}

pass = os.Getenv("PASS")
if pass == "" {
  pass = "1234567890"
}
```

## Build

You need to have Golang and Yarn/Node.js installed on your system.

```
go get $GOPATH/src/github.com/terra-project/faucet
cd $GOPATH/src/github.com/terra-project/faucet
dep ensure

cd frontend
yarn && yarn build
cd ..
```

## Deploy

This will run the faucet on port 8080. It's highly recommended that you run a reverse proxy with rate limiting in front of this app.

```
go run faucet.go RECAPTCHA_SERVER_SIDE_SECRET
```

## Optional: Caddy

Included in this repo is an example `Caddyfile` that lets you run an TLS secured faucet that is rate limited to 1 claim per IP per day.

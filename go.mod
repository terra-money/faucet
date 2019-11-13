module github.com/terra-project/faucet

require (
	github.com/btcsuite/btcd v0.0.0-20190605094302-a0d1e3e36d50 // indirect
	github.com/cosmos/cosmos-sdk v0.0.0-00010101000000-000000000000
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/dpapathanasiou/go-recaptcha v0.0.0-20180330231321-0e9736be20f9
	github.com/syndtr/goleveldb v1.0.0
	github.com/tendermint/go-amino v0.15.0 // indirect
	github.com/tendermint/iavl v0.12.2 // indirect
	github.com/tendermint/tendermint v0.31.11
	github.com/tendermint/tmlibs v0.0.0-20180607034639-640af0205d98
	github.com/terra-project/core v0.2.6
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
)

replace github.com/cosmos/cosmos-sdk => github.com/YunSuk-Yeo/cosmos-sdk v0.35.5-terra

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

go 1.13

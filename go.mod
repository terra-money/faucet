module github.com/terra-project/faucet

require (
	github.com/btcsuite/btcd v0.0.0-20190605094302-a0d1e3e36d50 // indirect
	github.com/cosmos/cosmos-sdk v0.0.0-00010101000000-000000000000
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/dpapathanasiou/go-recaptcha v0.0.0-20180330231321-0e9736be20f9
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/syndtr/goleveldb v1.0.0
	github.com/tendermint/go-amino v0.15.0 // indirect
	github.com/tendermint/iavl v0.12.2 // indirect
	github.com/tendermint/tendermint v0.31.5
	github.com/tendermint/tmlibs v0.0.0-20180607034639-640af0205d98
	github.com/terra-project/core v0.2.1
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	golang.org/x/crypto v0.0.0-20190411191339-88737f569e3a // indirect
	google.golang.org/genproto v0.0.0-20180831171423-11092d34479b // indirect
)

replace github.com/cosmos/cosmos-sdk => github.com/YunSuk-Yeo/cosmos-sdk v0.34.7-terra

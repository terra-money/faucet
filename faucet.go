package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/cors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tmlibs/bech32"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/terra-money/core/v2/app"
	"github.com/terra-money/core/v2/app/params"

	//"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	//"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/client"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var mnemonic string
var port string
var privKey cryptotypes.PrivKey

var sequence uint64
var accountNumber uint64
var mtx sync.Mutex

type Network struct {
	// DEFAULTS
	chainID string
	denom   string
	prefix  string
	lcdURL  string

	// COMPUTED
	faucetAddress string
	cdc           params.EncodingConfig
	txBuilder     client.TxBuilder
}

var networks = []Network{
	{
		chainID: "alliance-testnet-1",
		denom:   "stake",
		prefix:  "alliance",
		lcdURL:  "http://3.75.187.158:1317",
	},
	{
		chainID: "pisco-1",
		denom:   "uluna",
		prefix:  "terra",
		lcdURL:  "http://18.194.243.144:1317",
	},
}

const (
	MicroUnit   = int64(1e6)
	accountPath = "m/44'/118'/0'/0/0"
	coinType    = 118
	SendAmount  = 5 * MicroUnit
)

const (
	requestLimitSecs = 30
	mnemonicVar      = "MNEMONIC"
	portVar          = "PORT"
)

// Claim wraps a faucet claim
type Claim struct {
	Address string `json:"address"`
	ChainID string `json:"chainID"`
}

// Coin is the same as sdk.Coin
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

func setCodecs(net *Network) {
	net.cdc = app.MakeEncodingConfig()

	config := sdk.GetConfig()

	config.SetCoinType(coinType)
	config.SetBech32PrefixForAccount(net.prefix, net.prefix+"pub")
	config.SetBech32PrefixForValidator(net.prefix+"valoper", net.prefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(net.prefix+"valcons", net.prefix+"valconspub")

	net.txBuilder = app.MakeEncodingConfig().TxConfig.NewTxBuilder()
}

func setKeys() {
	derivedPriv, err := hd.Secp256k1.Derive()(mnemonic, "", accountPath)
	if err != nil {
		panic(err)
	}

	privKey = hd.Secp256k1.Generate()(derivedPriv)
	pubKey := privKey.PubKey()

	for i := 0; i < len(networks); i++ {
		network := &networks[i]
		address, err := bech32.ConvertAndEncode(network.prefix, pubKey.Address())
		if err != nil {
			panic(err)
		}

		network.faucetAddress = address
	}
}

func loadAccountInfo(net Network) {
	// Query current faucet sequence
	url := fmt.Sprintf("%v/cosmos/auth/v1beta1/accounts/%v", net.lcdURL, net.faucetAddress)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bodyStr := string(body)
	var seq uint64

	if strings.Contains(bodyStr, `"sequence"`) {
		seq, _ = strconv.ParseUint(parseRegexp(`"sequence": "?(\d+)"?`, bodyStr), 10, 64)
	} else {
		seq = 0
	}

	sequence = atomic.LoadUint64(&seq)

	if strings.Contains(bodyStr, `"account_number"`) {
		accountNumber, _ = strconv.ParseUint(parseRegexp(`"account_number": "?(\d+)"?`, bodyStr), 10, 64)
	} else {
		accountNumber = 0
	}

	fmt.Printf("Faucet: %v \n address: %v\n number: %v\n sequence: %v\n", net.lcdURL, net.faucetAddress, accountNumber, sequence)
}

type CoreCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type BalanceResponse struct {
	Balance CoreCoin `json:"balance"`
}

func getBalance(address, lcdURL string) (amount int64) {
	url := fmt.Sprintf("%v/cosmos/bank/v1beta1/balances/%v/by_denom?denom=stake", lcdURL, address)
	response, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	var res BalanceResponse

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&res)

	if err != nil {
		panic(err)
	}

	i, _ := strconv.ParseInt(res.Balance.Amount, 10, 64)
	return i
}

func parseRegexp(regexpStr string, target string) (data string) {
	// Capture seqeunce string from json
	r := regexp.MustCompile(regexpStr)
	groups := r.FindStringSubmatch(string(target))

	if len(groups) != 2 {
		os.Exit(1)
	}

	// Convert sequence string to int64
	data = groups[1]
	return
}

// RequestLog stores the Log of a Request
type RequestLog struct {
	Coins     []Coin    `json:"coin"`
	Requested time.Time `json:"updated"`
}

func (requestLog *RequestLog) dripCoin(denom string) error {

	// try to update coin
	for idx, coin := range requestLog.Coins {
		if coin.Denom == denom {
			fmt.Printf("\nAMOUNT: %v | SendAmount %v\n", requestLog.Coins[idx].Amount, SendAmount*2)
			if (requestLog.Coins[idx].Amount + SendAmount) > SendAmount*2 {
				return errors.New("amount limit exceeded")
			}

			requestLog.Coins[idx].Amount += SendAmount
			return nil
		}
	}

	// first drip for denom
	requestLog.Coins = append(requestLog.Coins, Coin{Denom: denom, Amount: SendAmount})
	return nil
}

func checkAndUpdateLimit(db *leveldb.DB, account []byte, net Network) error {
	address, _ := bech32.ConvertAndEncode(net.prefix, account)
	addressBalance := getBalance(address, net.lcdURL)

	if addressBalance >= SendAmount*2 {
		return errors.New("amount limit exceeded")
	}

	var requestLog RequestLog

	logBytes, _ := db.Get(account, nil)
	now := time.Now()

	if logBytes != nil {
		jsonErr := json.Unmarshal(logBytes, &requestLog)
		if jsonErr != nil {
			return jsonErr
		}

		// check interval limt
		intervalSecs := now.Sub(requestLog.Requested).Seconds()
		if intervalSecs < requestLimitSecs {
			return errors.New("please wait a while for another tap")
		}

		// reset log if month was changed
		if requestLog.Requested.Month() != now.Month() {
			requestLog.Coins = []Coin{}
		}

		// check amount limit
		err := requestLog.dripCoin(net.denom)
		if err != nil {
			return err
		}
	}

	// update requested time
	requestLog.Requested = now
	logBytes, _ = json.Marshal(requestLog)
	updateErr := db.Put(account, logBytes, nil)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func drip(net Network, encodedAddress string, amount int64, isDetectMismatch bool) string {
	net.txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(net.denom, sdk.NewInt(200_000))))
	net.txBuilder.SetGasLimit(200_000)
	net.txBuilder.SetMemo("Terra Faucet")
	net.txBuilder.SetTimeoutHeight(0)

	coins := sdk.NewCoins(sdk.NewCoin(net.denom, sdk.NewInt(amount)))

	from, err := sdk.AccAddressFromBech32(net.faucetAddress)
	if err != nil {
		panic(err)
	}

	to, err := sdk.AccAddressFromBech32(encodedAddress)
	if err != nil {
		panic(err)
	}

	sendMsg := banktypes.NewMsgSend(from, to, coins)
	net.txBuilder.SetMsgs(sendMsg)

	return signAndBroadcast(net, isDetectMismatch)
}

func createGetCoinsHandler(db *leveldb.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				http.Error(w, err.(error).Error(), 400)
			}
		}()

		// Decode JSON response from front end
		// and assign the value to claim variable
		var claim Claim
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&claim)
		if err != nil {
			panic(err)
		}

		// setup network
		var net Network
		for _, n := range networks {
			if n.chainID == claim.ChainID {
				net = n
				break
			}
		}
		// Network not found
		if net.chainID == "" {
			panic(`Network not found for chainID: ` + claim.ChainID)
		}

		// Setup Codecs
		setCodecs(&net)

		// Load account
		loadAccountInfo(net)

		// make sure address is bech32
		readableAddress, decodedAddress, err := bech32.DecodeAndConvert(claim.Address)
		if err != nil {
			panic(err)
		}

		// re-encode the address in bech32
		encodedAddress, err := bech32.ConvertAndEncode(readableAddress, decodedAddress)
		if err != nil {
			panic(err)
		}

		// Limiting request speed
		limitErr := checkAndUpdateLimit(db, decodedAddress, net)
		if limitErr != nil {
			panic(limitErr)
		}

		// send the coins!
		mtx.Lock()
		defer mtx.Unlock()

		fmt.Printf("[%v][REQUEST] Address: %v Coins: %v%v\n", time.Now().UTC().Format(time.RFC3339), encodedAddress, SendAmount, net.denom)
		body := drip(net, encodedAddress, SendAmount, true)

		// Sequence mismatch if the body length is zero
		if len(body) == 0 {
			// Reload for self healing and re-drip
			loadAccountInfo(net)
			body = drip(net, encodedAddress, SendAmount, true)

			// Another try without loading....
			if len(body) == 0 {
				atomic.AddUint64(&sequence, 1)
				body = drip(net, encodedAddress, SendAmount, false)
			}
		}

		if len(body) != 0 {
			atomic.AddUint64(&sequence, 1)
		}

		fmt.Printf("[%v][RESPONSE] Sequence: %v Body: %v\n", time.Now().UTC().Format(time.RFC3339), sequence, body)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"amount": %v, "response": %v}`, SendAmount, body)
	}
}

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx   string `json:"tx_bytes"`
	Mode string `json:"mode"`
}

func signAndBroadcast(net Network, isDetectMismatch bool) string {
	var broadcastReq BroadcastReq
	pubKey := privKey.PubKey()
	signMode := net.cdc.TxConfig.SignModeHandler().DefaultMode()

	signerData := signing.SignerData{
		ChainID:       net.chainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
	sigData := txsigning.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sigv2 := txsigning.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: sequence,
	}
	net.txBuilder.SetSignatures(sigv2)
	bytesToSign, err := net.cdc.TxConfig.SignModeHandler().GetSignBytes(signMode, signerData, net.txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	sig, err := privKey.Sign(bytesToSign)
	if err != nil {
		panic(err)
	}

	sigData = txsigning.SingleSignatureData{
		SignMode:  signMode,
		Signature: sig,
	}
	sigv2 = txsigning.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: sequence,
	}

	net.txBuilder.SetSignatures(sigv2)

	// encode signed tx
	bz, err := net.cdc.TxConfig.TxEncoder()(net.txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	// prepare to broacast
	broadcastReq.Tx = base64.StdEncoding.EncodeToString(bz)
	broadcastReq.Mode = "BROADCAST_MODE_SYNC"

	txBz, err := json.Marshal(broadcastReq)
	if err != nil {
		panic(err)
	}

	// broadcast
	url := fmt.Sprintf("%v/cosmos/tx/v1beta1/txs", net.lcdURL)
	response, err := http.Post(url, "application/json", bytes.NewReader(txBz))
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	stringBody := string(body)
	fmt.Print(stringBody)

	if response.StatusCode != 200 {
		err := fmt.Errorf("status: %v, message: %v", response.Status, stringBody)
		panic(err)
	}

	code, err := strconv.ParseUint(parseRegexp(`"code": ?(\d+)?`, stringBody), 10, 64)
	if err != nil {
		panic(err)
	}
	if code != 0 {
		msg := parseRegexp(`"raw_log": "?(\d+)"?`, stringBody)
		panic(fmt.Errorf("transaction failed: %s", msg))
	}

	if isDetectMismatch && strings.Contains(stringBody, "sequence mismatch") {
		return ""
	}

	return stringBody
}

func main() {
	mnemonic = os.Getenv(mnemonicVar)
	// splitted by "_" (underscore) instead of " " (space)
	// because the spaces are not correctly computed as
	// single arguments on docker
	mnemonic = strings.ReplaceAll(mnemonic, "_", " ")
	if mnemonic == "" {
		panic("MNEMONIC is required")
	}
	if !bip39.IsMnemonicValid(mnemonic) {
		panic("invalid mnemonic: " + mnemonic)
	}

	port = os.Getenv(portVar)
	if port == "" {
		port = "4501"
	}

	db, err := leveldb.OpenFile("db/ipdb", nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	setKeys()

	// Setup the faucet server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, `OK!`)
	})
	mux.HandleFunc("/claim", createGetCoinsHandler(db))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://3.75.187.158", "http://localhost:3000"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	for _, net := range networks {
		loadAccountInfo(net)
	}
	fmt.Printf("Server listening on port: %s\n", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		log.Fatal("failed to start server", err)
	}
}

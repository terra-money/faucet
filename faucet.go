package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/rs/cors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tmlibs/bech32"
	"github.com/tomasen/realip"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/mars-protocol/hub/app"

	//"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	//"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/client"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var mnemonic string
var recaptchaKey string
var port string
var lcdURL string
var chainID string
var privKey cryptotypes.PrivKey

//var privKey crypto.PrivKey
var address string
var sequence uint64
var accountNumber uint64
var cdc *app.EncodingConfig
var mtx sync.Mutex
var isClassic bool

type PagerdutyConfig struct {
	token     string
	user      string
	serviceID string
}

var pagerdutyConfig PagerdutyConfig

const ( // new core hasn't these yet.
	MicroUnit              = int64(1e6)
	fullFundraiserPath     = "m/44'/118'/0'/0/0"
	accountPubKeyPrefix    = "marspub"
	validatorAddressPrefix = "marsvaloper"
	validatorPubKeyPrefix  = "marsvaloperpub"
	consNodeAddressPrefix  = "marsvalcons"
	consNodePubKeyPrefix   = "marsvalconspub"

	CoinType    = 118
	CoinPurpose = 44
)

var amountTable = map[string]int64{
	app.BondDenom: 5 * MicroUnit,
}

const (
	requestLimitSecs      = 30
	mnemonicVar           = "MNEMONIC"
	recaptchaKeyVar       = "RECAPTCHA_KEY"
	portVar               = "PORT"
	lcdUrlVar             = "LCD_URL"
	chainIDVar            = "CHAIN_ID"
	pagerdutyTokenVar     = "PAGERDUTY_TOKEN"
	pagerdutyUserVar      = "PAGERDUTY_USER"
	pagerdutyServiceIDVar = "PAGERDUTY_SERVICE_ID"
)

// Claim wraps a faucet claim
type Claim struct {
	Address  string `json:"address"`
	Response string `json:"response"`
	Denom    string `json:"denom"`
}

// Coin is the same as sdk.Coin
type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

func newCodec() *app.EncodingConfig {
	ec := app.MakeEncodingConfig()

	config := sdk.GetConfig()
	config.SetCoinType(CoinType)
	config.SetFullFundraiserPath(fullFundraiserPath)
	config.SetBech32PrefixForAccount(app.AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()

	return &ec
}

func loadAccountInfo() {
	// Query current faucet sequence
	url := fmt.Sprintf("%v/cosmos/auth/v1beta1/accounts/%v", lcdURL, address)
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

	fmt.Printf("loadAccountInfo: address %v account# %v sequence %v\n", address, accountNumber, sequence)
}

type CoreCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type BalanceResponse struct {
	Balance CoreCoin `json:"balance"`
}

func getBalance(address string) (amount int64) {
	url := fmt.Sprintf("%v/cosmos/bank/v1beta1/balances/%v/by_denom?denom=umars", lcdURL, address)
	response, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	var res BalanceResponse

	decoder := json.NewDecoder(response.Body)
	decoderErr := decoder.Decode(&res)

	if decoderErr != nil {
		panic(decoderErr)
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
	amount := amountTable[denom]

	// try to update coin
	for idx, coin := range requestLog.Coins {
		if coin.Denom == denom {
			if (requestLog.Coins[idx].Amount + amount) > amountTable[denom]*2 {
				return errors.New("amount limit exceeded")
			}

			requestLog.Coins[idx].Amount += amount
			return nil
		}
	}

	// first drip for denom
	requestLog.Coins = append(requestLog.Coins, Coin{Denom: denom, Amount: amount})
	return nil
}

func checkAndUpdateLimit(db *leveldb.DB, account []byte, denom string) error {
	address, _ := bech32.ConvertAndEncode(app.AccountAddressPrefix, account)

	if getBalance(address) >= amountTable[denom]*2 {
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
		dripErr := requestLog.dripCoin(denom)
		if dripErr != nil {
			return dripErr
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

func drip(encodedAddress string, denom string, amount int64, isDetectMismatch bool) string {
	builder := app.MakeEncodingConfig().TxConfig.NewTxBuilder()

	builder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(app.BondDenom, sdk.NewInt(200_000))))
	builder.SetGasLimit(150_000)
	builder.SetMemo("faucet")
	builder.SetTimeoutHeight(0)

	coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(amount)))
	from, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic("invalid from address")
	}
	to, err := sdk.AccAddressFromBech32(encodedAddress)
	if err != nil {
		panic("invalid to address")
	}
	sendMsg := banktypes.NewMsgSend(from, to, coins)
	builder.SetMsgs(sendMsg)

	return signAndBroadcast(builder, isDetectMismatch)
}

func createGetCoinsHandler(db *leveldb.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				http.Error(w, err.(error).Error(), 400)
			}
		}()

		var claim Claim

		// decode JSON response from front end
		decoder := json.NewDecoder(request.Body)
		decoderErr := decoder.Decode(&claim)

		if decoderErr != nil {
			panic(decoderErr)
		}

		amount, ok := amountTable[claim.Denom]

		if !ok {
			panic(fmt.Errorf("invalid denom; %v", claim.Denom))
		}

		// make sure address is bech32
		readableAddress, decodedAddress, decodeErr := bech32.DecodeAndConvert(claim.Address)
		if decodeErr != nil {
			panic(decodeErr)
		}
		// re-encode the address in bech32
		encodedAddress, encodeErr := bech32.ConvertAndEncode(readableAddress, decodedAddress)
		if encodeErr != nil {
			panic(encodeErr)
		}

		// make sure captcha is valid
		clientIP := realip.FromRequest(request)
		captchaResponse := claim.Response
		captchaPassed, captchaErr := recaptcha.Confirm(clientIP, captchaResponse)
		if captchaErr != nil {
			panic(captchaErr)
		}

		// Limiting request speed
		limitErr := checkAndUpdateLimit(db, decodedAddress, claim.Denom)
		if limitErr != nil {
			panic(limitErr)
		}

		// send the coins!
		if captchaPassed {
			mtx.Lock()
			defer mtx.Unlock()

			fmt.Println(time.Now().UTC().Format(time.RFC3339), "req", clientIP, encodedAddress, amount, claim.Denom)
			body := drip(encodedAddress, claim.Denom, amount, true)

			// Sequence mismatch if the body length is zero
			if len(body) == 0 {
				// Reload for self healing and re-drip
				loadAccountInfo()
				body = drip(encodedAddress, claim.Denom, amount, true)

				// Another try without loading....
				if len(body) == 0 {
					atomic.AddUint64(&sequence, 1)
					body = drip(encodedAddress, claim.Denom, amount, false)
				}
			}

			if len(body) != 0 {
				atomic.AddUint64(&sequence, 1)
			}

			fmt.Printf("%v seq %v %v\n", time.Now().UTC().Format(time.RFC3339), sequence, body)

			// Create an incident for broadcast error
			if (isClassic && strings.Contains(body, "code")) ||
				(!isClassic && !strings.Contains(body, "\"code\": 0")) {
				// createIncident(body)
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"amount": %v, "response": %v}`, amount, body)
		} else {
			err := errors.New("captcha failed, please refresh page and try again")
			panic(err)
		}
	}
}

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx   string `json:"tx_bytes"`
	Mode string `json:"mode"`
}

func signAndBroadcast(txBuilder client.TxBuilder, isDetectMismatch bool) string {
	var broadcastReq BroadcastReq

	pubKey := privKey.PubKey()

	// no need to sort amount because there's only one amount.

	// create signature v2
	signMode := cdc.TxConfig.SignModeHandler().DefaultMode()

	signerData := signing.SignerData{
		ChainID:       chainID,
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
	txBuilder.SetSignatures(sigv2)
	bytesToSign, err := cdc.TxConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
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

	txBuilder.SetSignatures(sigv2)

	// encode signed tx
	bz, err := cdc.TxConfig.TxEncoder()(txBuilder.GetTx())
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
	url := fmt.Sprintf("%v/cosmos/tx/v1beta1/txs", lcdURL)
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

	if response.StatusCode != 200 {
		err := fmt.Errorf("status: %v, message: %v", response.Status, stringBody)
		panic(err)
	}

	code, err := strconv.ParseUint(parseRegexp(`"code": ?(\d+)?`, stringBody), 10, 64)
	if err != nil {
		panic("failed to parse code from tx response")
	}
	if code != 0 {
		panic(parseRegexp(`"raw_log": "?(\d+)"?`, stringBody))
	}

	if isDetectMismatch && strings.Contains(stringBody, "sequence mismatch") {
		return ""
	}

	return stringBody
}

// func createIncident(body string) (*pagerduty.Incident, error) {
// 	client := pagerduty.NewClient(pagerdutyConfig.token)
// 	input := &pagerduty.CreateIncidentOptions{
// 		Title:   "Faucet had an error",
// 		Urgency: "low",
// 		Service: &pagerduty.APIReference{
// 			ID:   pagerdutyConfig.serviceID,
// 			Type: "service",
// 		},
// 		Body: &pagerduty.APIDetails{
// 			Details: body,
// 		},
// 	}

// 	return client.CreateIncidentWithContext(context.Background(), pagerdutyConfig.user, input)
// }

func main() {
	mnemonic = os.Getenv(mnemonicVar)

	if mnemonic == "" {
		panic("MNEMONIC variable is required")
	}
	if !bip39.IsMnemonicValid(mnemonic) {
		panic("invalid mnemonic")
	}

	recaptchaKey = os.Getenv(recaptchaKeyVar)

	if recaptchaKey == "" {
		panic("RECAPTCHA_KEY variable is required")
	}

	port = os.Getenv(portVar)

	if port == "" {
		port = "3000"
	}

	lcdURL = os.Getenv(lcdUrlVar)

	if lcdURL == "" {
		panic("LCD_URL variable is required")
	}

	chainID = os.Getenv(chainIDVar)

	if chainID == "" {
		panic("CHAIN_ID variable is required")
	}
	if strings.HasPrefix(chainID, "bombay") {
		isClassic = true
	} else {
		isClassic = false
	}

	// pagerdutyConfig.token = os.Getenv(pagerdutyTokenVar)
	// pagerdutyConfig.user = os.Getenv(pagerdutyUserVar)
	// pagerdutyConfig.serviceID = os.Getenv(pagerdutyServiceIDVar)

	// if pagerdutyConfig.token == "" || pagerdutyConfig.user == "" || pagerdutyConfig.serviceID == "" {
	// 	panic("PAGERDUTY_TOKEN, PAGERDUTY_USER, and PAGERDUTY_SERVICE_ID variables are required")
	// }

	db, err := leveldb.OpenFile("db/ipdb", nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	cdc = newCodec()

	derivedPriv, err := hd.Secp256k1.Derive()(mnemonic, "", fullFundraiserPath)
	if err != nil {
		panic(err)
	}

	//privKey = *secp256k1.GenPrivKeyFromSecret(derivedPriv)
	privKey = hd.Secp256k1.Generate()(derivedPriv)
	pubk := privKey.PubKey()
	address, err = bech32.ConvertAndEncode(app.AccountAddressPrefix, pubk.Address())
	if err != nil {
		panic(err)
	}

	// Load account number and sequence
	loadAccountInfo()

	recaptcha.Init(recaptchaKey)

	// Application server.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	})
	mux.HandleFunc("/claim", createGetCoinsHandler(db))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://faucet.marsprotocol.io", "http://localhost", "localhost", "http://localhost:3000", "http://localhost:8080"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		log.Fatal("failed to start server", err)
	}
}

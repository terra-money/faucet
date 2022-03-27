package main

import (
	"bytes"
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
	"time"

	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tmlibs/bech32"
	"github.com/tomasen/realip"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bip39 "github.com/cosmos/go-bip39"

	"github.com/terra-project/core/app"
	core "github.com/terra-project/core/types"

	"github.com/rs/cors"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var mnemonic string
var recaptchaKey string
var port string
var lcdURL string
var chainID string
var privKey crypto.PrivKey
var address string
var sequence uint64
var accountNumber uint64
var cdc *codec.Codec

var amountTable = map[string]int64{
	core.MicroLunaDenom: 10 * core.MicroUnit,
}

const (
	requestLimitSecs = 30
	mnemonicVar      = "MNEMONIC"
	recaptchaKeyVar  = "RECAPTCHA_KEY"
	portVar          = "PORT"
	lcdUrlVar        = "LCD_URL"
	chainIDVar       = "CHAIN_ID"
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

func newCodec() *codec.Codec {
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetCoinType(core.CoinType)
	config.SetFullFundraiserPath(core.FullFundraiserPath)
	config.SetBech32PrefixForAccount(core.Bech32PrefixAccAddr, core.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(core.Bech32PrefixValAddr, core.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(core.Bech32PrefixConsAddr, core.Bech32PrefixConsPub)
	config.Seal()

	return cdc
}

func main() {
	db, err := leveldb.OpenFile("db/ipdb", nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	mnemonic = os.Getenv(mnemonicVar)

	if mnemonic == "" {
		panic("MNEMONIC variable is required")
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

	cdc = newCodec()

	seed := bip39.NewSeed(mnemonic, "")
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, core.FullFundraiserPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	privKey = secp256k1.PrivKeySecp256k1(derivedPriv)
	pubk := privKey.PubKey()
	address, err = bech32.ConvertAndEncode("terra", pubk.Address())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Load account number and sequence
	loadAccountInfo()

	recaptcha.Init(recaptchaKey)

	// Pprof server.
	go func() {
		log.Fatal(http.ListenAndServe("localhost:8081", nil))
	}()

	// Application server.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	})
	mux.HandleFunc("/claim", createGetCoinsHandler(db))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://faucet.terra.money"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		log.Fatal("failed to start server", err)
	}
}

func loadAccountInfo() {
	// Query current faucet sequence
	url := fmt.Sprintf("%v/auth/accounts/%v", lcdURL, address)
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
	if strings.Contains(bodyStr, `"sequence"`) {
		sequence, _ = strconv.ParseUint(parseRegexp(`"sequence":"?(\d+)"?`, bodyStr), 10, 64)
	} else {
		sequence = 0
	}

	if strings.Contains(bodyStr, `"account_number"`) {
		accountNumber, _ = strconv.ParseUint(parseRegexp(`"account_number":"?(\d+)"?`, bodyStr), 10, 64)
	} else {
		accountNumber = 0
	}

	fmt.Printf("loadAccountInfo: address %v account# %v sequence %v\n", address, accountNumber, sequence)
}

func parseRegexp(regexpStr string, target string) (data string) {
	// Capture seqeunce string from json
	r := regexp.MustCompile(regexpStr)
	groups := r.FindStringSubmatch(string(target))

	if len(groups) != 2 {
		fmt.Printf("cannot find data")
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
			if (requestLog.Coins[idx].Amount + amount) > amountTable[denom]*10 {
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

		// reset log if date was changed
		if requestLog.Requested.Day() != now.Day() {
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
	data := strings.TrimSpace(fmt.Sprintf(`{
		"type": "core/StdTx",
		"value": {
			"msg": [{
				"type": "bank/MsgSend",
				"value": {
					"from_address": "%v",
					"to_address": "%v",
					"amount": [{
						"denom": "%v",
						"amount": "%v"
					}]
				}
			}],
			"fee": {
				"amount": [{
					"denom": "ukrw",
					"amount": "25500000"
				}],
				"gas": "150000"
			},
			"signatures": [],
			"memo": "%v",
			"timeout_height": "0"
		}
	}`, address, encodedAddress, denom, amount, "faucet"))

	return signAndBroadcast([]byte(data), isDetectMismatch)
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
			body := drip(encodedAddress, claim.Denom, amount, true)

			// Sequence mismatch if the body length is zero
			if len(body) == 0 {
				// Reload for self healing and re-drip
				loadAccountInfo()
				body = drip(encodedAddress, claim.Denom, amount, true)

				// Another try without loading....
				if len(body) == 0 {
					sequence = sequence + 1
					body = drip(encodedAddress, claim.Denom, amount, false)
				}
			}

			if len(body) != 0 {
				sequence = sequence + 1
			}

			fmt.Println(time.Now().UTC().Format(time.RFC3339), encodedAddress, "[1] ", amount, claim.Denom)
			fmt.Println(body)

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
	Tx         auth.StdTx `json:"tx"`
	Mode       string     `json:"mode"`
	Sequences  []uint64   `json:"sequences" yaml:"sequences"`
	FeeGranter string     `json:"fee_granter" yaml:"fee_granter"`
}

func signAndBroadcast(txJSON []byte, isDetectMismatch bool) string {
	var broadcastReq BroadcastReq
	var stdTx auth.StdTx

	cdc.MustUnmarshalJSON(txJSON, &stdTx)

	// Sort denom
	for _, msg := range stdTx.Msgs {
		msg, ok := msg.(bank.MsgSend)
		if ok {
			msg.Amount.Sort()
		}
	}

	signBytes := auth.StdSignBytes(chainID, accountNumber, sequence, stdTx.Fee, stdTx.Msgs, stdTx.Memo)
	sig, err := privKey.Sign(signBytes)
	if err != nil {
		panic(err)
	}

	sigs := []auth.StdSignature{{
		PubKey:    privKey.PubKey(),
		Signature: sig}}
	tx := auth.NewStdTx(stdTx.Msgs, stdTx.Fee, sigs, stdTx.Memo)
	broadcastReq.Tx = tx
	broadcastReq.Mode = "sync"
	broadcastReq.Sequences = []uint64{sequence}

	bz := cdc.MustMarshalJSON(broadcastReq)

	url := fmt.Sprintf("%v/txs", lcdURL)
	response, err := http.Post(url, "application/json", bytes.NewReader(bz))
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

	if isDetectMismatch && strings.Contains(stringBody, "sequence mismatch") {
		return ""
	}

	return string(body)
}

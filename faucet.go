package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tmlibs/bech32"
	"github.com/tomasen/realip"
)

var key string
var node string
var chain string
var pass string

var amountTable = map[string]int{
	"luna": 10,
	"krw":  10000,
	"usd":  10,
	"sdr":  10,
	"gbp":  10,
	"eur":  10,
	"jpy":  1000,
	"cny":  100,
}

var dailyLimitTable = map[string]int{
	"luna": 100,
	"krw":  100000,
	"usd":  100,
	"sdr":  100,
	"gbp":  10,
	"eur":  100,
	"jpy":  10000,
	"cny":  1000,
}

const (
	requestLimitSecs = 30
)

type claimStruct struct {
	Address  string
	Response string
	Denom    string
}
type Coin struct {
	Denom  string `json:"denom"`
	Amount int    `json:"amount"`
}

func main() {
	db, err := leveldb.OpenFile("db/ipdb", nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

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
		pass = "12345678"
	}

	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <reCaptcha private key>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	} else {
		recaptcha.Init(os.Args[1])

		http.Handle("/", http.FileServer(http.Dir("./frontend/dist/")))
		http.HandleFunc("/claim", createGetCoinsHandler(db))

		if err := http.ListenAndServe("127.0.0.1:3000", nil); err != nil {
			log.Fatal("failed to start server", err)
		}
	}
}

func executeCmd(command string, writes ...string) {
	cmd, wc, _ := goExecute(command)

	for _, write := range writes {
		_, _ = wc.Write([]byte(write + "\n"))
	}
	_ = cmd.Wait()
}

func goExecute(command string) (cmd *exec.Cmd, pipeIn io.WriteCloser, pipeOut io.ReadCloser) {
	cmd = getCmd(command)
	pipeIn, _ = cmd.StdinPipe()
	pipeOut, _ = cmd.StdoutPipe()
	go cmd.Start()
	time.Sleep(time.Second)
	return cmd, pipeIn, pipeOut
}

func getCmd(command string) *exec.Cmd {
	// split command into command and args
	split := strings.Split(command, " ")

	var cmd *exec.Cmd
	if len(split) == 1 {
		cmd = exec.Command(split[0])
	} else {
		cmd = exec.Command(split[0], split[1:]...)
	}

	return cmd
}

type RequestLog struct {
	Coins     []Coin    `json:"coin"`
	Requested time.Time `json:"updated"`
}

func (requestLog *RequestLog) dripCoin(denom string) error {
	amount := amountTable[denom]

	// try to update coin
	for idx, coin := range requestLog.Coins {
		if coin.Denom == denom {
			if (coin.Amount + amount) > dailyLimitTable[denom] {
				return errors.New("exceed denom limit")
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
			return errors.New("too fast")
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

func createGetCoinsHandler(db *leveldb.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var claim claimStruct

		defer func() {
			if err := recover(); err != nil {
				http.Error(w, err.(error).Error(), 400)
			}
		}()

		// decode JSON response from front end
		decoder := json.NewDecoder(request.Body)
		decoderErr := decoder.Decode(&claim)
		if decoderErr != nil {
			panic(decoderErr)
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
			return
		}

		// send the coins!
		if captchaPassed {
			amount := amountTable[claim.Denom]
			sendFaucet := fmt.Sprintf(
				"terracli tx send %v %v%v --from %v --chain-id %v --fees 2luna",
				encodedAddress, amount, claim.Denom, key, chain)
			fmt.Println(time.Now().UTC().Format(time.RFC3339), encodedAddress, "[1] ", amount, claim.Denom)
			executeCmd(sendFaucet, pass)

		} else {
			fmt.Println("Captcha Failed")
		}

		return
	}
}

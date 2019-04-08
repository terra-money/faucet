package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	recaptcha "github.com/dpapathanasiou/go-recaptcha"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tmlibs/bech32"
	"github.com/tomasen/realip"
)

var key string
var node string
var chain string
var pass string
var sequence int64

var amountTable = map[string]int{
	MicroLunaDenom: 10 * MicroUnit,
	MicroKRWDenom:  10000 * MicroUnit,
	MicroUSDDenom:  10 * MicroUnit,
	MicroSDRDenom:  10 * MicroUnit,
	MicroGBPDenom:  10 * MicroUnit,
	MicroEURDenom:  10 * MicroUnit,
	MicroJPYDenom:  1000 * MicroUnit,
	MicroCNYDenom:  100 * MicroUnit,
}

const (
	requestLimitSecs = 30

	keyVar      = "key"
	nodeVar     = "node"
	chainIDVar  = "chain-id"
	passwordVar = "pass"
)

// Claim wraps a faucet claim
type Claim struct {
	Address  string
	Response string
	Denom    string
}

// Coin is the same as sdk.Coin
type Coin struct {
	Denom  string `json:"denom"`
	Amount int    `json:"amount"`
}

// Env wraps env variables stored in env.json
type Env struct {
	Key   string `json:"key"`
	Node  string `json:"node"`
	Chain string `json:"chain-id"`
	Pass  string `json:"pass"`
}

func readEnvFile() {
	data, err := ioutil.ReadFile("./env.json")
	if err != nil {
		fmt.Print(err)
	}

	var env Env
	err = json.Unmarshal(data, &env)
	if err != nil {
		fmt.Println("error:", err)
	}

	os.Setenv(keyVar, env.Key)
	os.Setenv(nodeVar, env.Node)
	os.Setenv(chainIDVar, env.Chain)
	os.Setenv(passwordVar, env.Pass)
}

func main() {
	db, err := leveldb.OpenFile("db/ipdb", nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

  readEnvFile()
	key = os.Getenv(keyVar)
	if key == "" {
		key = "faucet"
	}

	node = os.Getenv(nodeVar)
	if node == "" {
		node = "http://localhost:46657"
	}

	chain = os.Getenv(chainIDVar)
	if chain == "" {
		chain = "soju-0006"
	}

	pass = os.Getenv(passwordVar)
	if pass == "" {
		pass = "12345678"
	}

	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <reCaptcha private key>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	} else {
		recaptcha.Init(os.Args[1])

		fmt.Println("chain:", chain)
		fmt.Println("node:", node)

		// Query current faucet sequence
		queryCommand := fmt.Sprintf(
			"terracli query account terra12c5s58hnc3c0pjr5x7u68upsgzg2r8fwq5nlsy --chain-id %v --node %v -o json",
			chain, node)
		_, _, out := goExecute(queryCommand)

		// Capture seqeunce string from json
		r := regexp.MustCompile(`"sequence":"(\d+)"`)
		groups := r.FindStringSubmatch(out)

		if len(groups) != 2 {
			fmt.Printf("cannot find sequence")
			os.Exit(1)
		}

		// Convert sequence string to int64
		sequence, _ = strconv.ParseInt(groups[1], 10, 64)

		http.Handle("/", http.FileServer(http.Dir("./frontend/build/")))
		http.HandleFunc("/claim", createGetCoinsHandler(db))

		if err := http.ListenAndServe(":3000", nil); err != nil {
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

func goExecute(command string) (cmd *exec.Cmd, pipeIn io.WriteCloser, output string) {
	cmd = getCmd(command)
	pipeIn, _ = cmd.StdinPipe()
	// pipeOut, _ = cmd.StdoutPipe()

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	go cmd.Start()
	time.Sleep(time.Second)
	return cmd, pipeIn, stdBuffer.String()
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
			if (coin.Amount + amount) > amountTable[denom]*10 {
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
		var claim Claim

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
		}

		// send the coins!
		if captchaPassed {
			amount := amountTable[claim.Denom]
			sendFaucet := fmt.Sprintf(
				"terracli tx send %v %v%v --from %v --chain-id %v --fees 10mluna --node %v --async --sequence %v",
				encodedAddress, amount, claim.Denom, key, chain, node, sequence)
			fmt.Println(time.Now().UTC().Format(time.RFC3339), encodedAddress, "[1] ", amount, claim.Denom)
			sequence = sequence + 1
			executeCmd(sendFaucet, pass)

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "{\"amount\": %v}", amount)
		} else {
			fmt.Println("Captcha Failed")
		}

		return
	}
}

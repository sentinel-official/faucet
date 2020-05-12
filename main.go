package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	hub "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/faucet/types"
	"github.com/sentinel-official/faucet/utils"
)

var (
	chainID    string
	rpcAddress string
	from       string
	password   string
	coins      string

	secret = os.Getenv("RECAPTCHA_SECRET")

	_cli *types.CLI
)

func init() {
	flag.StringVar(&chainID, "chain-id", "sentinel-turing-2", "chain id")
	flag.StringVar(&rpcAddress, "rpc-address", "127.0.0.1:26657", "rpc server address")
	flag.StringVar(&from, "from", "faucet", "from account name")
	flag.StringVar(&password, "password", "", "from account password")
	flag.StringVar(&coins, "coins", "100000000tsent", "coins to transfer")
	flag.Parse()
}

type transferRequest struct {
	Address           string `json:"address"`
	ReCaptchaResponse string `json:"re_captcha_response"`
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	var body transferRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErrorToResponse(w, 400, &types.Error{
			Message: "failed to unmarshal the body",
			Info:    err.Error(),
		})
		return
	}

	if body.Address == "" {
		utils.WriteErrorToResponse(w, 400, &types.Error{
			Message: "address field is empty",
		})
		return
	}

	if body.ReCaptchaResponse == "" {
		utils.WriteErrorToResponse(w, 400, &types.Error{
			Message: "re_captcha_response field is empty",
		})
		return
	}

	if err := utils.ReCaptchaVerify(secret, body.ReCaptchaResponse, ""); err != nil {
		utils.WriteErrorToResponse(w, 400, &types.Error{
			Message: "failed to verify",
			Info:    err.Error(),
		})
		return
	}

	txRes, err := _cli.Transfer(body.Address, coins)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.Error{
			Message: "failed to transfer the coins",
			Info:    err.Error(),
		})
		return
	}

	w.Header().Add("Content-Type", "application/json")
	utils.WriteResultToResponse(w, 200, txRes)
}

func main() {
	var err error

	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(hub.Bech32PrefixAccAddr, hub.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(hub.Bech32PrefixValAddr, hub.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(hub.Bech32PrefixConsAddr, hub.Bech32PrefixConsPub)
	cfg.Seal()

	_cli, err = types.NewCLI(chainID, rpcAddress, from, password)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	router.Name("Transfer").
		Methods("POST").
		Path("/transfer").
		HandlerFunc(transferHandler)

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowedMethods([]string{"POST", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))
	log.Fatal(http.ListenAndServe(":8000", cors(router)))
}

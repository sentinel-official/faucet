package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hub "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/faucet/types"
	"github.com/sentinel-official/faucet/utils"
)

var (
	_cli   *types.CLI
	coins  string
	secret = os.Getenv("RECAPTCHA_SECRET")
)

func init() {
	var (
		chainID    string
		rpcAddress string
		from       string
		password   string
		err        error
	)

	flag.StringVar(&chainID, "chain-id", "sentinel-turing-1", "chain id")
	flag.StringVar(&rpcAddress, "rpc-address", "127.0.0.1:26657", "rpc server address")
	flag.StringVar(&from, "from", "faucet", "from account name")
	flag.StringVar(&password, "password", "", "from account password")
	flag.StringVar(&coins, "coins", "100000000tsent", "coins to transfer")
	flag.Parse()

	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(hub.Bech32PrefixAccAddr, hub.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(hub.Bech32PrefixValAddr, hub.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(hub.Bech32PrefixConsAddr, hub.Bech32PrefixConsPub)
	cfg.Seal()

	_cli, err = types.NewCLI(chainID, rpcAddress, from, password)
	if err != nil {
		panic(err)
	}
}

func transferCoinsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.FormValue("address")
	if address == "" {
		utils.WriteErrorToResponse(w, 400, &types.Error{
			Message: "address field is empty",
		})
		return
	}

	rcr := r.FormValue("g-recaptcha-response")
	if rcr == "" {
		utils.WriteErrorToResponse(w, 400, &types.Error{
			Message: "g-recaptcha-response field is empty",
		})
		return
	}

	if err := utils.ReCaptchaVerify(secret, rcr, ""); err != nil {
		utils.WriteErrorToResponse(w, 400, &types.Error{
			Message: "failed to verify",
			Info:    err.Error(),
		})
		return
	}

	txRes, err := _cli.TransferCoins(address, coins)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.Error{
			Message: "failed to transfer coins",
			Info:    err.Error(),
		})
		return
	}

	utils.WriteResultToResponse(w, 200, txRes)
}

func main() {
	http.HandleFunc("/transfer", transferCoinsHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

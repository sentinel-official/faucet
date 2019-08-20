package types

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	txBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/sentinel-official/hub/app"
	tmLog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	tm "github.com/tendermint/tendermint/types"
)

type CLI struct {
	ctx      context.CLIContext
	txb      txBuilder.TxBuilder
	password string
	mutex    sync.Mutex
}

func NewCLI(chainID, rpcAddress, from, password string) (*CLI, error) {
	cdc := app.MakeCodec()
	tm.RegisterEventDatas(cdc)

	kb, err := keys.NewKeyBaseFromDir(app.DefaultCLIHome)
	if err != nil {
		log.Println("failed to initialize the key base", err)
		return nil, err
	}

	keyInfo, err := kb.Get(from)
	if err != nil {
		log.Println("failed to get the key info from key base", err)
		return nil, err
	}

	_client := client.NewHTTP(rpcAddress, "/websocket")
	verifier, err := proxy.NewVerifier(
		chainID, filepath.Join(app.DefaultNodeHome, "faucet_lite"), _client, tmLog.NewNopLogger(), 10)

	if err != nil {
		log.Println("failed to initialize the verifier", err)
		return nil, err
	}

	ctx := context.CLIContext{
		Codec:         cdc,
		Client:        _client,
		Keybase:       kb,
		Output:        os.Stdout,
		OutputFormat:  "text",
		NodeURI:       rpcAddress,
		From:          keyInfo.GetName(),
		AccountStore:  auth.StoreKey,
		BroadcastMode: "sync",
		Verifier:      verifier,
		VerifierHome:  app.DefaultNodeHome,
		FromAddress:   keyInfo.GetAddress(),
		FromName:      keyInfo.GetName(),
		SkipConfirm:   true,
	}.WithAccountDecoder(cdc)

	account, err := ctx.GetAccount(keyInfo.GetAddress().Bytes())
	if err != nil {
		log.Println("failed to get account", err)
		return nil, err
	}

	txb := txBuilder.NewTxBuilder(utils.GetTxEncoder(cdc),
		account.GetAccountNumber(), account.GetSequence(), 1000000000,
		1.0, false, chainID,
		"", sdk.Coins{}, sdk.DecCoins{}).
		WithKeybase(kb)

	return &CLI{
		ctx:      ctx,
		txb:      txb,
		password: password,
	}, nil
}

func (c *CLI) completeAndBroadcastTxSync(messages []sdk.Msg) (*sdk.TxResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	txBytes, err := c.txb.BuildAndSign(c.ctx.FromName, c.password, messages)
	if err != nil {
		log.Println("failed to build and sign messages", err)
		return nil, err
	}

	node, err := c.ctx.GetNode()
	if err != nil {
		log.Println("failed to get node", err)
		return nil, err
	}

	res, err := node.BroadcastTxSync(txBytes)
	if res != nil && res.Code == 0 {
		c.txb = c.txb.WithSequence(c.txb.Sequence() + 1)
	}

	txRes := sdk.NewResponseFormatBroadcastTx(res)
	if txRes.Code != 0 {
		return &txRes, fmt.Errorf(txRes.String())
	}

	return &txRes, err
}

func (c *CLI) Transfer(to, coins string) (*sdk.TxResponse, error) {
	toAddress, err := sdk.AccAddressFromBech32(to)
	if err != nil {
		log.Println("failed to parse the address", err)
		return nil, err
	}

	amount, err := sdk.ParseCoins(coins)
	if err != nil {
		log.Println("failed to parse the coins", err)
		return nil, err
	}

	messages := []sdk.Msg{
		bank.NewMsgSend(c.ctx.FromAddress, toAddress, amount),
	}

	return c.completeAndBroadcastTxSync(messages)
}

module github.com/sentinel-official/faucet

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.35.0
	github.com/sentinel-official/hub v0.1.0
	github.com/tendermint/tendermint v0.31.5
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5

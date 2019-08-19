package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sentinel-official/faucet/types"
)

func WriteErrorToResponse(w http.ResponseWriter, code int, _error interface{}) {
	res := types.Response{
		Success: false,
		Error:   _error,
	}

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

func WriteResultToResponse(w http.ResponseWriter, code int, result interface{}) {
	res := types.Response{
		Success: true,
		Result:  result,
	}

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

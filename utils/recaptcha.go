package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/sentinel-official/faucet/types"
)

const (
	_url = "https://www.google.com/recaptcha/api/siteverify"
)

func ReCaptchaVerify(secret, rcr, rip string) error {
	values := url.Values{
		"secret":   {secret},
		"response": {rcr},
		"remoteip": {rip},
	}

	r, err := http.PostForm(_url, values)
	if err != nil {
		log.Println("failed to post form", err)
		return err
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			panic(err)
		}
	}()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read the body", err)
		return err
	}

	var resp types.ReCaptchaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Println("failed to unmarshal the body", err)
		return err
	}

	if !resp.Success {
		return fmt.Errorf(strings.Join(resp.ErrorCodes, ", "))
	}

	return nil
}

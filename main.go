package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/getlantern/systray"
)

type bitsoResponse struct {
	Success bool            `json:success`
	Payload currencyPayload `json:payload`
}

type currencyPayload struct {
	Success bool   `json:success`
	High    string `json:high`
	Last    string `json:last`
	Created string `json:created_at`
	Book    string `json:book`
	Volume  string `json:volume`
	Vwap    string `json:vwap`
	Low     string `json:low`
	Ask     string `json:ask`
	Bid     string `json:bid`
}

const bitsoAPI = "https://api.bitso.com/v3/ticker/?book="
const convertTo = "_mxn"
const pingTime = 10 // in seconds, consider Bitso limit is 300 request p/minute.

var currencies = []string{"btc", "eth"}
var currentCurrency = currencies[0]

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/" + currentCurrency + ".ico"))
	systray.SetTitle("Bitso!")
	systray.SetTooltip("Loading data.")

	btcItem := systray.AddMenuItem("Bitcoin", "")
	ethItem := systray.AddMenuItem("Ethereum", "")

	for _, c := range currencies {
		go func(c string) {
			for {
				res := fetchBitsoData(c)
				body, _ := ioutil.ReadAll(res.Body)
				bitsoResponse := new(bitsoResponse)
				err := json.Unmarshal(body, &bitsoResponse)
				if err != nil && !bitsoResponse.Success {
					fmt.Println(err)
				}
				payload := bitsoResponse.Payload
				fmt.Println(payload)

				time.Sleep(pingTime * time.Second)
			}
		}(c)
	}

	go func() {
		for {
			select {
			case <-btcItem.ClickedCh:
				setDefaultCurrency("btc")
			case <-ethItem.ClickedCh:
				setDefaultCurrency("eth")
			}
		}
	}()
}

func updateCurrency(c string) {

}

func setDefaultCurrency(c string) {
	systray.SetIcon(getIcon("assets/" + c + ".ico"))

}

func onExit() {}

func fetchBitsoData(c string) *http.Response {
	res, err := http.Get(bitsoAPI + c + convertTo)
	if err != nil {
		panic(err)
	}
	return res
}

func updateSystray(t string, tt string) {

}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

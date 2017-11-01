package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/getlantern/systray"
)

type bitsoResponse struct {
	Success bool            `json:success`
	Payload currencyPayload `json:payload`
}

type currencyPayload struct {
	Success      bool   `json:success`
	High         string `json:high`
	Last         string `json:last`
	Created      string `json:created_at`
	Book         string `json:book`
	Volume       string `json:volume`
	Vwap         string `json:vwap`
	Low          string `json:low`
	Ask          string `json:ask`
	Bid          string `json:bid`
	DisplayValue string
}

const bitsoAPI = "https://api.bitso.com/v3/ticker/?book="
const convertTo = "_mxn"
const pingTime = 10 // in seconds, consider Bitso limit is 300 request p/minute and we make len(currencies) queries per time

var currencies = []string{"btc", "eth"}
var currentCurrency = currencies[0]
var storedValues = map[string]currencyPayload{}

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
				payload.DisplayValue = humanReadable(payload.Bid)
				storedValues[c] = payload

				switch c {
				case "btc":
					btcItem.SetTitle("Btc: $" + payload.DisplayValue)
					btcItem.SetTooltip("Updated on " + payload.Created)
				case "eth":
					ethItem.SetTitle("Eth: $" + payload.DisplayValue)
					ethItem.SetTooltip("Updated on " + payload.Created)
				}

				updateSystray()

				time.Sleep(pingTime * time.Second)
			}
		}(c)
	}

	// Listeners for menu items
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

func updateCurrency(c string, payload currencyPayload) {

}

func setDefaultCurrency(c string) {
	systray.SetIcon(getIcon("assets/" + c + ".ico"))
	currentCurrency = c
	updateSystray()
}

func onExit() {}

func fetchBitsoData(c string) *http.Response {
	res, err := http.Get(bitsoAPI + c + convertTo)
	if err != nil {
		panic(err)
	}
	return res
}

func humanReadable(s string) string {
	i, _ := strconv.ParseFloat(s, 64)
	return string(humanize.Commaf(i))
}

func updateSystray() {
	systray.SetTitle("$" + storedValues[currentCurrency].DisplayValue)
	systray.SetTooltip("Updated on " + storedValues[currentCurrency].Created)
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

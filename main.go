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
	Created      string `json:"created_at"`
	Book         string `json:book`
	Volume       string `json:volume`
	Vwap         string `json:vwap`
	Low          string `json:low`
	Ask          string `json:ask`
	Bid          string `json:bid`
	DisplayValue string
	UpdatedOn    string
}

const bitsoAPI = "https://api.bitso.com/v3/ticker/?book="
const convertTo = "_mxn"
const pingTime = 10 // in seconds, consider Bitso limit is 300 request p/minute and we make len(currencies) queries per time

var currencies = []string{"btc", "eth", "xrp"}
var currentCurrency = currencies[0]
var storedValues = map[string]currencyPayload{}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	sysicon, err := getIcon("assets/" + currentCurrency + ".ico")
	if err == nil {
		systray.SetIcon(sysicon)
	}
	systray.SetTitle("Bitso!")
	systray.SetTooltip("Loading data.")

	btcItem := systray.AddMenuItem("Bitcoin", "")
	ethItem := systray.AddMenuItem("Ethereum", "")
	xrpItem := systray.AddMenuItem("Ripple", "")

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
				payload.DisplayValue = humanizeCurrency(payload.Last)
				payload.UpdatedOn = humanizeDate(payload.Created)
				storedValues[c] = payload

				switch c {
				case "btc":
					btcItem.SetTitle("Btc: $" + payload.DisplayValue)
					btcItem.SetTooltip("Updated on " + payload.UpdatedOn)
				case "eth":
					ethItem.SetTitle("Eth: $" + payload.DisplayValue)
					ethItem.SetTooltip("Updated on " + payload.UpdatedOn)
				case "xrp":
					xrpItem.SetTitle("Xrp: $" + payload.DisplayValue)
					xrpItem.SetTooltip("Updated on " + payload.UpdatedOn)
				}

				updateSystray()

				time.Sleep(pingTime * time.Second)
			}
		}(c)
		time.Sleep(100 * time.Millisecond) // avoid concurrent map writes, need to refactor this way
	}

	// Listeners for menu items
	go func() {
		for {
			select {
			case <-btcItem.ClickedCh:
				setDefaultCurrency("btc")
			case <-ethItem.ClickedCh:
				setDefaultCurrency("eth")
			case <-xrpItem.ClickedCh:
				setDefaultCurrency("xrp")
			}
		}
	}()
}

func setDefaultCurrency(c string) {
	sysicon, err := getIcon("assets/" + currentCurrency + ".ico")
	if err == nil {
		systray.SetIcon(sysicon)
	}
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

func humanizeCurrency(s string) string {
	i, _ := strconv.ParseFloat(s, 64)
	return string(humanize.Commaf(i))
}

func humanizeDate(s string) string {
	t, _ := time.Parse("2006-01-02T15:04:05+00:00", s)
	return t.Format("15:04:05")
}

func updateSystray() {
	systray.SetTitle("$" + storedValues[currentCurrency].DisplayValue)
	systray.SetTooltip("Updated on " + storedValues[currentCurrency].UpdatedOn)
}

func getIcon(s string) ([]byte, error) {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b, err
}

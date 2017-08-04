package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	listenPort   = ":80"
	ratesFile    = "./rates.json"
	ratesURLBase = "https://openexchangerates.org/api/latest.json?base=USD&app_id="
)

var (
	rateStore      *RateStore
	rateStoreMutex *sync.Mutex
	refreshMutex   *sync.Mutex
	ratesURL       string

	ticker  *time.Ticker
	trigger chan bool
	quit    chan struct{}
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing argument: App ID")
		return
	}

	ratesURL = ratesURLBase + strings.TrimSpace(os.Args[1])

	rateStoreMutex = &sync.Mutex{}
	refreshMutex = &sync.Mutex{}

	// request rates once every hour (less than 1000 times per month)
	ticker = time.NewTicker(1 * time.Hour)
	quit = make(chan struct{})
	trigger = make(chan bool)
	go refresh()
	trigger <- true

	http.HandleFunc("/convert", convertHandler)
	err := http.ListenAndServe(listenPort, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	close(quit)
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	from := r.Form.Get("from")
	to := r.Form.Get("to")
	v := r.Form.Get("v")
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		fmt.Fprintf(w, `{"error":"can't parse value"}`)
		return
	}

	res, err := convert(f, from, to)
	if err != nil {
		fmt.Fprintf(w, `{"error":"incorrect parameter(s)"}`)
		return
	}

	str := strconv.FormatFloat(res, 'f', 2, 64)

	fmt.Fprintf(w, `{"result":`+str+`}`)
}

// RateStore is used to keep most recent rates in memory
type RateStore struct {
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}

func refresh() {
	for {
		select {
		case <-ticker.C:
			refreshCurrencyRates()
		case <-trigger:
			refreshCurrencyRates()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func convert(amount float64, base, to string) (float64, error) {
	rateStoreMutex.Lock()
	defer rateStoreMutex.Unlock()

	if rateStore == nil {
		return 0, errors.New("currency rates required")
	}

	baseValue, exists := rateStore.Rates[base]
	if !exists {
		return 0, errors.New("base value can't be found")
	}

	toValue, exists := rateStore.Rates[to]
	if !exists {
		return 0, errors.New("base value can't be found")
	}

	converted := amount * toValue / baseValue

	return converted, nil
}

func refreshCurrencyRates() {
	refreshMutex.Lock()
	defer refreshMutex.Unlock()

	rateBytes, err := ioutil.ReadFile(ratesFile)
	if err != nil {
		// file can't be found or read
		// request it
		err = requestRates()
		if err != nil {
			fmt.Println("can't get rates:", err)
			return
		}
		return
	}

	rateStoreMutex.Lock()
	err = json.Unmarshal(rateBytes, &rateStore)
	if err != nil {
		rateStoreMutex.Unlock()
		// file can't be read
		// request it
		err = requestRates()
		if err != nil {
			fmt.Println("can't get rates:", err)
			return
		}
		return
	}

	t := time.Unix(rateStore.Timestamp, 0)
	rateStoreMutex.Unlock()
	now := time.Now()

	// refresh if one hour elapsed
	if now.Sub(t) > time.Hour {
		err = requestRates()
		if err != nil {
			fmt.Println("can't get rates:", err)
			return
		}
		return
	}

	// don't do anything if rates already up to date
}

func requestRates() error {
	resp, err := http.Get(ratesURL)
	if err != nil {
		return err
	}

	jsonBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	rateStoreMutex.Lock()
	err = json.Unmarshal(jsonBytes, &rateStore)
	if err != nil {
		rateStoreMutex.Unlock()
		return err
	}

	rateStore.Timestamp = time.Now().Unix()
	jsonBytes, err = json.Marshal(&rateStore)
	rateStoreMutex.Unlock()

	err = ioutil.WriteFile(ratesFile, jsonBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

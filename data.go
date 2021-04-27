package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	source = "https://www.vattenfall.se/api/price/spot/pricearea/%s/%s/%s"
)

type Data struct {
	Timestamp time.Time
	Region    string
	Value     float64
	Currency  string
}

var (
	info      = map[string][]Data{}
	lastFetch = map[string]time.Time{}
	lock      = sync.RWMutex{}
)

func (d *Data) UnmarshalJSON(data []byte) error {
	type internal struct {
		Timestamp string  `json:"TimeStamp"`
		Value     float64 `json:"Value"`
		PriceArea string  `json:"PriceArea"`
	}

	var v internal
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	parsed, err := time.Parse("2006-01-02T15:04:05", v.Timestamp)
	if err != nil {
		return err
	}

	d.Timestamp = parsed
	d.Region = v.PriceArea
	d.Value = (v.Value * 100) / 10000
	d.Currency = "SEK"

	return nil
}

func fetch(date time.Time, region string) ([]Data, error) {
	lock.RLock()
	lf := lastFetch[region]
	lock.RUnlock()

	if lf.Add(30 * time.Minute).Before(date) {
		b, err := fetchFromURL(date, region)
		if err != nil {
			return nil, err
		}

		data := []Data{}
		err = json.Unmarshal(b, &data)
		if err != nil {
			return nil, err
		}

		lock.Lock()
		lastFetch[region] = date
		info[region] = data
		lock.Unlock()

		return data, nil
	}

	return info[region], nil
}

func fetchFromURL(date time.Time, region string) ([]byte, error) {
	res := fmt.Sprintf(
		source,
		date.Format("2006-01-02"),
		date.Format("2006-01-02"),
		region,
	)
	resp, err := http.Get(res)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", res, err)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected 200 OK when fetching: %s, got: %d", res, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

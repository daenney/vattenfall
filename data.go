package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

const (
	source    = "https://www.vattenfall.se/api/price/spot/pricearea/%s/%s/%s"
	userAgent = "vattenfall-exporter/%s (+https://github.com/daenney/vattenfall)"
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
	d.Value = math.Round(v.Value*100) / 10000
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
		date.AddDate(0, 0, 1).Format("2006-01-02"),
		region,
	)

	req, err := http.NewRequest("GET", res, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf(userAgent, version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", res, err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
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

// forecastHandler returns JSON in a format compatible with the Grafana JSON-API
// plugin. This allows you to display the forecast as Prometheus doesn't do values
// in the future
func forecastHandler(loc *time.Location, regions []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type point struct {
			Time   time.Time `json:"time"`
			Region string    `json:"region"`
			Value  float64   `json:"value"`
		}
		res := []point{}

		for _, reg := range regions {
			now := time.Now().In(loc)
			data, err := fetch(now, reg)
			if err != nil {
				log.Println(err)
			}
			for _, entry := range data {
				if entry.Timestamp.Before(now) {
					continue
				}
				res = append(res, point{Time: entry.Timestamp, Region: entry.Region, Value: entry.Value})
			}
		}
		enc := json.NewEncoder(w)
		err := enc.Encode(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

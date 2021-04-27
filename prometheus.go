package main

import (
	"io"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type vattenfallCollector struct {
	prices  *prometheus.Desc
	regions []string
	tz      *time.Location
}

func NewVattenfallCollector(regions []string, tz *time.Location) *vattenfallCollector {
	return &vattenfallCollector{
		prices: prometheus.NewDesc(
			"energy_price_per_kwh",
			"Energy price per kWh for a region",
			[]string{"region"}, prometheus.Labels{"currency": "SEK", "country": "SE"},
		),
		regions: regions,
		tz:      tz,
	}
}

func (v *vattenfallCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- v.prices
}

func (v *vattenfallCollector) Collect(ch chan<- prometheus.Metric) {
	for _, r := range v.regions {
		now := time.Now().In(v.tz)
		data, err := fetch(now, r)
		if err != nil {
			log.Println(err)
		}
		for _, entry := range data {
			if entry.Timestamp.Hour() == now.Hour() {
				ch <- prometheus.MustNewConstMetric(v.prices,
					prometheus.GaugeValue, entry.Value, entry.Region)
			}
		}
	}
}

func WriteMetricsTo(w io.Writer, g prometheus.Gatherer) error {
	mfs, err := g.Gather()
	if err != nil {
		return err
	}

	for _, mf := range mfs {
		if _, err := expfmt.MetricFamilyToText(w, mf); err != nil {
			return err
		}
	}

	return nil
}

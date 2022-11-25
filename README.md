# Vattenfall

This is a very basic Prometheus exporter for the Vattenfall electricity spot
prices in Sweden. You can explore the data on https://www.vattenfall.se/elavtal/elpriser/timpris-pa-elborsen/.

It'll export one metric, `energy_price_per_kwh` for each region:

```
# HELP energy_price_per_kwh Energy price per kWh for a region
# TYPE energy_price_per_kwh gauge
energy_price_per_kwh{country="SE",currency="SEK",region="SN1"} 0.4707
energy_price_per_kwh{country="SE",currency="SEK",region="SN2"} 0.4707
energy_price_per_kwh{country="SE",currency="SEK",region="SN3"} 0.4707
energy_price_per_kwh{country="SE",currency="SEK",region="SN4"} 0.4707
```

Data is cached for 30min in memory to not hammer Vattenfall each time you
scrape the collector (and their API is slow). Though energy prices are
fixed the day before, due to exchange rate fluctuations prices sometimes
update during the day.

## Usage

```
-output.file string
    write metrics to specified file (must have .prom extension)
-output.http string
    host:port to listen on for HTTP scrapes
-region value
    region to query for, SN1-4, can be passed multiple times
```

To run it as a Prometheus exporter that you can query over HTTP:

```sh
$ vattenfall -output.http=":9000" -region SN1 -region SN2 -region SN3 -region SN4
```

Please note that there's 2 endpoints `/metrics` which instruments the
collector itself, and `/prices` with the pricing info.

If you want to use it with the textfile collector, for example in an hourly cron:

```sh
$ vattenfall -output.file="/etc/prometheus/textfile/electricity.prom" -region SN1 -region SN2 -region SN3 -region SN4
```

Or to just get the values on the console:

```sh
$ vattenfall -region SN1 -region SN2 -region SN3 -region SN4
```

## Forecast

At `/forecast` the prices for up to the end of the next day are displayed. The
JSON on that endpoint can be used with the [Grafana JSON API datasource](https://grafana.com/grafana/plugins/marcusolsson-json-datasource/)
to display the forecast without having to ingest this data in a separate
database.

Add a JSON API datasource, name it whatever you want and point it at:
`http(s):url-of-vattenfall/forecast`. Then, configure a time series panel with
the new data source and:

* Options:
  * interval: 1h
  * relative time: +36h
* Query:
  * Fields:
    * `$.[*].time`, type: Time
    * `$.[*].value`, type: Number
    * `$.[*].region`, type: Auto (or String)
  * Experimental:
    * Group by: `region`
    * Metric: `value`

For the graph itself you want:
* Line interpolation: step after
* Unit: Swedish Krona (kr)

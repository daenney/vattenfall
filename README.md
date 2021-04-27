# Vattenfall

This is a very basic Prometheus exporter for the Vattenfall electricity spot
prices in Sweden. It'll export one metric, `energy_price_per_kwh` for each
region:

```
# HELP energy_price_per_kwh Energy price per kWh for a region
# TYPE energy_price_per_kwh gauge
energy_price_per_kwh{country="SE",currency="SEK",region="SN1"} 0.4707
energy_price_per_kwh{country="SE",currency="SEK",region="SN2"} 0.4707
energy_price_per_kwh{country="SE",currency="SEK",region="SN3"} 0.4707
energy_price_per_kwh{country="SE",currency="SEK",region="SN4"} 0.4707
```

Data is cached for 30min in memory to not hammer Vattenfall each time you
scrape the collector (and their API is slow), but still catch any price
adjustments that might occur.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type regionFlag []string

func (r *regionFlag) String() string {
	return strings.Join(*r, ", ")
}

func (r *regionFlag) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func main() {
	var regions regionFlag
	flag.Var(&regions, "region", "region to query for, SN1-4, can be passed multiple times")
	txtfile := flag.String("output.file", "", "write metrics to specified file (must have .prom extension)")
	addr := flag.String("output.http", "", "host:port to listen on for HTTP scrapes")
	showVersion := flag.Bool("version", false, "show version and build info")

	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "{\"version\": \"%s\", \"commit\": \"%s\", \"date\": \"%s\"}\n", version, commit, date)
		os.Exit(0)
	}

	loc, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		log.Fatalln(err)
	}

	c := prometheus.NewPedanticRegistry()
	c.MustRegister(NewVattenfallCollector(regions, loc))

	if *txtfile == "" && *addr == "" {
		WriteMetricsTo(os.Stdout, c)
		os.Exit(0)
	}

	if *txtfile != "" {
		if elems := strings.Split(*txtfile, "."); elems[len(elems)-1] != "prom" {
			log.Fatalln("filename must end with .prom extension:", *txtfile)
		}
		err := prometheus.WriteToTextfile(*txtfile, c)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}

	if *addr != "" {
		stp := make(chan os.Signal, 1)
		signal.Notify(stp, os.Interrupt, syscall.SIGTERM)
		ctx, ctxCancel := context.WithCancel(context.Background())
		defer ctxCancel()

		p := prometheus.NewPedanticRegistry()
		p.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		p.MustRegister(prometheus.NewGoCollector())

		listener, err := net.Listen("tcp", *addr)
		if err != nil {
			log.Fatalln(err)
		}

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(p, promhttp.HandlerOpts{}))
		mux.Handle("/prices", promhttp.InstrumentMetricHandler(p, promhttp.HandlerFor(c, promhttp.HandlerOpts{})))
		h := &http.Server{
			Handler: mux,
		}
		go func() {
			if err := h.Serve(listener); err != http.ErrServerClosed {
				log.Fatalln(err.Error())
			}
		}()

		log.Printf("exporter listening on: %s", listener.Addr().String())
		<-stp
		h.Shutdown(ctx)
	}
}

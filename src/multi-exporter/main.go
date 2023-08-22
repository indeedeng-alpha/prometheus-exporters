package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

func init() {
	prometheus.MustRegister(version.NewCollector("multi_exporter"))
}

func main() {
	addr := flag.String("listen-address", ":9998", "The address to listen on for HTTP requests.")
	flag.Parse()

	log.Infoln("Starting multi-exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/dns", DnsHandler)
	http.HandleFunc("/dns_normalized", DnsNormalizedHandler)
	http.HandleFunc("/icmp", IcmpHandler)
	log.Infof("Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

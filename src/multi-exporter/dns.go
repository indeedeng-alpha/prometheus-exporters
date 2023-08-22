package main

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func Query(domain string, nameserver string) (rtt time.Duration, err error) {
	c := new(dns.Client)
	c.Timeout = time.Duration(3) * time.Second

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	m.RecursionDesired = true

	rec, rtt, err := c.Exchange(m, net.JoinHostPort(nameserver, "53"))
	if err != nil {
		return rtt, err
	} else {
		if rec.Rcode == dns.RcodeSuccess || rec.Rcode == dns.RcodeNameError {
			return rtt, nil
		} else {
			return rtt, fmt.Errorf("Error: Received %s for %s\n", dns.RcodeToString[rec.Rcode], domain)
		}
	}
}

func RandQuery(nameserver string) (rtt time.Duration, err error) {
	letters := "abcdefghijklmnopqrstuvwxyz"
	domain := ""
	for i := 0; i < 10; i++ {
		domain += string(letters[rand.Int()%len(letters)])
	}

	return Query(domain, nameserver)
}

func DnsHandler(w http.ResponseWriter, r *http.Request) {
	dnsFailure := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dns_failure",
		Help: "Is 1 when DNS request failed, 0 when request succeeded.",
	})
	dnsTimeout := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dns_timeout",
		Help: "Is 1 when failure was caused by a timeout, 0 otherwise.",
	})
	dnsRtt := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dns_rtt_seconds",
		Help: "Total DNS request round trip time.",
	})

	registry := prometheus.NewRegistry()
	registry.MustRegister(dnsFailure)
	registry.MustRegister(dnsTimeout)
	registry.MustRegister(dnsRtt)

	params := r.URL.Query()
	if nameservers, exists := params["nameserver"]; exists {
		if len(nameservers) > 1 {
			http.Error(w, fmt.Sprintf("too many \"nameserver\" parameters"), http.StatusBadRequest)
			return
		}
		nameserver := nameservers[0]
		if len(nameserver) == 0 {
			http.Error(w, fmt.Sprintf("empty \"nameserver\" parameter"), http.StatusBadRequest)
			return
		}

		rtt, err := RandQuery(nameserver)
		dnsFailure.Set(1)
		dnsRtt.Set(math.Inf(1))
		dnsTimeout.Set(0)
		if err != nil {
			log.Error(err)
			if err, ok := err.(net.Error); ok && err.Timeout() {
				dnsTimeout.Set(1)
			}
		} else {
			dnsFailure.Set(0)
			dnsRtt.Set(float64(rtt) / 1e9)
		}

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	} else {
		http.Error(w, "missing \"nameserver\" parameter", http.StatusBadRequest)
	}
}

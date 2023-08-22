package main

import (
	"fmt"
	"math"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

type medfilt struct {
	window  []float64
	current int
	medium  float64
}

func (m *medfilt) Insert(value float64) {
	initial := m.window[m.current]
	m.window[m.current] = value
	if initial == 0 {
		// Instead of looping over all the values, we use Welford's method to simultaneously compute the mean
		// and the variance. Idea and equations from here:
		// http://jonisalonen.com/2013/deriving-welfords-method-for-computing-variance/
		oldavg := m.medium
		m.medium += (value - oldavg) / float64(m.current+1)
	} else {
		// Calculate and add the marginal difference in the window
		// average to the window average. As before, this turns the O(n)
		// median calculation into an O(1) calculation--this means we
		// can make the window as large as we want without any
		// performance degradation. Idea and equation from here:
		// http://jonisalonen.com/2014/efficient-and-accurate-rolling-standard-deviation/
		oldavg := m.medium
		newavg := oldavg + (value-initial)/float64(len(m.window))
		m.medium = newavg
	}
	m.current += 1
	m.current %= len(m.window)
}

var servers = struct {
	sync.RWMutex
	filters map[string]*medfilt
}{filters: make(map[string]*medfilt)}

func newMedfilt(length int) *medfilt {
	m := new(medfilt)
	m.window = make([]float64, length)
	return m
}

func addNewRtt(nameserver string, rtt_sample float64) {
	servers.Lock()
	defer servers.Unlock()

	if _, exists := servers.filters[nameserver]; !exists {
		servers.filters[nameserver] = newMedfilt(100)
	}

	servers.filters[nameserver].Insert(rtt_sample)
}

func getMediumRtt(nameserver string) float64 {
	servers.RLock()
	defer servers.RUnlock()

	if _, exists := servers.filters[nameserver]; exists {
		return servers.filters[nameserver].medium
	} else {
		return 0
	}
}

func DnsNormalizedHandler(w http.ResponseWriter, r *http.Request) {
	dnsNormalizedFailure := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dns_normalized_failure",
		Help: "Is 1 when DNS request failed, 0 when request succeeded.",
	})
	dnsNormalizedRtt := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dns_normalized_rtt_seconds",
		Help: "Total DNS request round trip time, minus filtered ping RTTs.",
	})

	registry := prometheus.NewRegistry()
	registry.MustRegister(dnsNormalizedFailure)
	registry.MustRegister(dnsNormalizedRtt)

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

		// Ping first
		icmp_rtt, err := Ping(nameserver)
		icmp_failure := 1
		if err != nil {
			log.Error(err)
		} else {
			icmp_failure = 0
		}

		// Now query DNS
		dns_rtt, err := RandQuery(nameserver)
		dns_failure := 1
		if err != nil {
			log.Error(err)
		} else {
			dns_failure = 0
		}
		dnsNormalizedFailure.Set(float64(dns_failure))

		// Format output
		dnsNormalizedRtt.Set(math.Inf(1))
		if icmp_failure == 0 {
			addNewRtt(nameserver, float64(icmp_rtt)/1e9)
		}
		if dns_failure == 0 {
			icmp_medium_rtt := getMediumRtt(nameserver)
			dnsNormalizedRtt.Set(float64(dns_rtt)/1e9 - icmp_medium_rtt)
		}

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	} else {
		http.Error(w, "missing \"nameserver\" parameter", http.StatusBadRequest)
	}
}

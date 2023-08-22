package main

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/aristanetworks/goarista/monotime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func Ping(target string) (rtt time.Duration, err error) {
	timeout := time.Duration(3) * time.Second

	// Create socket to listen for the ICMP response
	socket, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return time.Duration(0), err
	}
	defer socket.Close()

	ip, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		return time.Duration(0), err
	}

	id, err := GetRand(0xffff + 1)
	if err != nil {
		return time.Duration(0), err
	}
	seq, err := GetRand(0xffff + 1)
	if err != nil {
		return time.Duration(0), err
	}

	// Build the ICMP message
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: int(id & 0xffff), Seq: int(seq & 0xffff),
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		return time.Duration(0), err
	}

	// Build the expected response for comparison
	m.Type = ipv4.ICMPTypeEchoReply
	br, err := m.Marshal(nil)
	if err != nil {
		return time.Duration(0), err
	}

	// Allocate packet buffer
	p := make([]byte, 1500)

	// Start the timeout and monotonic timers and send the ping
	start_time := time.Now()
	s := monotime.Now()
	l, err := socket.WriteTo(b, ip)
	if err != nil {
		return time.Duration(0), err
	}
	if l != len(b) {
		log.Warnf("Sent ping message is shorter than expected. Expected %d, got %d.", len(b), l)
	}

	if err := socket.SetReadDeadline(start_time.Add(timeout)); err != nil {
		return time.Duration(0), err
	}

	for {
		n, peer, err := socket.ReadFrom(p)
		r := monotime.Since(s)
		if err != nil {
			return time.Duration(0), err
		}

		// Is this the droid we're looking for?
		if peer.String() == ip.String() && bytes.Compare(p[:n], br) == 0 {
			return r, nil
		}
	}
}

func IcmpHandler(w http.ResponseWriter, r *http.Request) {
	icmpFailure := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "icmp_failure",
		Help: "Is 1 when ping failed, 0 when ping succeeded.",
	})
	icmpTimeout := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "icmp_timeout",
		Help: "Is 1 when failure was caused by a timeout, 0 otherwise.",
	})
	icmpRtt := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "icmp_rtt_seconds",
		Help: "Total ping round trip time.",
	})

	registry := prometheus.NewRegistry()
	registry.MustRegister(icmpFailure)
	registry.MustRegister(icmpTimeout)
	registry.MustRegister(icmpRtt)

	params := r.URL.Query()
	if targets, exists := params["target"]; exists {
		if len(targets) > 1 {
			http.Error(w, fmt.Sprintf("too many \"target\" parameters"), http.StatusBadRequest)
			return
		}
		target := targets[0]
		if len(target) == 0 {
			http.Error(w, fmt.Sprintf("empty \"target\" parameter"), http.StatusBadRequest)
			return
		}

		rtt, err := Ping(target)
		icmpFailure.Set(1)
		icmpRtt.Set(math.Inf(1))
		icmpTimeout.Set(0)
		if err != nil {
			log.Error(err)
			if err, ok := err.(net.Error); ok && err.Timeout() {
				icmpTimeout.Set(1)
			}
		} else {
			icmpFailure.Set(0)
			icmpRtt.Set(float64(rtt) / 1e9)
		}

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	} else {
		http.Error(w, "missing \"target\" parameter", http.StatusBadRequest)
	}
}

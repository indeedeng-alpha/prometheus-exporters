package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aristanetworks/goarista/monotime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	labels   = []string{"neighbor", "direction", "dn"}
	failures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "drsrepl_failures_total",
			Help: "Total number of DRS replication failures.",
		},
		labels,
	)
	lastSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "drsrepl_last_success_time",
			Help: "Time of last successful DRS replication, in unixtime.",
		},
		labels,
	)
	queryTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "drsrepl_query_duration_seconds",
		Help: "Total time it took to query and process the DRS replication information.",
	})
)

var (
	addr = flag.String("listen-address", ":9990", "The address to listen on for HTTP requests.")
	test = flag.String("test", "", "The file you'd like to test with, if not running on a Domain Controller.")
)

func extractMetrics(d string) error {
	l := prometheus.Labels{
		"neighbor":  "",
		"direction": "",
		"dn":        "",
	}

	c := float64(time.Now().Unix())

	section := "header"
	scanner := bufio.NewScanner(strings.NewReader(d))
	for scanner.Scan() {
		line := scanner.Text()
		switch line {
		case "==== INBOUND NEIGHBORS ====":
			section = "inbound"
		case "==== OUTBOUND NEIGHBORS ====":
			section = "outbound"
		case "==== KCC CONNECTION OBJECTS ====":
			section = "kcc_conn"
		default:
			if len(line) > 0 {
				if section == "inbound" || section == "outbound" {
					l["direction"] = section
					if !strings.HasPrefix(line, "\t") {
						l["dn"] = line
					} else if strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, "\t\t") {
						line = strings.TrimLeft(line, "\t")
						l["neighbor"] = strings.Split(line, " ")[0]
					} else if strings.HasPrefix(line, "\t\tLast attempt") {
					} else if strings.HasPrefix(line, "\t\t") && strings.Contains(line, "consecutive failure") {
						line = strings.TrimLeft(line, "\t")
						f, err := strconv.Atoi(strings.Split(line, " ")[0])
						if err != nil {
							return err
						}
						// FIXME: Only add if the current time of the last success before failure is
						// newer than the previous one. Otherwise, only add the difference between
						// the two.
						failures.With(l).Add(float64(f))
					} else if strings.HasPrefix(line, "\t\tLast success") {
						if strings.Contains(line, "NTTIME") {
							lastSuccess.With(l).Set(c)
						} else {
							const longForm = "Mon Jan  2 15:04:05 2006 MST"
							t, err := time.Parse(longForm, strings.Split(line, " @ ")[1])
							if err != nil {
								return err
							}
							lastSuccess.With(l).Set(float64(t.Unix()))
						}

						// Reset
						l = prometheus.Labels{
							"neighbor":  "",
							"direction": "",
							"dn":        "",
						}
					}
				}
			}
		}
	}
	return nil
}

func drsreplHandler(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(failures)
	registry.MustRegister(lastSuccess)
	registry.MustRegister(queryTime)

	// Get the current DRS replication status and time how long the command runs for.
	var start uint64
	var duration float64
	var out []byte
	var err error
	if len(*test) > 0 {
		start = monotime.Now()
		out, err = exec.Command("cat", *test).Output()
		duration = float64(monotime.Since(start).Seconds())
	} else {
		start = monotime.Now()
		out, err = exec.Command("sudo", "/usr/bin/samba-tool", "drs", "showrepl").Output()
		duration = float64(monotime.Since(start).Seconds())
	}
	if err != nil {
		log.Error(err)
		http.Error(w, fmt.Sprintf("Error: %v", err), 500)
		return
	}

	extractMetrics(string(out))
	queryTime.Set(duration)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func init() {
	prometheus.MustRegister(version.NewCollector("drsrepl_exporter"))
}

func main() {
	flag.Parse()

	log.Infoln("Starting DRS replication status exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/drsrepl", drsreplHandler)
	log.Infof("Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

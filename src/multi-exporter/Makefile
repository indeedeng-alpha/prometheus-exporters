all: multi-exporter

multi-exporter: main.go icmp.go dns.go dns_normalized.go
	go build \
		-ldflags "\
		-X github.com/prometheus/common/version.Version=r$(shell git rev-list --count HEAD) \
		-X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD) \
		-X github.com/prometheus/common/version.Revision=$(shell git rev-parse HEAD) \
		-X github.com/prometheus/common/version.BuildUser=$(shell whoami)@$(shell hostname) \
		-X github.com/prometheus/common/version.BuildDate=$(shell date -u "+%Y%m%d-%H:%M:%S") \
		"
	sudo setcap cap_net_raw+ep $@

deps:
	go get -u \
		golang.org/x/net/icmp \
		golang.org/x/net/ipv4 \
		github.com/aristanetworks/goarista/monotime \
		github.com/miekg/dns \
		github.com/prometheus/client_golang/prometheus \
		github.com/prometheus/client_golang/prometheus/promhttp \
		github.com/prometheus/common/log \
		github.com/prometheus/common/version


clean:
	rm -f multi-exporter

.PHONY: all clean deps

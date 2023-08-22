all: drs-replication-exporter

drs-replication-exporter: main.go
	go build \
		-ldflags "\
		-X github.com/prometheus/common/version.Version=r$(shell git rev-list --count HEAD) \
		-X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD) \
		-X github.com/prometheus/common/version.Revision=$(shell git rev-parse HEAD) \
		-X github.com/prometheus/common/version.BuildUser=$(shell whoami)@$(shell hostname) \
		-X github.com/prometheus/common/version.BuildDate=$(shell date -u "+%Y%m%d-%H:%M:%S") \
		"

clean:
	rm -f drs-replication-exporter

.PHONY: all clean

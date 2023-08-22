# multi-exporter

## External Dependencies

    $ make deps

## Building

    $ make

## Usage

```
Usage of ./multi-exporter:
  -listen-address string
        The address to listen on for HTTP requests. (default ":9998")
  -log.format value
        Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true" (default "logger:stderr")
  -log.level value
        Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal] (default "info")
```

## Example Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'dns_exporter'
    dns_sd_configs:
      - names: ['_ldap._tcp.dc._msdcs.example.com']
    metrics_path: /dns
    relabel_configs:
      - source_labels: [__address__]
        regex: (.*):[0-9]+
        target_label: __param_nameserver
      - source_labels: [__param_nameserver]
        target_label: instance
      - target_label: __address__
        replacement: 'localhost:9998'
      - source_labels: [__address__]
        target_label: probe
  - job_name: 'dns_normalized_exporter'
    dns_sd_configs:
      - names: ['_ldap._tcp.dc._msdcs.example.com']
    metrics_path: /dns_normalized
    relabel_configs:
      - source_labels: [__address__]
        regex: (.*):[0-9]+
        target_label: __param_nameserver
      - source_labels: [__param_nameserver]
        target_label: instance
      - target_label: __address__
        replacement: 'localhost:9998'
      - source_labels: [__address__]
        target_label: probe
  - job_name: 'icmp_exporter'
    dns_sd_configs:
      - names: ['_ldap._tcp.dc._msdcs.example.com']
    metrics_path: /icmp
    relabel_configs:
      - source_labels: [__address__]
        regex: (.*):[0-9]+
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: 'localhost:9998'
      - source_labels: [__address__]
        target_label: probe
```

### Notes

* The `replacement: 'localhost:9998'` under `relabel_configs` is the addresses
  of the exporter you want to scrape data from. So, if you want to be able to
  log ICMP RTTs to some server from another endpoint, you would add another
  job with a relabel config that sets the target address to a different
  exporter.

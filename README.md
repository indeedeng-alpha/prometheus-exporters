# prometheus-exporters

This is a collection of miscellaneous [Prometheus][prometheus]
[exporters][exporters].

## Introduction

The [multi-exporter][multi-exporter] is very similar to the
[blackbox-exporter][blackbox], but only does ICMP and DNS checks. RTT values
produced by the multi-exporter are more accurate than those produced by the
blackbox-exporter, and can be performed against arbitrary hosts via URL
parameters.

## License

prometheus-exporters is licensed under the [Apache 2 license](LICENSE.txt).

[prometheus]: https://prometheus.io/
[exporters]: https://prometheus.io/docs/instrumenting/exporters/
[multi-exporter]: src/multi-exporter
[blackbox]: https://github.com/prometheus/blackbox_exporter

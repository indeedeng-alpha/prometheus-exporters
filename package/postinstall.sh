#!/bin/bash

# Fix permissions because FPM can't
setcap cap_net_raw+ep /usr/bin/multi-exporter
chmod a+rx /usr/bin/multi-exporter
chmod a+r /usr/lib/systemd/system/prometheus-multi-exporter.service
chmod a+rx /usr/share/prometheus-multi-exporter

getent group prometheus-multi-exporter >/dev/null || groupadd -r prometheus-multi-exporter
getent passwd prometheus-multi-exporter >/dev/null || \
	useradd -r -g prometheus-multi-exporter -d /var/lib/prometheus-multi-exporter -s /sbin/nologin \
	-c "Prometheus Multi-Exporter User" prometheus-multi-exporter

exit 0

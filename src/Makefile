SUBDIRS := $(wildcard */.)

all: $(SUBDIRS)

$(SUBDIRS): export GOOS=linux
$(SUBDIRS): export GOARCH=amd64
$(SUBDIRS):
	$(MAKE) -C $@

deps:
	$(MAKE) -C multi-exporter deps

.PHONY: all deps $(SUBDIRS)

NAME := prometheus-multi-exporter

ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SRCDIR := $(ROOT_DIR)/../src
PKGDIR := $(ROOT_DIR)/pkg
BINDIR := $(PKGDIR)/usr/bin
SHRDIR := $(PKGDIR)/usr/share/$(NAME)
SMDDIR := $(PKGDIR)/usr/lib/systemd/system
RPMDIR := $(ROOT_DIR)/rpm

all: $(NAME)

$(NAME): $(BINDIR)/multi-exporter $(SMDDIR)/prometheus-multi-exporter.service VERSION
	mkdir $(RPMDIR)
	fpm \
		-s dir \
		-t rpm \
		-p $(RPMDIR)/$(NAME)-VERSION.ARCH.rpm \
		-n $(NAME) \
		-v $(shell cat VERSION) \
		--after-install postinstall.sh \
		-C $(PKGDIR) \
		usr

$(SMDDIR)/%: $(ROOT_DIR)/systemd/%
	install -Dm 644 $^ $@

$(BINDIR)/multi-exporter: $(SRCDIR)/multi-exporter/multi-exporter
	install -D $^ $@

VERSION:
	echo $(shell printf "r%s.%s" "$$(git rev-list --count HEAD)" "$$(git rev-parse --short HEAD)") > $@

clean:
	rm -rf pkg rpm src VERSION

.PHONY: all clean $(NAME)

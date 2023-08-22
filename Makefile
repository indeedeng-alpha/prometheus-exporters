all: src

package: src
	$(MAKE) -C $@

src: deps
	$(MAKE) -C $@

deps:
	$(MAKE) -C src deps

.PHONY: all deps package src

.PHONY: src/assets/img

BIN ?= diecast

all:
	$(BIN) -a 127.0.0.1:0

src/assets/img:
	cd src/assets/img && make

local:
	$(BIN) -L debug --disable-commands -m /go-webfriend:src src

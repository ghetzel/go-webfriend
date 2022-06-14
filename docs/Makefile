.PHONY: src/assets/img

BIN ?= diecast

UIFILES += assets/js/codemirror.js
UIFILES += assets/js/codemirror/hints/friendscript-hint.js
UIFILES += assets/js/codemirror/modes/friendscript.js
UIFILES += assets/css/codemirror.css
UIFILES += assets/css/editor/webfriend.css

all: $(UIFILES)
	$(BIN) -a 127.0.0.1:0

.PHONY: $(UIFILES)
$(UIFILES):
	@echo "Copy $(@)"
	@test -d "src/$(dir $(@))" || mkdir -p "src/$(dir $(@))"
	@cp ../server/ui/$(subst assets/,,$(@)) src/$(@)

src/assets/img:
	cd src/assets/img && make

local: $(UIFILES)
	$(BIN) -L debug --disable-commands -p local=true -m /go-webfriend:src src

SVG := $(wildcard *.svg)

rasterize: $(SVG)

.PHONY: $(SVG)
$(SVG):
	@echo "$(basename $(@))-36.png"
	@inkscape "$(@)" --without-gui --export-png="$(basename $(@))-36.png" --export-height=36 > /dev/null
	@echo "$(basename $(@))-48.png"
	@inkscape "$(@)" --without-gui --export-png="$(basename $(@))-48.png" --export-height=48 > /dev/null
	@echo "$(basename $(@))-64.png"
	@inkscape "$(@)" --without-gui --export-png="$(basename $(@))-64.png" --export-height=64 > /dev/null
	@echo "$(basename $(@))-128.png"
	@inkscape "$(@)" --without-gui --export-png="$(basename $(@))-128.png" --export-height=128 > /dev/null


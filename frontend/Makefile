all: watch

SHELL := /bin/bash
ENVFILE = if [ -e envfile ]; then set -a; source envfile; fi;
install:
	yarn install

compile: install
	$(ENVFILE) ./node_modules/webpack/bin/webpack.js \
		--progress --colors --bail

prod-compile: install
	$(ENVFILE) IS_PRODUCTION=true ./node_modules/webpack/bin/webpack.js \
		--progress --colors --bail

watch: install
	$(ENVFILE) ./node_modules/webpack/bin/webpack.js \
		--progress --colors --watch --content-base ../resources/frontend/

.PHONY: install compile prod-compile watch

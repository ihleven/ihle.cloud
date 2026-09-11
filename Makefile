OUT := ihlvn
DIR := $(shell pwd)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD | tr -d '\040\011\012\015\n')
GIT_DESC := $(shell git describe --tags --always --long --dirty) # v3.2.45-alpha_dattler-12-g4308ca9-dirty
TIMESTAMP := $(shell date +"%FT%T%z")

LDFLAGS:= "-X main.BUILD_DIR=${DIR} \
           -X main.BUILD_TIME=$(TIMESTAMP) \
           -X main.BUILD_OUTPUT=$(OUT) \
		   -X main.BUILD_BRANCH=$(BRANCH) \
           -X main.GIT_DESCRIPTION=${GIT_DESC}" 

all: ui build

version:
	@CGO_ENABLED=0 go run -ldflags=${LDFLAGS} main.go --version

ui:
	cd ui; bun install; bun run generate

build: 
	@CGO_ENABLED=0 go build -o ${OUT} -ldflags=${LDFLAGS} 

run:
	@CGO_ENABLED=0 go run -ldflags=${LDFLAGS} main.go

install: ui
	go install -ldflags=${LDFLAGS} 

linux: ui
	CGO_ENABLED=0 GOOS=linux GOARCH=$(DEPLOY_ARCH) go build -v -o ${OUT}-linux-$(DEPLOY_ARCH) -ldflags=${LDFLAGS}

# Deployment from the working tree, uncommitted changes included.
#
# Where the deployment goes is not in this file: a host and a path describe one
# person's server, not the project, and the repository is not the place to
# publish them. They come from deploy/.env (untracked, see deploy/.env.example),
# from the environment, or from the command line:
#
#   make deploy DEPLOY_HOST=me@other.host DEPLOY_PATH=/srv/app
#
# The file is read as makefile syntax, so it holds plain KEY=value lines. What
# it sets wins over the environment, which is how make treats any assignment;
# the command line wins over both. So an environment variable configures what
# the file leaves out, and `make deploy DEPLOY_HOST=...` overrides either.
DEPLOY_ENV ?= deploy/.env
-include $(DEPLOY_ENV)

# Defaults for the settings that describe the build rather than the target.
DEPLOY_BIN  ?= $(OUT)
DEPLOY_ARCH ?= amd64
# A deploy key is usually not in the agent, hence a configurable transport:
# DEPLOY_SSH="ssh -i ./id_rsa"
DEPLOY_SSH  ?= ssh

# Checked before anything is built or sent. Without this an unset target reaches
# rsync as ":/public/", which fails somewhere less obvious.
require-deploy-target:
	@[ -n "$(DEPLOY_HOST)" ] && [ -n "$(DEPLOY_PATH)" ] || { \
		echo "DEPLOY_HOST and DEPLOY_PATH are not set."; \
		echo "Copy deploy/.env.example to deploy/.env and fill it in, or pass them on the command line."; \
		exit 1; }

# deploy builds from disk and ships it. Nothing is committed or tagged, so the
# version string carries -dirty when the tree is: a hand-rolled deploy stays
# identifiable afterwards through --version.
#
# The binary is stopped first because Linux refuses to overwrite a running
# executable, and .env is never sent — configuration lives on the server and has
# to survive a deploy.
deploy: require-deploy-target linux
	@echo "deploying $(GIT_DESC) to $(DEPLOY_HOST):$(DEPLOY_PATH)"
	@$(DEPLOY_SSH) $(DEPLOY_HOST) '$(DEPLOY_PATH)/stop || true'
	rsync -az -e "$(DEPLOY_SSH)" --delete ui/.output/public/ $(DEPLOY_HOST):$(DEPLOY_PATH)/public/
	rsync -az -e "$(DEPLOY_SSH)" ${OUT}-linux-$(DEPLOY_ARCH) $(DEPLOY_HOST):$(DEPLOY_PATH)/$(DEPLOY_BIN)
	rsync -az -e "$(DEPLOY_SSH)" deploy/start deploy/stop $(DEPLOY_HOST):$(DEPLOY_PATH)/
	@$(DEPLOY_SSH) $(DEPLOY_HOST) '$(DEPLOY_PATH)/start'
	@echo "deployed. Check the log for the 'serving <url> on port <n>' line."

# What deploy would change, without changing it. Worth running first: the SPA
# sync deletes remote files that are no longer in the build.
deploy-check: require-deploy-target linux
	@echo "--- frontend"
	@rsync -azn -e "$(DEPLOY_SSH)" --delete --itemize-changes ui/.output/public/ $(DEPLOY_HOST):$(DEPLOY_PATH)/public/
	@echo "--- binary"
	@rsync -azn -e "$(DEPLOY_SSH)" --itemize-changes ${OUT}-linux-$(DEPLOY_ARCH) $(DEPLOY_HOST):$(DEPLOY_PATH)/$(DEPLOY_BIN)

cms:
	rm -r cms || true
	mkdir -p cms/{mgmt,content,pkg}
	cp ../../src/webcc-content/cms/go.mod ./cms/go.mod
	cp -r  ../../src/webcc-content/cms/mgmt/{gitrepo,search} ./cms/mgmt/
	cp ../../src/webcc-content/cms/content/*.go ./cms/content
	cp -r ../../src/webcc-content/cms/permission ./cms/
	cp -r ../../src/webcc-content/cms/pkg/errors ./cms/pkg/

clean:
	-@rm ${OUT} ${OUT}-linux-$(DEPLOY_ARCH) gin-bin
	-@rm -rf ui/.nuxt 
	-@rm -rf ui/.output 
	-@rm -rf ui/node_modules 

.PHONY: all version ui build run install linux require-deploy-target deploy deploy-check cms clean




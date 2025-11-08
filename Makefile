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
	GOOS=linux GOARCH=amd64 go build -v -o ${OUT}-linux-amd64 -ldflags=${LDFLAGS} 

opalstack: linux 
	scp -r ${OUT}-linux-amd64 ihle@opal6.opalstack.com:/home/ihle/apps/tschabrun/tschabrun; \
	scp -r ui/.output/public ihle@opal6.opalstack.com:/home/ihle/apps/tschabrun/public

cms:
	cp -r  ../../src/webcc-content/cms/mgmt/{gitrepo,search,usecase} ./cms/mgmt/
	cp ../../src/webcc-content/cms/content/* ./cms/content

clean:
	-@rm ${OUT} ${OUT}-linux-amd64 gin-bin
	-@rm -rf ui/.nuxt 
	-@rm -rf ui/.output 
	-@rm -rf ui/node_modules 

.PHONY: all version ui build run install linux opalstack cms clean




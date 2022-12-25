OUT := cloud
PKG := github.com/ihleven/ihle.cloud
VERSION := $(shell git describe --always --long --dirty)
CLIENT_ID:=$(CLIENT_ID)
CLIENT_SECRET:=$(CLIENT_SECRET)
LDFLAGS:= "-X main.version=${VERSION} -X main.CLIENT_ID=${CLIENT_ID} -X main.CLIENT_SECRET=${CLIENT_SECRET}" 


all: nuxt build

nuxt:
	yarn generate

build: 
	cd cli/cld; go build  -ldflags=${LDFLAGS} 

run: build
	cd cli/cld; go build -o ${OUT} -ldflags=${LDFLAGS}; ./${OUT}

install: nuxt
	cd cli; go install -ldflags=${LDFLAGS} 

linux: nuxt
	cd cli/cld;  GOOS=linux GOARCH=amd64 go build -v -o ${OUT}-linux-amd64 -ldflags=${LDFLAGS} 

opalstack: linux 
	scp -r ./cli/cld/cloud-linux-amd64 .output/public ihle@opal6.opalstack.com:/home/ihle/apps/ihle_cloud/


.PHONY: all nuxt build install linux opalstack
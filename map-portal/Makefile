# Makefile for map-console project
#
# Targets:
#
# dependencies:
# local build: need install docker, nodejs and yarn
# standard build: only need install docker, all build will in docker container.
#
# -----used for local develop----
# local:	build_package build_image
# build_package: build package in local
#    example: make build_package
# build_image: build image in local
#    example: make build_image -e IMAGEREGISTRY=registry.example.com -e IMAGEREPO=map
#
# run: run console in local
#    example: make run -e MAP_API_ENDPOINT="http://192.168.2.105:9800" -e DEX_IDP_ENDPOINT="http://192.168.2.105:5556" -e IMAGEREGISTRY=registry.example.com -e IMAGEREPO=map
#
# -----used for image release and standard image build-----
# release: release_image push_image
# release_image: standard image build in docker container
# push_image: push image to registry.
#    example: make push_image -e IMAGEREGISTRY=registry.example.com -e IMAGEREPO=map
#
# -----used for clean -----
# cleanall: clean_package clean_images
# clean_package: clean package build output
# clean_images: clean images in local
#    example: make clean_image -e IMAGEREGISTRY=registry.example.com -e IMAGEREPO=map


IMAGEREGISTRY?=harbor.huaruicloud.cn
IMAGEREPO?=map
YARN_REGISTRY?=https://mirrors.cloud.tencent.com/npm/
# used for run console in local
MAP_API_ENDPOINT="http://192.168.2.105:9800"
DEX_IDP_ENDPOINT="http://192.168.2.105:5556"

# version parameters
BUILDPATH=$(CURDIR)
PACKAGEPATH=dist
SRCPATH=./src
VERSION ?= $(shell cat ${BUILDPATH}/VERSION)
IMAGEVERSION="${IMAGEREGISTRY}/${IMAGEREPO}/map-console:${VERSION}"
LASTIMAGEVERSION="${IMAGEREGISTRY}/${IMAGEREPO}/map-console:latest"

# docker parameters
DOCKERCMD=$(shell which docker)
DOCKERLOGIN=$(DOCKERCMD) login
DOCKERIMAGE=$(DOCKERCMD) images
DOCKERBUILD=$(DOCKERCMD) build
DOCKERRMI=$(DOCKERCMD) rmi
DOCKERTAG=$(DOCKERCMD) tag
DOCKERPUSH=$(DOCKERCMD) push
DOCKERRM=$(DOCKERCMD) rm
DOCKERRUN=$(DOCKERCMD) run

.PHONY: local
local: build_package build_image

build_package:
	$(info "====build package for map console...====")
	yarn config set registry ${YARN_REGISTRY} && yarn && yarn build

build_image:
	$(info "====build image for map console in local...====")
	${DOCKERBUILD} -f ${BUILDPATH}/Dockerfiles/Dockerfile-local -t ${IMAGEVERSION} .

run:
	$(info "====run console image in local====")
	@${DOCKERRM} -f map-console && \
	${DOCKERRUN} -d -e MAP_API_ENDPOINT=${MAP_API_ENDPOINT} -e DEX_IDP_ENDPOINT=${DEX_IDP_ENDPOINT} -e TZ='Asia/Shanghai' -p 0.0.0.0:80:80 --restart unless-stopped --name map-console ${IMAGEVERSION}

.PHONY: release
release: release_image push_image

release_image:
	$(info "====build image of console for release...====")
	${DOCKERBUILD} --build-arg YARN_REGISTRY=${YARN_REGISTRY} -f ${BUILDPATH}/Dockerfiles/Dockerfile -t ${IMAGEVERSION} . && ${DOCKERTAG} ${IMAGEVERSION} ${LASTIMAGEVERSION}

push_image:
		$(info "====push image to registry...====")
		${DOCKERLOGIN} ${IMAGEREGISTRY} && \
		${DOCKERPUSH} ${IMAGEVERSION} && \
		${DOCKERPUSH} ${LASTIMAGEVERSION}

.PHONY: cleanall
cleanall: clean_package clean_images

clean_package:
	@if [ -d $(BUILDPATH)/dist ] ; then rm -rf $(BUILDPATH)/dist ; fi

clean_images:
	@if [ $(shell ${DOCKERIMAGE} -q $(IMAGEVERSION)) != "" ]; then ${DOCKERRMI} -f ${IMAGEVERSION} ;fi
	@if [ $(shell ${DOCKERIMAGE} -q $(LASTIMAGEVERSION)) != "" ]; then ${DOCKERRMI} -f ${LASTIMAGEVERSION} ;fi

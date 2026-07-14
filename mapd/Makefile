mkfile_path = $(abspath $(lastword $(MAKEFILE_LIST)))
project_root = $(dir $(mkfile_path))
build_path = $(project_root)/bin
platform = $(shell uname -s)

TANZU_STANDARD_REPO_VERSION = v2.1.0
IMAGE_REGISTRY ?= registry-stg.eng.vmware.com
GOPRIVATE = gitlab.eng.vmware.com

LDFLAGS := -s -w
VERSION ?= $(shell cat ${project_root}/VERSION)
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd.Version=$(VERSION)
BRANCH ?= $(shell git name-rev --name-only HEAD)
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd.Branch=$(BRANCH)
REVISION ?= $(shell git rev-parse --short HEAD | tr [:lower:] [:upper:])
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd.Revision=$(REVISION)
BUILD_DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd.BuildDate=$(BUILD_DATE)
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd.PublicImageRegistry=$(IMAGE_REGISTRY)
ifeq ($(platform),Darwin)
	version_major = $(shell uname -r | cut -f1 -d.)
	ifeq ($(shell test ${version_major} -ge 22 && echo TRUE),TRUE)
		# Ventura 13 or later
		ADMIN_POLICY_BASE64 ?= $(shell base64 -i ${project_root}/docs/role/admin)
		USER_POLICY_BASE64 ?= $(shell base64 -i ${project_root}/docs/role/user)
		GUEST_POLICY_BASE64 ?= $(shell base64 -i ${project_root}/docs/role/guest)
	else
		# Monterey 12 or before
		ADMIN_POLICY_BASE64 ?= $(shell base64 ${project_root}/docs/role/admin)
		USER_POLICY_BASE64 ?= $(shell base64 ${project_root}/docs/role/user)
		GUEST_POLICY_BASE64 ?= $(shell base64 ${project_root}/docs/role/guest)
	endif
else
	ADMIN_POLICY_BASE64 ?= $(shell base64 -w 0 ${project_root}/docs/role/admin)
	USER_POLICY_BASE64 ?= $(shell base64 -w 0 ${project_root}/docs/role/user)
	GUEST_POLICY_BASE64 ?= $(shell base64 -w 0 ${project_root}/docs/role/guest)
endif
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model.RoleBindingPolicyAdminRawBase64=$(ADMIN_POLICY_BASE64)
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model.RoleBindingPolicyUserRawBase64=$(USER_POLICY_BASE64)
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model.RoleBindingPolicyGuestRawBase64=$(GUEST_POLICY_BASE64)
LDFLAGS += -X gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu.TKGDefaultStandardRepoVersion=$(TANZU_STANDARD_REPO_VERSION)

all: build

testsuite: test race benchmark

install:
	@GOPRIVATE=${GOPRIVATE} go install gitlab.eng.vmware.com/cn-se-solution/mapd/cmd/...

test:
	@GOPRIVATE=${GOPRIVATE} go test -cpu 1,4 -timeout 5m gitlab.eng.vmware.com/cn-se-solution/mapd/...

race:
	@GOPRIVATE=${GOPRIVATE} go test -race -cpu 1,4 -timeout 7m gitlab.eng.vmware.com/cn-se-solution/mapd/...

benchmark:
	@GOPRIVATE=${GOPRIVATE} go test -bench . -cpu 1,4 -timeout 10m gitlab.eng.vmware.com/cn-se-solution/mapd/...

devel:
	@GOPRIVATE=${GOPRIVATE} go mod tidy

build:
	@rm -rf ${build_path} && \
		mkdir -p ${build_path} && \
		d=$$(date +%s) && \
		GOPRIVATE=${GOPRIVATE} go build -trimpath -v -o ${build_path} -ldflags "${LDFLAGS}" gitlab.eng.vmware.com/cn-se-solution/mapd/cmd/... && \
		echo "DONE. [$$(($$(date +%s)-d))s]" 1>&2

code: pkg/util/k8s/client pkg/graphql

pkg/graphql:
	@pushd ${project_root} &>/dev/null && \
		GOPRIVATE=${GOPRIVATE} go generate gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/resolver && \
		popd &>/dev/null

pkg/util/k8s/client:
	@pushd ${project_root} &>/dev/null && \
		GOPRIVATE=${GOPRIVATE} go generate gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/client && \
		popd &>/dev/null

clean:
	@GOPRIVATE=${GOPRIVATE} go clean -i gitlab.eng.vmware.com/cn-se-solution/mapd/... && \
		rm -rf ${project_root}/dist ${project_root}/bin ${project_root}/scripts/packer/tmp ${project_root}/scripts/packer/common/package/core.tar

.PHONY: \
	testsuite \
	install \
	test \
	race \
	benchmark \
	devel \
	build \
	code \
	pkg/graphql \
	pkg/util/k8s/client \
	clean

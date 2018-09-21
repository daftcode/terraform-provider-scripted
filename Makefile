OUT := ${PWD}/dist
VERSION := $$(${OUT}/version)
ifneq ($(DELVE_PORT),)
	VERSION := "${VERSION}+delve-$(DELVE_PORT)"
endif

TF_DIR := $(HOME)/.terraform.d
TF_PLUGINS := $(TF_DIR)/plugins
TF_SCHEMAS := $(TF_DIR)/schemas
GOPATH := $(if $(GOPATH),$(GOPATH),$(HOME)/go)
NAME := scripted
TF_NAME := terraform-provider-$(NAME)
BIN := ${TF_NAME}_$(VERSION)
BIN_PATH := $(GOPATH)/bin/$(BIN)
TEST ?= $$(go list ./...)

BUILD_CMD := go build -o "${OUT}/${BIN}"
ifneq ($(DELVE_PORT),)
	BUILD_CMD := go build -gcflags "all=-N -l" -o "${OUT}/${BIN}"
endif


all: fmt test build

fmt:
	go fmt ./${NAME} ./cmd/*

test: fmtcheck
	TF_ACC=1 TF_SCRIPTED_LOGGING_LOG_LEVEL=WARN go test -v ./${NAME}

debug_test:
	TF_ACC=1 TF_SCRIPTED_ENV_PREFIX=TFS_ TFS_LOGGING_LOG_LEVEL=TRACE go test -v ./${NAME}


build_cmds:
	for name in $$(ls ./cmd); do go build -o "${OUT}/$${name}" "./cmd/$${name}"; done

schema: build_cmds
	"${OUT}/generate-schema" "${TF_NAME}" "${OUT}"

docs: schema
	mkdir -p "docs/api"
	"${OUT}/generate-docs" "${OUT}/${BIN}.json" "docs/api"

build_provider_cur:
	echo -n "${VERSION}" > "${OUT}/VERSION"
	${BUILD_CMD}

build_provider_all: build_provider_cur
	GOOS=linux  GOARCH=amd64 go build -o "${OUT}/${BIN}-linux-amd64"
	GOOS=darwin GOARCH=amd64 go build -o "${OUT}/${BIN}-darwin-amd64"

build: schema docs build_provider_all
	(cd dist && sha256sum -b "${BIN}.json" "${BIN}-"* > "${BIN}.sha256sums")


install: docs build_provider_cur
	mkdir -p "${TF_PLUGINS}" "${TF_SCHEMAS}"
	cp "dist/${BIN}" "${BIN_PATH}"
	cp "${OUT}/${BIN}.json" "${TF_SCHEMAS}/${BIN}.json"
	ln -sfT "${BIN_PATH}" "${TF_PLUGINS}/${BIN}"

release:
	if [ "$$(cat "${OUT}/VERSION")" = "${VERSION}" ] ; then echo "tagging ${VERSION}"; else echo "version ${VERSION} is not built!"; exit 1; fi;
	git diff --quiet
	git tag -a "${VERSION}"
	git push --follow-tags

fmtcheck:
	l=`gofmt -l ${NAME}`; if [ -n "$$l" ]; then echo "Following needs formatting (gofmt):"; echo "$$l"; exit 1; fi

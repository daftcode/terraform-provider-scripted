VERSION := $$(cat VERSION)
OUT := ./dist
TF_DIR := $(HOME)/.terraform.d
TF_PLUGINS := $(TF_DIR)/plugins
TF_SCHEMAS := $(TF_DIR)/schemas
GOPATH := $(if $(GOPATH),$(GOPATH),$(HOME)/go)
NAME := scripted
TF_NAME := terraform-provider-$(NAME)
BIN := $(TF_NAME)_$(VERSION)
BIN_PATH := $(GOPATH)/bin/$(BIN)
TEST ?= $$(go list ./...)

all: fmt test build

fmt:
	go fmt ./${NAME} ./cmd/*

test: fmtcheck
	TF_ACC=1 TF_SCRIPTED_LOGGING_LOG_LEVEL=WARN go test -v ./${NAME}

build_cmds:
	for name in $$(ls ./cmd); do go build -o "${OUT}/$${name}" "./cmd/$${name}"; done

schema: build_cmds
	"${OUT}/generate-schema" "${TF_NAME}" "${VERSION}" "${OUT}"

docs: schema
	mkdir -p "docs/api"
	"${OUT}/generate-docs" "${OUT}/${TF_NAME}.json" "docs/api"

build_provider:
	echo -n "${VERSION}" > "${OUT}/VERSION"
	go build -o "${OUT}/${TF_NAME}"

build: schema docs build_provider

install: build
	mkdir -p "${TF_PLUGINS}" "${TF_SCHEMAS}"
	cp "dist/${TF_NAME}" "${BIN_PATH}"
	cp "${OUT}/${TF_NAME}.json" "${TF_SCHEMAS}/${BIN}.json"
	ln -sfT "${BIN_PATH}" "${TF_PLUGINS}/${BIN}"

release:
	if [ "$$(cat "${OUT}/VERSION")" = "${VERSION}" ] ; then echo "tagging ${VERSION}"; else echo "version ${VERSION} is not built!"; exit 1; fi;
	git diff --quiet
	git tag -a "${VERSION}"
	git push --follow-tags

fmtcheck:
	l=`gofmt -l ${NAME}`; if [ -n "$$l" ]; then echo "Following needs formatting (gofmt):"; echo "$$l"; exit 1; fi

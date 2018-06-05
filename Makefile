VERSION := v1.0.0
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

all: fmt test schema docs provider

fmt:
	go fmt ${NAME}/*

test: fmtcheck
	TF_ACC=1 TF_SCRIPTED_LOGGING_LOG_LEVEL=WARN go test -v ./${NAME}

build-cmds:
	for name in $$(ls ./cmd); do go build -o "${OUT}/$${name}" "./cmd/$${name}"; done

schema: build-cmds
	"${OUT}/generate-schema" "${TF_NAME}" "${VERSION}" "${OUT}"

docs: schema
	"${OUT}/generate-docs" "${OUT}/${TF_NAME}.json"

provider:
	go build -o "${OUT}/${TF_NAME}"

build: provider

install: schema docs provider
	mkdir -p "${TF_PLUGINS}" "${TF_SCHEMAS}"
	cp "dist/${TF_NAME}" "${BIN_PATH}"
	cp "${OUT}/${TF_NAME}.json" "${TF_SCHEMAS}"
	ln -sfT "${BIN_PATH}" "${TF_PLUGINS}/${BIN}"

fmtcheck:
	l=`gofmt -l ${NAME}`; if [ -n "$$l" ]; then echo "Following needs formatting (gofmt):"; echo "$$l"; exit 1; fi

PKG_LIST := $(shell go list ./... | grep -v /vendor/)

export SERVICE_NAME := pvzops
export PROJECT_NAME := pvzops

export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)

GOLANG_CI_LINT_VERSION ?= v1.64.8
OAPI_CODEGEN_VERSION ?= v2.3.0
MOCKGEN_VERSION ?= v0.5.0
GOOSE_VERSION := 3.20.0
SERVICE_NAME := pvzops

UNAME_S := $(shell uname -s)
UNAME_P := $(shell uname -p)

ifeq ($(UNAME_S),Linux)
	OSFLAG = linux
endif

ifeq ($(UNAME_S),Darwin)
	OSFLAG = osx
  ifeq ($(UNAME_P),arm)
    # protobuf team doesn't create releases
    # for Apple M1 (arm) chipset
	  UNAME_P = x86_64
  endif
  ifeq ($(UNAME_P),i386)
    # for Rosetta 2 emulator
	  UNAME_P = x86_64
  endif
endif

HOST_ARCH := "$(OSFLAG)-$(UNAME_P)"

.PHONY: all
all: build

.PHONY: build
build:
	go build -o bin/$(SERVICE_NAME) cmd/$(SERVICE_NAME)/main.go

.PHONY: run
run: build
	./bin/$(SERVICE_NAME)

.PHONY: clean
clean:
	rm -rf ./bin

./bin:
	mkdir -p ./bin

./migrations:
	mkdir -p ./migrations

./bin/gomock: | ./bin
	go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: ./bin/codegen
./bin/codegen:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)

.PHONY: test
test:
	GOGC=off go test -short ./...

.PHONY: test-db
test-db:
	GOGC=off go test -tags dbtest -count=1 ./...

.PHONY: install
install:
	go install github.com/client9/misspell/cmd/misspell@latest
	go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)

./bin/goose: | ./bin
	go install github.com/pressly/goose/v3/cmd/goose@v$(GOOSE_VERSION)

PG_DSN ?= user:password@localhost:5432
DB_NAME := pvz
DB_NAME_TEST := pvzops_test

migrate: ./bin/goose ./migrations
	@echo "\n🆙 postgresql migrations"
	docker-compose exec db psql -U user -d postgres -c "drop database if exists $(DB_NAME);"
	docker-compose exec db psql -U user -d postgres -c "create database $(DB_NAME);"
	./bin/goose -v -dir migrations postgres "postgresql://$(PG_DSN)/$(DB_NAME)?sslmode=disable" up

# example usage:  make create-migration name=create_subtasks_tables
.PHONY: create migration
create-migration: | ./migrations
	goose -dir migrations create $(name) sql

.PHONY: db-up-no-docker
db-up-no-docker:
	psql postgres  -c "drop database if exists $(DB_NAME_TEST);"
	psql postgres  -c "create database $(DB_NAME_TEST)"
	goose -allow-missing -dir migrations postgres "dbname=$(DB_NAME_TEST) sslmode=disable " up

.PHONY: db-down-no-docker
db-down-no-docker:
	psql postgres  -c "drop database if exists $(DB_NAME_TEST);"

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: mocks
mocks: ./bin/gomock
	go generate ./...

.PHONY:
docker-up:
	@docker-compose up -d --build

.PHONY:
docker-down:
	@docker-compose down -v

.PHONY:  generate
generate: ./bin/codegen ./bin/gomock
	go generate ./...

.PHONY: pre
pre:
	chmod -R +x .pre-commit-checks
	pre-commit run --all-files

PKG_LIST := $(shell go list ./... | grep -v /vendor/)

export SERVICE_NAME := pvzops
export PROJECT_NAME := pvzops

export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)

GOLANG_CI_LINT_VERSION ?= v1.64.8
OAPI_CODEGEN_VERSION ?= v2.3.0
MOCKGEN_VERSION ?= v0.5.0
PROTOC_VERSION := 26.1
GOOSE_VERSION := 3.20.0
GRPCUI_VERSION := 1.4.1

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
PROTOC_ZIP := protoc-$(PROTOC_VERSION)-$(HOST_ARCH).zip

.PHONY: all
all: build test lint

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

./bin/protoc.zip: | ./bin
	curl -L https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP) -o ./bin/protoc.zip

./bin/protoc: ./bin/protoc.zip
	unzip -o ./bin/protoc.zip -d ./ bin/protoc

./proto/include: ./bin/protoc.zip
	unzip -o ./bin/protoc.zip -d ./api/proto include/*

./bin/protoc-gen-go: | ./bin
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

./bin/protoc-gen-go-grpc: | ./bin
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

./bin/gomock: | ./bin
	go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: ./bin/codegen
./bin/codegen:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)

PROTOS = $(wildcard ./api/proto/*.proto)

.PHONY: ./pkg
./pkg: PROTOC_OPT ?= module=gitlab.services.mts.ru/media/projects/puma/$(PROJECT_NAME)/pkg:./pkg
./pkg: ./bin/protoc ./proto/include ./bin/protoc-gen-go ./bin/protoc-gen-go-grpc $(PROTOS)
	mkdir -p $@
	protoc \
	-I ./api/proto \
	-I ./api/proto/include \
	--go_out=$(PROTOC_OPT) \
    --go-grpc_out=require_unimplemented_servers=false,$(PROTOC_OPT) \
    --descriptor_set_out=./api/proto/$(SERVICE_NAME).protoset  \
    --include_imports \
    ./api/proto/pvzops/*.proto

.PHONY: ./bin/golangci-lint
./bin/golangci-lint: | ./bin
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)

.PHONY: lint
lint: ./bin/golangci-lint
	./bin/golangci-lint run --timeout 5m ./...

.PHONY: fix-lint
fix-lint: ./bin/golangci-lint
	./bin/golangci-lint run --fix --timeout=5m ./...

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

./bin/grpcui: | ./bin
	go install github.com/fullstorydev/grpcui/cmd/grpcui@v$(GRPCUI_VERSION)

./bin/goose: | ./bin
	go install github.com/pressly/goose/v3/cmd/goose@v$(GOOSE_VERSION)

PG_DSN ?= postgres@localhost:5432
DB_NAME := pvsops
DB_NAME_TEST := pvsops_test

migrate: ./bin/goose ./migrations
	@echo "\n🆙 postgresql migrations"
	docker-compose exec pg psql -U postgres -c "drop database if exists $(DB_NAME);"
	docker-compose exec pg psql -U postgres -c "create database $(DB_NAME);"
	goose -v -dir migrations postgres "postgresql://$(PG_DSN)/$(DB_NAME)?sslmode=disable" up

migrate-test: ./bin/goose
	docker-compose -f docker-compose.yml up pg -d
	docker-compose exec pg psql -U postgres -c "drop database if exists $(DB_NAME_TEST);"
	docker-compose exec pg psql -U postgres -c "create database $(DB_NAME_TEST);"
	goose -v -allow-missing -dir migrations postgres "postgresql://$(PG_DSN)/$(DB_NAME_TEST)?sslmode=disable" up

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

.PHONY: test-local
test-local:
	psql postgres  -c "drop database if exists $(DB_NAME);"
	createdb $(DB_NAME)
	goose -allow-missing -dir migrations postgres "dbname=$(DB_NAME) sslmode=disable " up

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: mocks
mocks: ./bin/gomock
	go generate ./...

.PHONY: mock-clients
mock-clients: mocks
	mockgen -package=mock_originlive gitlab.services.mts.ru/media/projects/puma/origin-live/pkg/origin-live OriginLiveClient > internal/mocks/mock_originlive/mock_originlive.go &

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

# https://github.com/fullstorydev/grpcui
GRPC_LISTEN_PORT ?= 8082
.PHONY: grpcui
grpcui: ./bin/grpcui
	grpcui -protoset api/proto/$(SERVICE_NAME).protoset \
		-plaintext \
		-v \
		-port 30001 \
		127.0.0.1:$(GRPC_LISTEN_PORT)

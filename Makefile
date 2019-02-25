.PHONY : dep lint test integration coverage doc build run deploy
DIR := $(shell pwd)
IMAGE := $(shell basename $(DIR))
VERSION := $(shell git rev-parse --short HEAD)
ifeq ($(shell git diff --cached --quiet; echo $$?),1)
    VERSION := $(VERSION)$(shell id -un)
endif
ENVIRON := dev

dep:
	dep ensure

lint:
	golangci-lint run --config .golangci.yaml ./...

test:
	mkdir -p .coverage
	go test -v -cover -coverpkg=./... -coverprofile=.coverage/unit.cover.out ./...
	gocov convert .coverage/unit.cover.out | gocov-xml > .coverage/unit.xml

integration:
	mkdir -p .coverage
	go test -v -cover -coverpkg=./... -coverprofile=.coverage/integration.cover.out ./tests/
	gocov convert .coverage/integration.cover.out | gocov-xml > .coverage/integration.xml

coverage:
	mkdir -p .coverage
	gocovmerge .coverage/*.cover.out > .coverage/combined.cover.out
	gocov convert .coverage/combined.cover.out | gocov-xml > .coverage/combined.xml

doc:
	godoc -http ':9090'

build:
	docker build -t atlassian/$(IMAGE):$(VERSION) .

run:
	docker run -ti atlassian/$(IMAGE):$(VERSION)

deploy:
	docker push atlassian/$(IMAGE):$(VERSION)

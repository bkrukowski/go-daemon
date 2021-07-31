.EXPORT_ALL_VARIABLES:
GO_DAEMON_CONFIG = ./sample/.go-daemon.yml

tests-unit:
	go test -coverprofile=coverage.out ./cmd/... ./pkg/...

format-ci: OUTPUT = $(shell go fmt ./...)
format-ci:
	if [ -n "${OUTPUT}" ]; then echo 'Execute "make format"'; exit 1; fi

format:
	go fmt ./...

build: DATETIME = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
build: GITHASH = $(shell git rev-parse --short HEAD)
build: VERSION = dev-$(shell git rev-parse --abbrev-ref HEAD)
build: DIRTY_SUFFIX = $(shell git diff --quiet || echo '-dirty')
build: format
	[ -n "${GO_DAEMON_CACHE_BUILD}" ] && test -f app.bin || go build -v -ldflags="-X 'main.date=${DATETIME}' -X 'main.commit=${GITHASH}${DIRTY_SUFFIX}' -X 'main.version=${VERSION}'" -o app.bin main.go

help: build
	./app.bin help run

version: build
	./app.bin -v

clock: build
	./app.bin run clock -v --timeout=5s

ping-error: build
	./app.bin run ping-error -v --timeout=5s

ping-ok: build
	./app.bin run ping-ok -v --timeout=5s

ping: build
	./app.bin run --tag=ping -v --timeout=5s

inline: build
	./app.bin run inline -v --timeout=5s

ignore-signals: build
	./app.bin run ignore-signals -v --timeout=5s

TIMEOUT ?= 5s
FLAG_TIMEOUT = --timeout=$(TIMEOUT)

.EXPORT_ALL_VARIABLES:
GO_DAEMON_CONFIG = ./examples/.go-daemon.yml

tests-unit:
	go test -coverprofile=coverage.out ./cmd/... ./pkg/...

format-ci: OUTPUT = $(shell go fmt ./...)
format-ci:
	if [ -n "${OUTPUT}" ]; then\
		echo 'Execute "make format"';\
		exit 1;\
	fi

format:
	go fmt ./...

build: DATETIME = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
build: GITHASH = $(shell git rev-parse --short HEAD)
build: VERSION = dev-$(shell git rev-parse --abbrev-ref HEAD)
build: DIRTY_SUFFIX = $(shell git diff --quiet || echo '-dirty')
build: format
	if [ -n "${GO_DAEMON_CACHE_BUILD}" ] && test -f app.bin; then\
		echo "Do not re-compile, used cached binary";\
	else\
		go build -v -ldflags="-X 'main.date=${DATETIME}' -X 'main.commit=${GITHASH}${DIRTY_SUFFIX}' -X 'main.version=${VERSION}'" -o app.bin main.go;\
	fi

help: build
	./app.bin help run

version: build
	./app.bin -v

clock: build
	./app.bin run clock -v ${FLAG_TIMEOUT}

ping-error: build
	./app.bin run ping-error -v ${FLAG_TIMEOUT}

ping-ok: build
	./app.bin run ping-ok -v ${FLAG_TIMEOUT}

ping: build
	./app.bin run --tag=ping -v ${FLAG_TIMEOUT}

inline: build
	./app.bin run inline -v ${FLAG_TIMEOUT}

ignore-signals: build
	./app.bin run ignore-signals -v ${FLAG_TIMEOUT}

GODEP_BIN := $(GOPATH)/bin/godep
GODEP := $(shell [ -x $(GODEP_BIN) ] && echo $(GODEP_BIN) || echo '')

# Initialize version flag
GO_LDFLAGS := -X $(shell go list ./).GitCommit=$(shell git rev-parse --short HEAD 2>/dev/null)

default: build

bin/docker-machine-driver-vscale:
	go build -i -ldflags "$(GO_LDFLAGS)" \
	-o ./bin/docker-machine-driver-vscale ./bin

build: clean bin/docker-machine-driver-vscale

clean:
	$(RM) bin/docker-machine-driver-vscale

install: bin/docker-machine-driver-vscale
	cp -f ./bin/docker-machine-driver-vscale $(GOPATH)/bin/

test-acceptance:
	test/integration/run-bats.sh test/integration/bats/

dep-save:
	$(if $(GODEP), , \
		$(error Please install godep: go get github.com/tools/godep))
	$(GODEP) save $(shell go list ./... | grep -v vendor/)

dep-restore:
	$(if $(GODEP), , \
		$(error Please install godep: go get github.com/tools/godep))
	$(GODEP) restore -v
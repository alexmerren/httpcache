GO ?= go

SOURCE = round_tripper.go
CONTAINER = http-cache
SOURCE_PATH = /go/src/github.com/alexmerren/httpcache
TEST = .
DRUN = docker run -v `pwd`:$(SOURCE_PATH) -i -t $(CONTAINER)

build:
	docker build -t $(CONTAINER) .

test:
	$(DRUN) $(GO) test $(LDFLAGS) ./ -run $(TEST) -timeout 10s $(ARGS) -v

fmt:
	$(DRUN) $(GO) fmt ./...

vet:
	$(DRUN) $(GO) vet ./.

bash:
	$(DRUN) /bin/bash
APPNAME:=tv_grab_url
ARCH_POSTFIX?=
RELEASE?=0
CGO_ENABLED?=0

ifeq ($(RELEASE), 1)
	# Strip debug information from the binary
	GO_LDFLAGS+=-s -w
endif
GO_LDFLAGS:=-ldflags="$(GO_LDFLAGS)"


.PHONY: default
default: test

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) go build $(GO_LDFLAGS) -o ./build/$(APPNAME) -v cmd/main.go

.PHONY: test
test: build
	CGO_ENABLED=$(CGO_ENABLED) go test -v ./...

.PHONY: clean
clean:
	rm -rf ./build
	rm -rf ./dist

.PHONY: install
install: build
	mkdir -p ./dist
	cp ./build/$(APPNAME) ./dist/$(APPNAME)$(ARCH_POSTFIX)
	chmod +x ./dist/$(APPNAME)$(ARCH_POSTFIX)

ORGURL=github.com/moustafab
APPNAME=tekton-es-logs
REPOURL=${ORGURL}/${APPNAME}
COMMIT:=$(shell git rev-parse --verify HEAD)
SHORTSHA:=$(shell git rev-parse --short HEAD)
BRANCH:=$(shell git rev-parse --abbrev-ref HEAD)
VERSION?=$(BRANCH)-$(SHORTSHA)
BRANCH_VERSION:=$(shell echo $(BRANCH) | tr '/' '-')
DATE:=$(shell date +%FT%T%z)
USER?=unkown
RELEASE?=0
OUTPUT_DIR=build

GOPATH?=$(shell go env GOPATH)
GO_LDFLAGS+=-X 'main.gitVersion=$(VERSION)'
ifeq ($(RELEASE), 1)
	# Strip debug information from the binary
	GO_LDFLAGS+=-s -w
endif
GO_LDFLAGS:=-ldflags="$(GO_LDFLAGS)"

DOCKER_ACCOUNT:=moustafab
DOCKER_IMAGE:=$(DOCKER_ACCOUNT)/$(APPNAME)

# See: https://docs.docker.com/engine/reference/commandline/tag/#extended-description
# A tag name must be valid ASCII and may contain lowercase and uppercase letters, digits, underscores, periods and dashes.
# A tag name may not start with a period or a dash and may contain a maximum of 128 characters.
DOCKER_TAG:=$(shell echo $(VERSION) | tr '/' '-')

LEVEL=debug

.PHONY: default
default: build

GOLANGCILINTVERSION:=1.18.0
GOLANGCILINT=$(GOPATH)/bin/golangci-lint
$(GOLANGCILINT):
	curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v$(GOLANGCILINTVERSION)

.PHONY: build
build:
	mkdir -p ./${OUTPUT_DIR}
	go build -mod=mod $(GO_LDFLAGS) -o ./${OUTPUT_DIR}/

.PHONY: lint
lint: $(GOLANGCILINT)
	golangci-lint run

.PHONY: format
format:
	gofmt -s -w

.PHONY: test
test:
	go test -mod=mod $(GO_LDFLAGS) -v

.PHONY: clean
clean:
	rm -rf ./${OUTPUT_DIR}

.PHONY: docker-build
docker-build:
	docker build --network host --build-arg VERSION=$(VERSION) --tag $(DOCKER_IMAGE):$(BRANCH_VERSION) .
	docker tag $(DOCKER_IMAGE):$(BRANCH_VERSION) $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(BRANCH_VERSION)
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

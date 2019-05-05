GOTOOLS = github.com/Masterminds/glide
ENI_LIB?=$(HOME)/.devchain/eni/lib
CGO_LDFLAGS = -L$(ENI_LIB) -Wl,-rpath,$(ENI_LIB)
CGO_LDFLAGS_ALLOW = "-I.*"
UNAME = $(shell uname)

all: get_vendor_deps install print_logo

get_vendor_deps: tools
	glide install
	@# cannot use ctx (type *"gopkg.in/urfave/cli.v1".Context) as type
	@# *"github.com/second-state/devchain/vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave/cli.v1".Context ...
	@rm -rf vendor/github.com/ethereum/go-ethereum/vendor/gopkg.in/urfave

install:
	@echo "\n--> Installing the Second State DevChain\n"
ifeq ($(UNAME), Linux)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go install ./cmd/devchain
endif
ifeq ($(UNAME), Darwin)
	CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go install ./cmd/devchain
endif
	@echo "\n\nThe Second State DevChain has successfully installed!"

tools:
	@echo "--> Installing tools"
	go get $(GOTOOLS)
	@echo "--> Tools installed successfully"

build: get_vendor_deps
ifeq ($(UNAME), Linux)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go build -o build/devchain ./cmd/devchain
endif
ifeq ($(UNAME), Darwin)
	CGO_LDFLAGS_ALLOW="$(CGO_LDFLAGS_ALLOW)" go build -o build/devchain ./cmd/devchain
endif

NAME := wangshishuo/devchain
LATEST := ${NAME}:latest
#GIT_COMMIT := $(shell git rev-parse --short=8 HEAD)
#IMAGE := ${NAME}:${GIT_COMMIT}

docker_image:
	docker build -t ${LATEST} .

push_tag_image:
	docker tag ${LATEST} ${NAME}:${TAG}
	docker push ${NAME}:${TAG}

push_image:
	docker push ${LATEST}

dist:
	docker run --rm -e "BUILD_TAG=${BUILD_TAG}" -v "${CURDIR}/scripts":/scripts --entrypoint /bin/sh -t ${LATEST} /scripts/dist.ubuntu.sh
	docker build -t ${NAME}:centos -f Dockerfile.centos .
	docker run --rm -e "BUILD_TAG=${BUILD_TAG}" -v "${CURDIR}/scripts":/scripts --entrypoint /bin/sh -t ${NAME}:centos /scripts/dist.centos.sh
	rm -rf build/dist && mkdir -p build/dist && mv -f scripts/*.zip build/dist/ && cd build/dist && sha256sum *.* > SHA256SUMS

print_logo:
	@echo "Visit our website < https://www.secondstate.io/ >, to learn more.\n"

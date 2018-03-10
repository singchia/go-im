BINARY = go-im
VERSION?= 0.10

GOARCH :=
GOOS :=
ifeq ($(OS),Windows_NT)
	GOOSs += windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		GOARCH += amd64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		GOARCH += ia32
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		GOOS += linux
	endif
	ifeq ($(UNAME_S),Darwin)
		GOOS += darwin
	endif

	UNAME_P := $(shell uname -p)
	ifeq ($(UNAME_P),x86_64)
		GOARCH += amd64
	endif
	ifneq ($(filter %86,$(UNAME_P)),)
		GOARCH += ia32
	endif
	ifneq ($(filter arm%,$(UNAME_P)),)
		GOARCH += arm
	endif
endif

ifeq ($(strip ${GOOS}),darwin)
	GOARCH = amd64
endif

# Symlink into GOPATH
CURRENT_DIR=$(shell pwd)
BUILD_DIR=${CURRENT_DIR}/

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'"

# Build the project
all: clean default

default:
	cd ${BUILD_DIR}; \
	GOOS=$(strip ${GOOS}) GOARCH=$(strip ${GOARCH}) go build ${LDFLAGS} -o ${BINARY}-$(strip ${GOOS})-$(strip ${GOARCH}) . ; \
	cd - >/dev/null

clean:
	-rm -f ${BINARY}-*

.PHONY: default clean

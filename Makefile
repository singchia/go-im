BINARY = go-im
GOARCH = amd64

VERSION?=0.1

# Symlink into GOPATH
CURRENT_DIR=$(shell pwd)
BUILD_DIR=${CURRENT_DIR}/

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'"

# Build the project
all: clean linux darwin windows

linux:
	cd ${BUILD_DIR}; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} . ; \
	cd - >/dev/null

darwin:
	cd ${BUILD_DIR}; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH} . ; \
	cd - >/dev/null

windows:
	cd ${BUILD_DIR}; \
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-windows-${GOARCH}.exe . ; \
	cd - >/dev/null


clean:
	-rm -f ${BINARY}-*

.PHONY: linux darwin windows clean

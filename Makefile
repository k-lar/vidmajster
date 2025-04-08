PREFIX ?= /usr
BINARY_NAME=vidmajster

all: build

build:
	go build -o ${BINARY_NAME} .

build-macos-old:
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY_NAME}-amd64-macos .

build-macos:
	GOOS=darwin GOARCH=arm64 go build -o ${BINARY_NAME}-macos .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o ${BINARY_NAME}.exe .

build-all:
	GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME} .
	GOOS=darwin GOARCH=amd64 go build -o ${BINARY_NAME}-macos-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o ${BINARY_NAME}-macos .
	GOOS=windows GOARCH=amd64 go build -o ${BINARY_NAME}.exe .

run:
	go build -o ${BINARY_NAME} .
	./${BINARY_NAME}

install:
	@# Create the bin directory if it doesn't exist (Mac doesn't support install -D)
	@if [ ! -d $(DESTDIR)$(PREFIX)/bin ]; then \
		mkdir -p $(DESTDIR)$(PREFIX)/bin; \
	fi

	@install -m755 ${BINARY_NAME} $(DESTDIR)$(PREFIX)/bin/${BINARY_NAME}

uninstall:
	@rm -f $(DESTDIR)$(PREFIX)/bin/${BINARY_NAME}

clean:
	go clean
	rm ${BINARY_NAME}


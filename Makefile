.PHONY: build test vet lint fixtures clean

build:
	go build -o bin/ariadne ./cmd/ariadne

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

# Compiles the hand-written Java fixture under testdata/fixture for two
# bytecode versions and commits the .class/.jar output, so `go test` never
# needs a JDK. Requires JDK 11+ (release 8 target) and JDK 21+ (release 21
# target) on PATH or via JAVAC8/JAVAC21.
JAVAC8 ?= javac
JAVAC21 ?= javac
fixtures:
	./scripts/build-fixtures.sh "$(JAVAC8)" "$(JAVAC21)"

clean:
	rm -rf bin

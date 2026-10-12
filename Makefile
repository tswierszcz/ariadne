.PHONY: build test vet lint fixtures clean e2e-llm

build:
	go build -o bin/ariadne ./cmd/ariadne

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

# One real LLM call with the real prompts (needs the ARIADNE_LLM_* variables
# in .env; costs a fraction of a cent). Not part of `make test`.
e2e-llm:
	set -a; . ./.env; set +a; ARIADNE_LLM_E2E=1 go test ./internal/vulnmap -run TestLLM_realEndpoint -v -count=1

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

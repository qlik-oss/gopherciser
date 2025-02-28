GO := go
PKG_NAME := gopherciser
BIN_NAME := gopherciser
PREFIX := .
TEST_REPORTS := $(PREFIX)/.cover
BIN := build
OSFLAG :=
ifeq ($(OS),Windows_NT)
	OSFLAG += $(BIN)/$(BIN_NAME).exe
else
	UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	OSFLAG += ./$(BIN)/$(BIN_NAME)
endif
ifeq ($(UNAME_S),Darwin)
	OSFLAG += ./$(BIN)/$(BIN_NAME)_osx
endif
endif

.PHONY: clean build lint test alltests initwiki genwiki build-docker attribution test-cover

# Compile Go packages
build: clean
	./scripts/build.sh $(PREFIX) $(BIN) $(BIN_NAME)
	$(OSFLAG) version

# Lint the code
lint:
	./scripts/lint.sh 

# Minimum level of linting for PR's to be accepted
lint-min:
	./scripts/lint.sh MIN

# Clear and clean folder
clean:
	./scripts/clean.sh $(PREFIX) $(BIN) $(TEST_REPORTS)

# Build documentation
docbuild:
	cp ./docs/Readme.txt ./build/Readme.txt

# Build for current platform only, does not clean, does not do full rebuild, does not create folders, does not set the version nor strip DWARF tables etc.
# Meant to be used during development only
quickbuild:
	go build -mod=readonly -o $(OSFLAG)

# Run standard tests
test:
	go test -race -mod=readonly ./...

# Run all tests with verbose output
alltests:
	set -eu
	go test -race -mod=readonly -v ./... -count=1 -cover -coverprofile=coverage.csv
	go tool cover -html=coverage.csv -o coverage.html

# Run quickbuild test and linting. Good to run e.g. before pushing to remote
verify: quickbuild test lint-min

# init submodule and get latest version
initwiki:
	git submodule update --init --recursive --remote
	git submodule foreach --recursive git checkout master
	git submodule foreach --recursive git pull

# generate config and action documenation
genwiki: initwiki
	set -e
	go generate
	go run ./generatedocs/cmd/generatemarkdown $(PARAM) --wiki ./gopherciser.wiki

# build docker image
build-docker: 
	DOCKERBUILD=y ./scripts/build.sh $(PREFIX) $(BIN) $(BIN_NAME)

# Generate licences.txt
attribution:
	go install github.com/google/go-licenses@latest
	./scripts/createattribution.sh

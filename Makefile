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

.PHONY: clean build unit-test-cover unit-test-cover-ext codeclimate lint test alltests

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

# Unit test and coverage
unit-test-cover: clean
	./scripts/unit-test-cover.sh $(TEST_REPORTS) $(BIN_NAME)

# Unit test and cover extended, creates html file for viewing: ${TEST_REPORTS}/c.html
unit-test-cover-ext: unit-test-cover
	./scripts/unit-test-cover-ext.sh $(TEST_REPORTS)

# Clear and clean folder
clean:
	./scripts/clean.sh $(PREFIX) $(BIN) $(TEST_REPORTS)

# Build documentation
docbuild:
	cp ./docs/Readme.txt ./build/Readme.txt

# Build for current platform only, does not clean, does not do full rebuild, does not create folders, does not set the version nor strip DWARF tables etc.
# Meant to be used during development only
quickbuild:
	GO111MODULE=on go build -mod=readonly -o $(OSFLAG)

# Run standard tests
test:
	GO111MODULE=on go test -race -mod=readonly ./...

# Run all tests with verbose output
alltests:
	GO111MODULE=on go test -race -mod=readonly -v ./... -count=1

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

.PHONY: test cover all lint tools
BUILD_DIR ?= $(abspath ./artifacts)
DEBUG ?= 0

ifeq ($(DEBUG),1)
GO_TEST := @go test -v
else
GO_TEST := @go test
endif

fmtcheck = @if goimports -l $(1) | read var; then echo "goimports check failed for $(1):\n `goimports -d $(1)`"; exit 1; fi

vet:
	@echo "+ Vet"
	@go vet .

lint:
	@echo "+ Linting package"
	@golint .
	$(call fmtcheck, .)

get-deps:
	@echo "+ Downloading dependencies"
	@go get ./...
	@go get -t ./...

test:
	@echo "+ Testing package"
	$(GO_TEST) .

cover: test
	@echo "+ Tests Coverage"
	@mkdir -p $(BUILD_DIR)
	@touch $(BUILD_DIR)/cover.out
	@go test -coverprofile=$(BUILD_DIR)/cover.out
	@go tool cover -html=$(BUILD_DIR)/cover.out -o=$(BUILD_DIR)/coverage.html

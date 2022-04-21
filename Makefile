mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

.PHONY: build
build:
	./provision/before-commit.sh ci
	./function/before-commit.sh ci

.PHONY: ci-pr
ci-pr: build

.PHONY: ci-main
ci-main: build

.PHONY: ci-release
ci-release: build

.PHONY: lint-function
lint-function:
	./hack/verify-lint.sh $(mkfile_dir)/function

.PHONY: lint-provision
lint-provision:
	./hack/verify-lint.sh $(mkfile_dir)/provision

.PHONY: lint
lint: lint-function lint-provision

.PHONY: test-provision
test-provision:
	@cd provision; \
	echo "Running tests for provision"; \
	go test -coverprofile=cover.out ./... ;\
	echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')" ;\
	rm cover.out ; \
	cd ..;

.PHONY: test
test: test-provision

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

.PHONY: build
build:
	./provision/before-commit.sh ci
	./install/before-commit.sh ci
	./parallel-install/before-commit.sh ci
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

.PHONY: lint-install
lint-install:
	./hack/verify-lint.sh $(mkfile_dir)/install

.PHONY: lint-parallel-install
lint-parallel-install:
	./hack/verify-lint.sh $(mkfile_dir)/parallel-install

.PHONY: lint
lint: lint-function lint-provision lint-install lint-parallel-install

.PHONY: test-provision
test-provision:
	@cd provision; \
	echo "Running tests for provision"; \
	go test -coverprofile=cover.out ./... ;\
	echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')" ;\
	rm cover.out ; \
	cd ..;

.PHONY: test-install
test-install:
	@cd install; \
	echo "Running tests for install"; \
	go test -coverprofile=cover.out ./... ;\
	echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')" ;\
	rm cover.out ; \
	cd ..;

.PHONY: test-parallel-install
test-parallel-install:
	@cd parallel-install; \
	echo "Running tests for parallel-install"; \
	go test -coverprofile=cover.out ./... ;\
	echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')" ;\
	rm cover.out ; \
	cd ..;

.PHONY: test
test: test-provision test-install test-parallel-install
test123

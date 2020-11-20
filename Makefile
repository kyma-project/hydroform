.PHONY: build
build:
	./provision/before-commit.sh ci
	./install/before-commit.sh ci
	./parallel-install/before-commit.sh ci

.PHONY: ci-pr
ci-pr: build

.PHONY: ci-master
ci-master: build

.PHONY: ci-release
ci-release: build

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
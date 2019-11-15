.PHONY: build
build:
	./before-commit.sh ci
	# TODO: we should provide separate Prow jobs for provisioning and installation packages after the repository will be adjusted
	$(MAKE) -C install build

.PHONY: ci-pr
ci-pr: build

.PHONY: ci-master
ci-master: build

.PHONY: ci-release
ci-release: build

.PHONY: test
test:
	go test -coverprofile=cover.out ./...
	@echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')"
	@rm cover.out

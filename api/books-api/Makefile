dependencies:
	go get .

run:
	go run main.go

GOTESTSUM := go run gotest.tools/gotestsum@v1.12.0
TEST_PACKAGES := $(shell go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)

.PHONY: test test-verbose

test:
	@test "$(APP_ENV)" = test || (echo "APP_ENV=test is required"; exit 1)
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required"; exit 1)
	@case "$(DATABASE_URL)" in */writerslife_test\?*|*/writerslife_test) ;; *) echo "DATABASE_URL must target writerslife_test"; exit 1;; esac
	@go test ./... >/dev/null
	@$(GOTESTSUM) --format testname $(TEST_PACKAGES)

test-verbose:
	@test "$(APP_ENV)" = test || (echo "APP_ENV=test is required"; exit 1)
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required"; exit 1)
	@case "$(DATABASE_URL)" in */writerslife_test\?*|*/writerslife_test) ;; *) echo "DATABASE_URL must target writerslife_test"; exit 1;; esac
	@go test ./... >/dev/null
	@$(GOTESTSUM) --format standard-verbose $(TEST_PACKAGES)


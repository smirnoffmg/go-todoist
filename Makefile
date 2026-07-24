.PHONY: fmt lint test test-integration tidy ci

fmt:
	golangci-lint fmt

lint:
	golangci-lint run

test:
	go test -race -cover ./...

# Requires TODOIST_TOKEN. Hits the real Todoist API; creates and deletes its own
# scratch data.
test-integration:
	go test -tags=integration -run Integration -v ./...

tidy:
	go mod tidy

ci: lint test

DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
PACKAGES = $(shell go list ./...)
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
		          -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

all: deps format
		@mkdir -p dist/
		@bash --norc -i ./scripts/build.sh

deps:
		@echo "--> Installing build dependencies"
		@go get -d -v ./... $(DEPS)

updatedeps: deps
		@echo "--> Updating build dependencies"
		@go get -d -f -u ./... $(DEPS)

format: deps
		@echo "--> Running go fmt"
		@go fmt $(PACKAGES)

vet:
		@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
			go get golang.org/x/tools/cmd/vet; \
		fi
		@echo "--> Running go tool vet $(VETARGS) ."
		@go tool vet $(VETARGS) . ; if [ $$? -eq 1 ]; then \
			echo ""; \
			echo "Found suspicious constructs. Please check."; \
		fi

.PHONY: all deps vet

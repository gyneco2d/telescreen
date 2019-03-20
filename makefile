BUILD_DIR	:= $(GOPATH)/bin
NAME     	:= telescreen

.PHONY: build
build:
		@go build -o $(BUILD_DIR)/$(NAME) ./cmd/$(NAME)

.PHONY: run
run:
		@$(BUILD_DIR)/$(NAME)

.PHONY: fmt
fmt:
		@go fmt ./cmd/$(NAME)

.PHONY: test
test:	fmt build run

.PHONY: clean
clean:
		rm -rf $(GOPATH)/bin/$(NAME)

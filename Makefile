BUILD_DIR = bin
BIN_NAME = audiodrive
DATA_DIR=/var/lib/$(BIN_NAME)

.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	go build -o ./$(BUILD_DIR)/$(BIN_NAME) ./main.go

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: install
install: build
	echo "Installing $(BIN_NAME) ..."; \
	sudo install -m 755 $(BIN_NAME) /usr/local/bin/$(BIN_NAME); \
	echo "Done."
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
	@echo "Creating data directory if it doesn't exist..."
	@sudo mkdir -p $(DATA_DIR)
	@sudo chown $(USER):$(USER) $(DATA_DIR)
	@echo "Installing binary..."
	@sudo install -m 755 $(BUILD_DIR)/$(BIN_NAME) $(BIN_INSTALL_PATH)
	@echo "Install complete."
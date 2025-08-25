BUILD_DIR = bin
BIN_NAME = audiodrive
DATA_DIR=/var/lib/$(BIN_NAME)
INSTALL_DIR =/usr/local/bin

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
	@sudo install -m 755 $(BUILD_DIR)/$(BIN_NAME) $(INSTALL_DIR)/$(BIN_NAME)
	@echo "Install complete."
	
.PHONY: uninstall
uninstall:
    @echo -n "Are you sure? [Y/n] " && read ans && [ $${ans:-Y} != Y ] && echo "Aborted" && exit 1
    @echo "Removing data directory..."
    @sudo rm -rf $(DATA_DIR)
    @echo "Removing binary..."
    @sudo rm $(INSTALL_DIR)/$(BIN_NAME)
    @echo "Uninstalled."
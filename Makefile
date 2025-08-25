BUILD_DIR = bin

.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	go build -o ./$(BUILD_DIR)/audiodrive-server ./audiodrive.go

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

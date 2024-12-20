DEST_DIR=target/sbot
SRC_BIN=bin
CHAT_TEMPLATES_SRC=chat_templates
COMMAND_HISTORY=sbot_command_history.txt
SETTING_SRC=setting.json
ENV.EXAMPLE=.env.example
APP_VERSION := $(shell cat VERSION)
ENV=.env

BINARYS=../$(DEST_DIR)/bin/.
APP_NAME=sbot

# Phony targets (not real files)
.PHONY: all build clean

# Default target
all: build

# Build the Go project
build: clean
	@echo "Started build."
	@mkdir -p $(DEST_DIR)
	@cp -r $(CHAT_TEMPLATES_SRC) $(DEST_DIR)
	@cp $(SETTING_SRC) $(DEST_DIR)
	@touch $(DEST_DIR)/$(COMMAND_HISTORY)
	@cp $(ENV.EXAMPLE) $(DEST_DIR)/$(ENV)
	@cd $(SRC_BIN) && go build -o $(BINARYS)/$(APP_NAME)
	@find target -type f -exec chmod 700 {} + && find target -type d -exec chmod 700 {} +
	@echo "build complete."

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARYS)
	@rm -rf $(DEST_DIR)/$(CHAT_TEMPLATES_SRC)
	@rm -f $(DEST_DIR)/$(SETTING_SRC)
	@rm -f $(DEST_DIR)/$(COMMAND_HISTORY)
	@rm -f $(DEST_DIR)/$(ENV)
	@echo "Cleanup complete."

release: clean build

	@cd target; zip -r sbot.zip sbot && mv sbot.zip /tmp/sbot_$(APP_VERSION).zip;
	@echo "release completed"


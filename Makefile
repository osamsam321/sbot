DEST_DIR=target/.sbot
SRC_BIN=bin
OPENAI_KEY=open_api_key.txt
PROMPTS_SRC=prompts
COMMAND_HISTORY=sbot_command_history.txt
SETTING_SRC=setting.json

BINARY=../$(DEST_DIR)/bin/sbot

# Phony targets (not real files)
.PHONY: all build clean

# Default target
all: build

# Build the Go project
build:
	@mkdir -p $(DEST_DIR)
	cp -r $(PROMPTS_SRC) $(DEST_DIR)
	cp $(SETTING_SRC) $(DEST_DIR)
	touch $(DEST_DIR)/$(OPENAI_KEY)
	touch $(DEST_DIR)/$(COMMAND_HISTORY)
	cd $(SRC_BIN) && go build -o $(BINARY)

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY)
	rm -rf $(DEST_DIR)/$(PROMPTS_SRC)
	rm -f $(DEST_DIR)/$(SETTING_SRC)
	rm -f $(DEST_DIR)/$(OPENAI_KEY)
	rm -f $(DEST_DIR)/$(COMMAND_HISTORY)
	@echo "Cleanup complete."

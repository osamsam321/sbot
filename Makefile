# Define the binary name
BINARY=sbot

# Define source and destination directories
DEST_DIR= /target/sbot

# Phony targets (not real files)
#.PHONY: all build move clean

# Default target
all: build move

# Build the Go project
build:
	go build -o $(BINARY)

# Move files after a successful build
move: build
	@echo "Moving files..."
	mv * $(DEST_DIR)/
	@echo "Files moved successfully."

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY)
	@echo "Cleanup complete."


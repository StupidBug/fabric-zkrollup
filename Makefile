TARGET ?= 

ifeq ($(TARGET), storage)
TARGET_FILE := ./cmd/storage/main.go
else ifeq ($(TARGET), transfer)
TARGET_FILE := ./cmd/transfer/main.go
endif

all: run

build:
	@echo "Building target: $(TARGET)"
	go build -gcflags="all=-N -l" -o bin/$(TARGET) $(TARGET_FILE)

run: clean build
	@echo "Running target: $(TARGET)"
	./bin/$(TARGET)

clean:
	@echo "Cleaning up..."
	@echo "deleting old wallet"
	@rm -rf ./wallet
	@rm -rf pkg/chaincode/wallet
	@rm -rf ./keystore
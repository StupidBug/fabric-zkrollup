TARGET ?= 

define CLEAN 
	@echo "deleting old wallet"
	@rm -rf ./wallet
	@rm -rf pkg/chaincode/wallet
	@rm -rf ./keystore
endef

ifeq ($(TARGET), storage)
define RUN_COMMAND
	go run ./cmd/storage/main.go
endef
else ifeq ($(TARGET), transfer)
define RUN_COMMAND
	go run ./cmd/transfer/main.go
endef
endif

run:
	@echo "Running target: $(TARGET)"
	${CLEAN}
	$(RUN_COMMAND)

clean:
	@echo "Cleaning up..."
	${CLEAN}
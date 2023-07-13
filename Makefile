DOCKER := $(shell which docker)

protoVer=0.13.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --user 0 --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files"
	@mkdir -p ./proto/ibc/lightclients/ethereum/v1
	@$(protoImage) sh ./scripts/protocgen.sh

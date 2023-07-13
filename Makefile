DOCKER := $(shell which docker)

protoVer=0.13.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --user 0 --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

.PHONY: proto-update-deps proto-gen
proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --user 0 --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update

proto-gen:
	@echo "Generating Protobuf files"
	@mkdir -p ./proto/ibc/lightclients/ethereum/v1
	@$(protoImage) sh ./scripts/protocgen.sh

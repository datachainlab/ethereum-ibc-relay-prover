ETHEREUM_IBC_PROTO ?= ./ethereum-ibc-rs/proto/definitions

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
	@sed "s/option\sgo_package.*;/option\ go_package\ =\ \"github.com\/datachainlab\/ethereum-lcp\/go\/light-clients\/ethereum\/types\";/g"\
		$(ETHEREUM_IBC_PROTO)/ibc/lightclients/ethereum/v1/ethereum.proto > ./proto/ibc/lightclients/ethereum/v1/ethereum.proto
	@$(protoImage) sh ./scripts/protocgen.sh

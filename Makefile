ETHEREUM_IBC_PROTO ?= ./ethereum-ibc-rs/proto/definitions

.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files"
	@rm -rf ./proto/ibc && cp -a $(ETHEREUM_IBC_PROTO)/ibc ./proto/
	docker run -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen:v0.3 sh ./scripts/protocgen.sh

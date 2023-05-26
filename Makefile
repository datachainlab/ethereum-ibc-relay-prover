ETHEREUM_IBC_PROTO ?= ./ethereum-ibc-rs/proto/definitions

.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files"
	@mkdir -p ./proto/ibc/lightclients/ethereum/v1
	@sed "s/option\sgo_package.*;/option\ go_package\ =\ \"github.com\/datachainlab\/ethereum-lcp\/go\/light-clients\/ethereum\/types\";/g"\
		$(ETHEREUM_IBC_PROTO)/ibc/lightclients/ethereum/v1/ethereum.proto > ./proto/ibc/lightclients/ethereum/v1/ethereum.proto
	@docker run -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen:v0.3 sh ./scripts/protocgen.sh

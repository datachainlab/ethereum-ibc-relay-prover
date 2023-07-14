#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto

buf generate --template buf.gen.gogo.yaml

cd ..

# move proto files to the right places
cp -r github.com/datachainlab/ethereum-ibc-relay-prover/* ./
rm -rf github.com

go mod tidy

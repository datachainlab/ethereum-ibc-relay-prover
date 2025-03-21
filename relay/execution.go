package relay

import (
	"context"
	"encoding/hex"
	"math/big"

	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func (pr *Prover) buildAccountUpdate(ctx context.Context, blockNumber uint64) (*lctypes.AccountUpdate, error) {
	proof, err := pr.executionClient.GetProof(
		ctx,
		pr.chain.Config().IBCAddress(),
		nil,
		big.NewInt(int64(blockNumber)),
	)
	if err != nil {
		return nil, err
	}
	pr.GetLogger().Info("buildAccountUpdate: get proof", "block_number", blockNumber, "ibc_address", pr.chain.Config().IBCAddress().String(), "account_proof", hex.EncodeToString(proof.AccountProofRLP), "storage_hash", hex.EncodeToString(proof.StorageHash[:]))
	return &lctypes.AccountUpdate{
		AccountProof:       proof.AccountProofRLP,
		AccountStorageRoot: proof.StorageHash[:],
	}, nil
}

func (pr *Prover) buildStateProof(ctx context.Context, path []byte, height int64) ([]byte, error) {
	// calculate slot for commitment
	storageKey := crypto.Keccak256Hash(append(
		crypto.Keccak256Hash(path).Bytes(),
		IBCCommitmentsSlot.Bytes()...,
	))
	storageKeyHex, err := storageKey.MarshalText()
	if err != nil {
		return nil, err
	}

	// call eth_getProof
	stateProof, err := pr.executionClient.GetProof(
		ctx,
		pr.chain.Config().IBCAddress(),
		[][]byte{storageKeyHex},
		big.NewInt(height),
	)
	if err != nil {
		return nil, err
	}
	return stateProof.StorageProofRLP[0], nil
}

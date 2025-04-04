package relay

import (
	"context"
	"encoding/hex"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func (pr *Prover) buildAccountUpdate(ctx context.Context, blockNumber uint64) (*lctypes.AccountUpdate, error) {
	result, err := BuildAccountUpdate(ctx, pr.executionClient, pr.chain.Config().IBCAddress(), blockNumber)
	if err != nil {
		return nil, err
	}
	pr.GetLogger().Info("buildAccountUpdate: get proof", "block_number", blockNumber, "ibc_address", pr.chain.Config().IBCAddress().String(), "account_proof", hex.EncodeToString(result.AccountProof), "storage_hash", hex.EncodeToString(result.AccountStorageRoot))
	return result, nil
}

func (pr *Prover) buildStateProof(ctx context.Context, path []byte, height int64) ([]byte, error) {
	return BuildStateProof(ctx, pr.executionClient, pr.chain.Config().IBCAddress(), path, height)
}

func BuildAccountUpdate(ctx context.Context, executionClient *client.ETHClient, address common.Address, blockNumber uint64) (*lctypes.AccountUpdate, error) {
	proof, err := executionClient.GetProof(
		ctx,
		address,
		nil,
		big.NewInt(int64(blockNumber)),
	)
	if err != nil {
		return nil, err
	}
	return &lctypes.AccountUpdate{
		AccountProof:       proof.AccountProofRLP,
		AccountStorageRoot: proof.StorageHash[:],
	}, nil
}

func BuildStateProof(ctx context.Context, executionClient *client.ETHClient, address common.Address, path []byte, height int64) ([]byte, error) {
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
	stateProof, err := executionClient.GetProof(
		ctx,
		address,
		[][]byte{storageKeyHex},
		big.NewInt(height),
	)
	if err != nil {
		return nil, err
	}
	return stateProof.StorageProofRLP[0], nil
}

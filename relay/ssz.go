package relay

import (
	"bytes"
	"encoding/binary"

	fastssz "github.com/prysmaticlabs/fastssz"
	"github.com/prysmaticlabs/prysm/v4/encoding/ssz"
	enginev1 "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
)

func generate_merkle_proof(leaves [][]byte, generalizedIndex int) ([][]byte, error) {
	node, err := fastssz.TreeFromChunks(leaves)
	if err != nil {
		return nil, err
	}
	// NOTE: seems that fastssz sets root as index=1, so the index is off by one from the ethereum consensus-spes
	// https://github.com/ethereum/consensus-specs/blob/c46c3945fd7fbd1226ece1f8d684c4b724b7bdab/ssz/merkle-proofs.md#generalized-merkle-tree-index
	proof, err := node.Prove(generalizedIndex + 1)
	if err != nil {
		return nil, err
	}
	return proof.Hashes, nil
}

func generate_execution_payload_proof(header *enginev1.ExecutionPayloadHeaderCapella, generalizedIndex int) ([][]byte, error) {
	var zero [32]byte
	return generate_merkle_proof([][]byte{
		header.ParentHash,
		ssz_bytes(header.FeeRecipient),
		header.StateRoot,
		header.ReceiptsRoot,
		ssz_bytes(header.LogsBloom),
		header.PrevRandao,
		ssz_uint64(header.BlockNumber),
		ssz_uint64(header.GasLimit),
		ssz_uint64(header.GasUsed),
		ssz_uint64(header.Timestamp),
		extraDataRootBytes(header.ExtraData),
		header.BaseFeePerGas,
		header.BlockHash,
		header.TransactionsRoot,
		header.WithdrawalsRoot,
		zero[:],
	}, generalizedIndex)
}

func ssz_bytes(bz []byte) []byte {
	hh := fastssz.NewHasher()
	hh.PutBytes(bz)
	root, err := hh.HashRoot()
	if err != nil {
		panic(err)
	}
	return root[:]
}

func ssz_uint64(v uint64) []byte {
	hh := fastssz.NewHasher()
	hh.PutUint64(v)
	root, err := hh.HashRoot()
	if err != nil {
		panic(err)
	}
	return root[:]
}

func extraDataRootBytes(tx []byte) []byte {
	bz, err := extraDataRoot(tx)
	if err != nil {
		panic(err)
	}
	return bz[:]
}

func extraDataRoot(bz []byte) ([32]byte, error) {
	chunkedRoots, err := ssz.PackByChunk([][]byte{bz})
	if err != nil {
		return [32]byte{}, err
	}
	const MAX_EXTRA_DATA_BYTES = 32

	maxLength := (MAX_EXTRA_DATA_BYTES + 31) / 32
	bytesRoot, err := ssz.BitwiseMerkleize(chunkedRoots, uint64(len(chunkedRoots)), uint64(maxLength))
	if err != nil {
		return [32]byte{}, err
	}
	bytesRootBuf := new(bytes.Buffer)
	if err := binary.Write(bytesRootBuf, binary.LittleEndian, uint64(len(bz))); err != nil {
		return [32]byte{}, err
	}
	bytesRootBufRoot := make([]byte, 32)
	copy(bytesRootBufRoot, bytesRootBuf.Bytes())
	return ssz.MixInLength(bytesRoot, bytesRootBufRoot), nil
}

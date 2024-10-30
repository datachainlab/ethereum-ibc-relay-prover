package beacon

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	enginev1 "github.com/prysmaticlabs/prysm/v5/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
)

// Beacon Types

type H256 = [32]byte
type BeaconBlockHeader = ethpb.BeaconBlockHeader
type ExecutionPayloadHeader = enginev1.ExecutionPayloadHeaderDeneb

type LightClientUpdate struct {
	AttestedHeader          BeaconBlockHeader
	NextSyncCommittee       SyncCommittee
	NextSyncCommitteeBranch []H256
	FinalizedHeader         BeaconBlockHeader
	FinalizedHeaderBranch   []H256
	SyncAggregate           SyncAggregate
	SignatureSlot           uint64
}

type Genesis struct {
	GenesisTimeSeconds    uint64
	GenesisValidatorsRoot [32]byte
	GenesisForkVersion    [4]byte
}

type StateFinalityCheckpoints struct {
	PreviousJustified Checkpoint
	CurrentJustified  Checkpoint
	Finalized         Checkpoint
}

type Checkpoint struct {
	Epoch uint64
	Root  [32]byte
}

func (lcu *LightClientUpdateData) ToProto() *lctypes.ConsensusUpdate {
	executionRoot, err := lcu.FinalizedHeader.Execution.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	return &lctypes.ConsensusUpdate{
		AttestedHeader:           lcu.AttestedHeader.ToProto(),
		NextSyncCommittee:        lcu.NextSyncCommittee.ToProto(),
		NextSyncCommitteeBranch:  convertToBytesSlice(lcu.NextSyncCommitteeBranch),
		FinalizedHeader:          lcu.FinalizedHeader.ToProto(),
		FinalizedHeaderBranch:    convertToBytesSlice(lcu.FinalityBranch),
		FinalizedExecutionRoot:   executionRoot[:],
		FinalizedExecutionBranch: convertToBytesSlice(lcu.FinalizedHeader.ExecutionBranch),
		SyncAggregate:            lcu.SyncAggregate.ToProto(),
		SignatureSlot:            uint64(lcu.SignatureSlot),
	}
}

func (sc *SyncCommittee) ToProto() *lctypes.SyncCommittee {
	return &lctypes.SyncCommittee{
		Pubkeys:         convertToBytesSlice(sc.PubKeys),
		AggregatePubkey: []byte(sc.AggregatePubKey),
	}
}

func (lcf *LightClientFinalityUpdate) ToProto() *lctypes.ConsensusUpdate {
	executionRoot, err := lcf.FinalizedHeader.Execution.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	return &lctypes.ConsensusUpdate{
		AttestedHeader:           lcf.AttestedHeader.ToProto(),
		NextSyncCommittee:        nil,
		NextSyncCommitteeBranch:  nil,
		FinalizedHeader:          lcf.FinalizedHeader.ToProto(),
		FinalizedHeaderBranch:    convertToBytesSlice(lcf.FinalityBranch),
		FinalizedExecutionRoot:   executionRoot[:],
		FinalizedExecutionBranch: convertToBytesSlice(lcf.FinalizedHeader.ExecutionBranch),
		SyncAggregate:            lcf.SyncAggregate.ToProto(),
		SignatureSlot:            uint64(lcf.SignatureSlot),
	}
}

func (h *LightClientHeader) ToProto() *lctypes.BeaconBlockHeader {
	return &lctypes.BeaconBlockHeader{
		Slot:          uint64(h.Beacon.Slot),
		ProposerIndex: uint64(h.Beacon.ProposerIndex),
		ParentRoot:    h.Beacon.ParentRoot,
		StateRoot:     h.Beacon.StateRoot,
		BodyRoot:      h.Beacon.BodyRoot,
	}
}

func (c *SyncAggregate) ToProto() *lctypes.SyncAggregate {
	return &lctypes.SyncAggregate{
		SyncCommitteeBits:      []byte(c.SyncCommitteeBits),
		SyncCommitteeSignature: []byte(c.SyncCommitteeSignature),
	}
}

func convertToBytesSlice(bzs []hexutil.Bytes) [][]byte {
	var s [][]byte
	for _, bz := range bzs {
		s = append(s, bz)
	}
	return s
}

func ToGenesis(res GenesisResponse) (*Genesis, error) {
	var (
		root        [32]byte
		forkVersion [4]byte
	)

	rootBytes, err := hex.DecodeString(strings.TrimPrefix(res.Data.GenesisValidatorsRoot, "0x"))
	if err != nil {
		return nil, err
	}
	if l := len(rootBytes); l != 32 {
		return nil, fmt.Errorf("invalid root bytes length: expected=32 actual=%v", l)
	}
	copy(root[:], rootBytes)

	gt, err := strconv.Atoi(res.Data.GenesisTime)
	if err != nil {
		return nil, err
	}
	forkVersionBytes, err := hex.DecodeString(strings.TrimPrefix(res.Data.GenesisForkVersion, "0x"))
	if err != nil {
		return nil, err
	}
	if l := len(forkVersionBytes); l != 4 {
		return nil, fmt.Errorf("invalid fork_version bytes length: expected=4 actual=%v", l)
	}
	copy(forkVersion[:], forkVersionBytes)

	return &Genesis{
		GenesisTimeSeconds:    uint64(gt),
		GenesisValidatorsRoot: root,
		GenesisForkVersion:    forkVersion,
	}, nil
}

func ToStateFinalityCheckpoints(res StateFinalityCheckpointResponse) (*StateFinalityCheckpoints, error) {
	var checkpoints StateFinalityCheckpoints
	var err error

	checkpoints.PreviousJustified.Epoch, err = epochStringToUint64(res.Data.PreviousJustified.Epoch)
	if err != nil {
		return nil, err
	}
	checkpoints.PreviousJustified.Root, err = rootHexStringToBytes(res.Data.PreviousJustified.Root)
	if err != nil {
		return nil, err
	}

	checkpoints.CurrentJustified.Epoch, err = epochStringToUint64(res.Data.CurrentJustified.Epoch)
	if err != nil {
		return nil, err
	}
	checkpoints.CurrentJustified.Root, err = rootHexStringToBytes(res.Data.CurrentJustified.Root)
	if err != nil {
		return nil, err
	}

	checkpoints.Finalized.Epoch, err = epochStringToUint64(res.Data.Finalized.Epoch)
	if err != nil {
		return nil, err
	}
	checkpoints.Finalized.Root, err = rootHexStringToBytes(res.Data.Finalized.Root)
	if err != nil {
		return nil, err
	}

	return &checkpoints, nil
}

func epochStringToUint64(s string) (uint64, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return uint64(v), nil
}

func rootHexStringToBytes(s string) ([32]byte, error) {
	var root [32]byte
	bz, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return root, nil
	}
	copy(root[:], bz)
	return root, nil
}

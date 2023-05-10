package types

import (
	commitmenttypes "github.com/cosmos/ibc-go/v4/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
)

var _ exported.ConsensusState = (*ConsensusState)(nil)

func (cs *ConsensusState) ClientType() string {
	return ClientType
}

// GetRoot returns the commitment root of the consensus state,
// which is used for key-value pair verification.
func (cs *ConsensusState) GetRoot() exported.Root {
	return commitmenttypes.NewMerkleRoot(cs.StorageRoot)
}

func (cs *ConsensusState) ValidateBasic() error {
	return nil
}

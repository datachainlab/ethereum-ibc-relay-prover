package types

import "fmt"

func (u *ConsensusUpdate) ValidateBasic() error {
	if u == nil {
		return fmt.Errorf("light client update cannot be nil")
	}
	if u.AttestedHeader == nil {
		return fmt.Errorf("attested header cannot be nil")
	}
	if u.FinalizedHeader == nil {
		return fmt.Errorf("finalized header cannot be nil")
	}
	if u.FinalizedHeaderBranch == nil {
		return fmt.Errorf("finalized header branch cannot be nil")
	}
	if u.FinalizedExecutionRoot == nil {
		return fmt.Errorf("finalized execution root cannot be nil")
	}
	if u.FinalizedExecutionBranch == nil {
		return fmt.Errorf("finalized execution branch cannot be nil")
	}
	if u.SyncAggregate == nil {
		return fmt.Errorf("sync aggregate cannot be nil")
	}
	if u.SignatureSlot == 0 {
		return fmt.Errorf("signature slot cannot be zero")
	}
	return nil
}

func (u *AccountUpdate) ValidateBasic() error {
	if u == nil {
		return fmt.Errorf("account update cannot be nil")
	}
	if u.AccountProof == nil {
		return fmt.Errorf("account proof cannot be nil")
	}
	if u.AccountStorageRoot == nil {
		return fmt.Errorf("account storage root cannot be nil")
	}
	return nil
}

func (u *ExecutionUpdate) ValidateBasic() error {
	if u == nil {
		return fmt.Errorf("execution update cannot be nil")
	}
	if u.StateRoot == nil {
		return fmt.Errorf("state root cannot be nil")
	}
	if u.StateRootBranch == nil {
		return fmt.Errorf("state root branch cannot be nil")
	}
	if u.BlockNumber == 0 {
		return fmt.Errorf("block number cannot be zero")
	}
	if u.BlockNumberBranch == nil {
		return fmt.Errorf("block number branch cannot be nil")
	}
	return nil
}

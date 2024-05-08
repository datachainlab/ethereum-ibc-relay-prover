package types

import (
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
)

var _ exported.ClientMessage = (*Header)(nil)

func (h *Header) ClientType() string {
	return ClientType
}

func (h *Header) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, h.ExecutionUpdate.BlockNumber)
}

func (h *Header) ValidateBasic() error {
	if err := h.ConsensusUpdate.ValidateBasic(); err != nil {
		return err
	}
	if err := h.ExecutionUpdate.ValidateBasic(); err != nil {
		return err
	}
	if err := h.AccountUpdate.ValidateBasic(); err != nil {
		return err
	}
	if h.Timestamp == 0 {
		return fmt.Errorf("timestamp cannot be zero")
	}
	return nil
}

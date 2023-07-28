package types

import (
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

var _ exported.ClientMessage = (*Header)(nil)

func (h *Header) ClientType() string {
	return ClientType
}

func (h *Header) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, h.ExecutionUpdate.BlockNumber)
}

func (h *Header) ValidateBasic() error {
	return nil
}

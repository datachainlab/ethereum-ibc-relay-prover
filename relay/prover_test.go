package relay

import (
	"fmt"
	"testing"

	"github.com/datachainlab/ethereum-ibc-relay-prover/beacon"
)

func TestBootstrap(t *testing.T) {
	res := beacon.LightClientBootstrapResponse{}
	proto := res.Data.CurrentSyncCommittee.ToProto()
	fmt.Println(proto.AggregatePubkey)
}

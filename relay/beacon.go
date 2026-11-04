package relay

import (
	"context"

	"github.com/datachainlab/ethereum-light-client-types/prover/beacon"
	lcrelay "github.com/datachainlab/ethereum-light-client-types/prover/relay"
	lctypes "github.com/datachainlab/ethereum-light-client-types/prover/types"
)

func (pr *Prover) secondsPerSlot() uint64 {
	return lcrelay.SecondsPerSlot(pr.config.Network)
}

func (pr *Prover) slotsPerEpoch() uint64 {
	return lcrelay.SlotsPerEpoch(pr.config.Network)
}

func (pr *Prover) epochsPerSyncCommitteePeriod() uint64 {
	return lcrelay.EpochsPerSyncCommitteePeriod(pr.config.Network)
}

func (pr *Prover) getPeriodBoundarySlot(period uint64) uint64 {
	return lcrelay.GetPeriodBoundarySlot(pr.config.Network, period)
}

func (pr *Prover) computeSyncCommitteePeriod(epoch uint64) uint64 {
	return lcrelay.ComputeSyncCommitteePeriod(pr.config.Network, epoch)
}

func (pr *Prover) computeEpoch(slot uint64) uint64 {
	return lcrelay.ComputeEpoch(pr.config.Network, slot)
}

func (pr *Prover) getPeriodWithBlockNumber(ctx context.Context, blockNumber uint64) (uint64, error) {
	return lcrelay.GetPeriodWithBlockNumber(ctx, pr.beaconClient, pr.executionClient.Raw(), pr.config.Network, pr.config.getForkParameters(), blockNumber)
}

// getConsensusStateSlotWithBlockNumber returns the beacon slot a consensus state created for
// `blockNumber` records. Post-Gloas this is not the slot the execution block itself was
// produced at, see lcrelay.GetConsensusStateSlotWithBlockNumber.
func (pr *Prover) getConsensusStateSlotWithBlockNumber(ctx context.Context, blockNumber uint64) (uint64, error) {
	return lcrelay.GetConsensusStateSlotWithBlockNumber(ctx, pr.beaconClient, pr.executionClient.Raw(), pr.config.Network, pr.config.getForkParameters(), blockNumber)
}

func (pr *Prover) buildExecutionUpdateFromFinalizedHeader(ctx context.Context, finalizedHeader *beacon.LightClientHeader) (*lctypes.ExecutionUpdate, error) {
	return lcrelay.BuildExecutionUpdateFromFinalizedHeader(ctx, pr.executionClient.Raw(), finalizedHeader, false)
}

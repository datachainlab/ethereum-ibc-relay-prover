package relay

import (
	"fmt"

	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/datachainlab/ethereum-ibc-relay-prover/beacon"
	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
)

// general
const (
	EXECUTION_STATE_ROOT_INDEX   = 17
	EXECUTION_BLOCK_NUMBER_INDEX = 21
)

// minimal preset
const (
	MINIMAL_SECONDS_PER_SLOT                 uint64 = 6
	MINIMAL_SLOTS_PER_EPOCH                  uint64 = 8
	MINIMAL_EPOCHS_PER_SYNC_COMMITTEE_PERIOD uint64 = 8
)

// mainnet preset
const (
	MAINNET_SECONDS_PER_SLOT                 uint64 = 12
	MAINNET_SLOTS_PER_EPOCH                  uint64 = 32
	MAINNET_EPOCHS_PER_SYNC_COMMITTEE_PERIOD uint64 = 256
)

func (pr *Prover) computeSyncCommitteePeriod(epoch uint64) uint64 {
	if pr.config.IsMainnetPreset() {
		return epoch / MAINNET_EPOCHS_PER_SYNC_COMMITTEE_PERIOD
	} else {
		return epoch / MINIMAL_EPOCHS_PER_SYNC_COMMITTEE_PERIOD
	}
}

func (pr *Prover) computeEpoch(slot uint64) uint64 {
	if pr.config.IsMainnetPreset() {
		return slot / MAINNET_SLOTS_PER_EPOCH
	} else {
		return slot / MINIMAL_SLOTS_PER_EPOCH
	}
}

func (pr *Prover) getLightClientBootstrap() (*beacon.LightClientBootstrap, error) {
	cps, err := pr.beaconClient.GetFinalityCheckpoints()
	if err != nil {
		return nil, err
	}
	res, err := pr.beaconClient.GetBootstrap(cps.Finalized.Root[:])
	if err != nil {
		return nil, err
	}
	return &res.Data, nil
}

// find the period corresponding to the client state's latest height (execution block number).
func (pr *Prover) findLCPeriodByHeight(height ibcexported.Height, latestPeriod uint64) (uint64, error) {
	var getPeriodStartingSlot = func(period uint64) uint64 {
		var (
			slotsPerEpoch   uint64
			epochsPerPeriod uint64
		)
		if pr.config.IsMainnetPreset() {
			slotsPerEpoch = MAINNET_SLOTS_PER_EPOCH
			epochsPerPeriod = MAINNET_EPOCHS_PER_SYNC_COMMITTEE_PERIOD
		} else {
			slotsPerEpoch = MINIMAL_SLOTS_PER_EPOCH
			epochsPerPeriod = MINIMAL_EPOCHS_PER_SYNC_COMMITTEE_PERIOD
		}
		return period * epochsPerPeriod * slotsPerEpoch
	}

	bn := height.GetRevisionHeight()

	// NOTE: In most cases, it will match the `latestPeriod` or `latestPeriod-1`.
	// TODO should use binary-search in combination for long period case?
	for p := int64(latestPeriod); p >= 0; p = p - 1 {
		slot := getPeriodStartingSlot(uint64(p))
		root, err := pr.beaconClient.GetBlockRoot(slot)
		if err != nil {
			return 0, err
		}
		// NOTE: I assumed the cost of bootstrap API to get the block number was cheaper than getBlock API (at least, this is true for the relayer)
		res, err := pr.beaconClient.GetBootstrap(root.Data.Root)
		if err != nil {
			return 0, err
		}
		if bn >= res.Data.Header.Execution.BlockNumber {
			return uint64(p), nil
		}
	}
	return 0, fmt.Errorf("something wong...: target_block_number=%v latest_period=%v", bn, latestPeriod)
}

func (pr *Prover) buildExecutionUpdate(executionHeader *beacon.ExecutionPayloadHeader) (*lctypes.ExecutionUpdate, error) {
	stateRootBranch, err := generate_execution_payload_proof(executionHeader, EXECUTION_STATE_ROOT_INDEX)
	if err != nil {
		return nil, err
	}
	blockNumberBranch, err := generate_execution_payload_proof(executionHeader, EXECUTION_BLOCK_NUMBER_INDEX)
	if err != nil {
		return nil, err
	}
	return &lctypes.ExecutionUpdate{
		StateRoot:         executionHeader.StateRoot,
		StateRootBranch:   stateRootBranch,
		BlockNumber:       executionHeader.BlockNumber,
		BlockNumberBranch: blockNumberBranch,
	}, nil
}

package relay

import (
	"fmt"

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

func (pr *Prover) secondsPerSlot() uint64 {
	if pr.config.IsMainnetPreset() {
		return MAINNET_SECONDS_PER_SLOT
	} else {
		return MINIMAL_SECONDS_PER_SLOT
	}
}

func (pr *Prover) slotsPerEpoch() uint64 {
	if pr.config.IsMainnetPreset() {
		return MAINNET_SLOTS_PER_EPOCH
	} else {
		return MINIMAL_SLOTS_PER_EPOCH
	}
}

func (pr *Prover) epochsPerSyncCommitteePeriod() uint64 {
	if pr.config.IsMainnetPreset() {
		return MAINNET_EPOCHS_PER_SYNC_COMMITTEE_PERIOD
	} else {
		return MINIMAL_EPOCHS_PER_SYNC_COMMITTEE_PERIOD
	}
}

// returns the first slot of the period
func (pr *Prover) getPeriodBoundarySlot(period uint64) uint64 {
	return period * pr.epochsPerSyncCommitteePeriod() * pr.slotsPerEpoch()
}

func (pr *Prover) computeSyncCommitteePeriod(epoch uint64) uint64 {
	return epoch / pr.epochsPerSyncCommitteePeriod()
}

func (pr *Prover) computeEpoch(slot uint64) uint64 {
	return slot / pr.slotsPerEpoch()
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

// find a period corresponding to a given execution block number
// CONTRACT: a returned period must be smaller than or equal to the period of `highestSlot`
func (pr *Prover) findPeriodByBlockNumber(bn uint64, highestSlot uint64) (uint64, error) {
	// NOTE: In most cases, it will match the `highestPeriod` or `highestPeriod-1`.
	// TODO should use binary-search in combination for long periods case?
	highestPeriod := pr.computeSyncCommitteePeriod(pr.computeEpoch(highestSlot))
	for p := int64(highestPeriod); p >= 0; p = p - 1 {
		var (
			root []byte
			err  error
		)
		if p == int64(highestPeriod) {
			root, err = pr.getFirstBlockRootInPeriod(uint64(p), &highestSlot)
		} else {
			root, err = pr.getFirstBlockRootInPeriod(uint64(p), nil)
		}
		if err != nil {
			return 0, err
		}
		// NOTE: I assumed the cost of bootstrap API to get the block number was cheaper than getBlock API (at least, this is true for the relayer)
		res, err := pr.beaconClient.GetBootstrap(root)
		if err != nil {
			return 0, err
		}
		if bn >= res.Data.Header.Execution.BlockNumber {
			return uint64(p), nil
		}
	}
	return 0, fmt.Errorf("something wong...: target_block_number=%v highest_period=%v highest_slot=%v", bn, highestPeriod, highestSlot)
}

// get the root of first block the period
func (pr *Prover) getFirstBlockRootInPeriod(period uint64, highestSlot *uint64) ([]byte, error) {
	var endSlot uint64
	if highestSlot == nil {
		endSlot = pr.getPeriodBoundarySlot(period+1) - 1
	} else {
		endSlot = *highestSlot
	}
	for i := pr.getPeriodBoundarySlot(period); i <= endSlot; i++ {
		root, err := pr.beaconClient.GetBlockRoot(i)
		if err != nil {
			return nil, err
		}
		if root.Data.Root != nil {
			return root.Data.Root, nil
		}
	}
	return nil, fmt.Errorf("getNearestBlockRoot: not found: period=%v endSlot=%v", period, endSlot)
}

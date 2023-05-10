package relay

import (
	"fmt"
	"math"

	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/hyperledger-labs/yui-ibc-solidity/pkg/relay/ethereum"
	"github.com/hyperledger-labs/yui-relayer/core"
)

const (
	Mainnet = "mainnet"
	Minimal = "minimal"
	Goerli  = "goerli"
	Sepolia = "sepolia"
)

var _ core.ProverConfig = (*ProverConfig)(nil)

func (prc ProverConfig) Build(chain core.Chain) (core.Prover, error) {
	ec, ok := chain.(*ethereum.Chain)
	if !ok {
		return nil, fmt.Errorf("unexpected chain: %T", chain)
	}
	return NewProver(ec, prc), nil
}

// NOTE the prover supports only the mainnet and minimal preset for now
func (prc *ProverConfig) IsMainnetPreset() bool {
	switch prc.Network {
	case Mainnet, Goerli, Sepolia:
		return true
	case Minimal:
		return false
	default:
		panic(fmt.Sprintf("unknown network: %v", prc.Network))
	}
}

func (prc *ProverConfig) getSecondsPerSlot() uint64 {
	if prc.IsMainnetPreset() {
		return 12
	} else {
		return 6
	}
}

func (prc *ProverConfig) getSlotsPerEpoch() uint64 {
	if prc.IsMainnetPreset() {
		return 32
	} else {
		return 8
	}
}

func (prc *ProverConfig) getEpochsPerSyncCommitteePeriod() uint64 {
	if prc.IsMainnetPreset() {
		return 256
	} else {
		return 8
	}
}

func (prc *ProverConfig) getForkParameters() *lctypes.ForkParameters {
	switch prc.Network {
	case Mainnet:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 0, 0},
			GenesisSlot:        0,
			Altair: &lctypes.Fork{
				Version: []byte{1, 0, 0, 0},
				Epoch:   74240,
			},
			Bellatrix: &lctypes.Fork{
				Version: []byte{2, 0, 0, 0},
				Epoch:   144896,
			},
			Capella: &lctypes.Fork{
				Version: []byte{3, 0, 0, 0},
				Epoch:   194048,
			},
			Eip4844: &lctypes.Fork{
				Version: []byte{4, 0, 0, 0},
				Epoch:   math.MaxUint64,
			},
		}
	case Minimal:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 0, 1},
			GenesisSlot:        0,
			Altair: &lctypes.Fork{
				Version: []byte{1, 0, 0, 1},
				Epoch:   0,
			},
			Bellatrix: &lctypes.Fork{
				Version: []byte{2, 0, 0, 1},
				Epoch:   0,
			},
			Capella: &lctypes.Fork{
				Version: []byte{3, 0, 0, 1},
				Epoch:   0,
			},
			Eip4844: &lctypes.Fork{
				Version: []byte{4, 0, 0, 1},
				Epoch:   math.MaxUint64,
			},
		}
	case Goerli:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 16, 32},
			GenesisSlot:        0,
			Altair: &lctypes.Fork{
				Version: []byte{1, 0, 16, 32},
				Epoch:   36660,
			},
			Bellatrix: &lctypes.Fork{
				Version: []byte{2, 0, 16, 32},
				Epoch:   112260,
			},
			Capella: &lctypes.Fork{
				Version: []byte{3, 0, 16, 32},
				Epoch:   162304,
			},
			Eip4844: &lctypes.Fork{
				Version: []byte{4, 0, 0, 0},
				Epoch:   math.MaxUint64,
			},
		}
	case Sepolia:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{144, 0, 0, 105},
			GenesisSlot:        0,
			Altair: &lctypes.Fork{
				Version: []byte{144, 0, 0, 112},
				Epoch:   50,
			},
			Bellatrix: &lctypes.Fork{
				Version: []byte{144, 0, 0, 113},
				Epoch:   100,
			},
			Capella: &lctypes.Fork{
				Version: []byte{144, 0, 0, 114},
				Epoch:   56832,
			},
			Eip4844: &lctypes.Fork{
				Version: []byte{4, 0, 0, 0},
				Epoch:   math.MaxUint64,
			},
		}
	default:
		panic(fmt.Sprintf("unknown network: %v", prc.Network))
	}
}

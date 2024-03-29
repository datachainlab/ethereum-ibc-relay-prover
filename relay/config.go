package relay

import (
	"fmt"
	"time"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
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
		return nil, fmt.Errorf("expected chain type is %T, but got %T", &ethereum.Chain{}, chain)
	}
	if err := prc.Validate(); err != nil {
		return nil, err
	}
	return NewProver(ec, prc), nil
}

func (prc ProverConfig) Validate() error {
	if prc.Network == "" {
		return fmt.Errorf("network is required")
	}
	if prc.BeaconEndpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	_, err := time.ParseDuration(prc.TrustingPeriod)
	if err != nil {
		return err
	}
	_, err = time.ParseDuration(prc.MaxClockDrift)
	if err != nil {
		return err
	}
	if prc.RefreshThresholdRate == nil {
		return fmt.Errorf("config attribute \"refresh_threshold_rate\" is required")
	}
	if prc.RefreshThresholdRate.Denominator == 0 {
		return fmt.Errorf("config attribute \"refresh_threshold_rate.denominator\" must not be zero")
	}
	if prc.RefreshThresholdRate.Numerator == 0 {
		return fmt.Errorf("config attribute \"refresh_threshold_rate.numerator\" must not be zero")
	}
	if prc.RefreshThresholdRate.Numerator > prc.RefreshThresholdRate.Denominator {
		return fmt.Errorf("config attribute \"refresh_threshold_rate\" must be less than or equal to 1.0: actual=%v/%v", prc.RefreshThresholdRate.Numerator, prc.RefreshThresholdRate.Denominator)
	}
	return nil
}

func (prc *ProverConfig) GetTrustingPeriod() time.Duration {
	if d, err := time.ParseDuration(prc.TrustingPeriod); err != nil {
		panic(err)
	} else {
		return d
	}
}

func (prc *ProverConfig) GetMaxClockDrift() time.Duration {
	if d, err := time.ParseDuration(prc.MaxClockDrift); err != nil {
		panic(err)
	} else {
		return d
	}
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

func (prc *ProverConfig) getForkParameters() *lctypes.ForkParameters {
	switch prc.Network {
	case Mainnet:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 0, 0},
			Forks: []*lctypes.Fork{
				{
					Version: []byte{4, 0, 0, 0},
					Epoch:   269568,
				},
				{
					Version: []byte{3, 0, 0, 0},
					Epoch:   194048,
				},
				{
					Version: []byte{2, 0, 0, 0},
					Epoch:   144896,
				},
				{
					Version: []byte{1, 0, 0, 0},
					Epoch:   74240,
				},
			},
		}
	case Minimal:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 0, 1},
			Forks: []*lctypes.Fork{
				{
					Version: []byte{4, 0, 0, 1},
					Epoch:   0,
				},
				{
					Version: []byte{3, 0, 0, 1},
					Epoch:   0,
				},
				{
					Version: []byte{2, 0, 0, 1},
					Epoch:   0,
				},
				{
					Version: []byte{1, 0, 0, 1},
					Epoch:   0,
				},
			},
		}
	case Goerli:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 16, 32},
			Forks: []*lctypes.Fork{
				{
					Version: []byte{4, 0, 16, 32},
					Epoch:   231680,
				},
				{
					Version: []byte{3, 0, 16, 32},
					Epoch:   162304,
				},
				{
					Version: []byte{2, 0, 16, 32},
					Epoch:   112260,
				},
				{
					Version: []byte{1, 0, 16, 32},
					Epoch:   36660,
				},
			},
		}
	case Sepolia:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{144, 0, 0, 105},
			Forks: []*lctypes.Fork{
				{
					Version: []byte{144, 0, 0, 115},
					Epoch:   132608,
				},
				{
					Version: []byte{144, 0, 0, 114},
					Epoch:   56832,
				},
				{
					Version: []byte{144, 0, 0, 113},
					Epoch:   100,
				},
				{
					Version: []byte{144, 0, 0, 112},
					Epoch:   50,
				},
			},
		}
	default:
		panic(fmt.Sprintf("unknown network: %v", prc.Network))
	}
}

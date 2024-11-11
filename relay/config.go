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

const (
	MAINNET_PRESET_SYNC_COMMITTEE_SIZE = 512
	MINIMAL_PRESET_SYNC_COMMITTEE_SIZE = 32
)

var (
	AltairSpec = lctypes.ForkSpec{
		FinalizedRootGindex:        105,
		CurrentSyncCommitteeGindex: 54,
		NextSyncCommitteeGindex:    55,
	}
	BellatrixSpec = lctypes.ForkSpec{
		FinalizedRootGindex:               AltairSpec.FinalizedRootGindex,
		CurrentSyncCommitteeGindex:        AltairSpec.CurrentSyncCommitteeGindex,
		NextSyncCommitteeGindex:           AltairSpec.NextSyncCommitteeGindex,
		ExecutionPayloadGindex:            25,
		ExecutionPayloadStateRootGindex:   18,
		ExecutionPayloadBlockNumberGindex: 22,
	}
	CapellaSpec = BellatrixSpec
	DenebSpec   = lctypes.ForkSpec{
		FinalizedRootGindex:               CapellaSpec.FinalizedRootGindex,
		CurrentSyncCommitteeGindex:        CapellaSpec.CurrentSyncCommitteeGindex,
		NextSyncCommitteeGindex:           CapellaSpec.NextSyncCommitteeGindex,
		ExecutionPayloadGindex:            CapellaSpec.ExecutionPayloadGindex,
		ExecutionPayloadStateRootGindex:   34,
		ExecutionPayloadBlockNumberGindex: 38,
	}
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
					Version: []byte{1, 0, 0, 0},
					Epoch:   74240,
					Spec:    &AltairSpec,
				},
				{
					Version: []byte{2, 0, 0, 0},
					Epoch:   144896,
					Spec:    &BellatrixSpec,
				},
				{
					Version: []byte{3, 0, 0, 0},
					Epoch:   194048,
					Spec:    &CapellaSpec,
				},
				{
					Version: []byte{4, 0, 0, 0},
					Epoch:   269568,
					Spec:    &DenebSpec,
				},
			},
		}
	case Minimal:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 0, 1},
			Forks: []*lctypes.Fork{
				{
					Version: []byte{1, 0, 0, 1},
					Epoch:   0,
					Spec:    &AltairSpec,
				},
				{
					Version: []byte{2, 0, 0, 1},
					Epoch:   0,
					Spec:    &BellatrixSpec,
				},
				{
					Version: []byte{3, 0, 0, 1},
					Epoch:   0,
					Spec:    &CapellaSpec,
				},
				{
					Version: []byte{4, 0, 0, 1},
					Epoch:   0,
					Spec:    &DenebSpec,
				},
			},
		}
	case Goerli:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{0, 0, 16, 32},
			Forks: []*lctypes.Fork{
				{
					Version: []byte{1, 0, 16, 32},
					Epoch:   36660,
					Spec:    &AltairSpec,
				},
				{
					Version: []byte{2, 0, 16, 32},
					Epoch:   112260,
					Spec:    &BellatrixSpec,
				},
				{
					Version: []byte{3, 0, 16, 32},
					Epoch:   162304,
					Spec:    &CapellaSpec,
				},
				{
					Version: []byte{4, 0, 16, 32},
					Epoch:   231680,
					Spec:    &DenebSpec,
				},
			},
		}
	case Sepolia:
		return &lctypes.ForkParameters{
			GenesisForkVersion: []byte{144, 0, 0, 105},
			Forks: []*lctypes.Fork{
				{
					Version: []byte{144, 0, 0, 112},
					Epoch:   50,
					Spec:    &AltairSpec,
				},
				{
					Version: []byte{144, 0, 0, 113},
					Epoch:   100,
					Spec:    &BellatrixSpec,
				},
				{
					Version: []byte{144, 0, 0, 114},
					Epoch:   56832,
					Spec:    &CapellaSpec,
				},
				{
					Version: []byte{144, 0, 0, 115},
					Epoch:   132608,
					Spec:    &DenebSpec,
				},
			},
		}
	default:
		panic(fmt.Sprintf("unknown network: %v", prc.Network))
	}
}

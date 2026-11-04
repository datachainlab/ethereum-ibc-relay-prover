package relay

import (
	"fmt"
	"time"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	lcrelay "github.com/datachainlab/ethereum-light-client-types/prover/relay"
	lctypes "github.com/datachainlab/ethereum-light-client-types/prover/types"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/coreutil"
	"github.com/hyperledger-labs/yui-relayer/otelcore"
)

var _ core.ProverConfig = (*ProverConfig)(nil)

func (prc ProverConfig) Build(chain core.Chain) (core.Prover, error) {
	ec, err := coreutil.UnwrapChain[*ethereum.Chain](chain)
	if err != nil {
		return nil, err
	}
	if err := prc.Validate(); err != nil {
		return nil, err
	}
	// Use chain, not ec, for the case where the chain is wrapped by another struct that implements core.Chain (e.g. tracing bridge)
	return otelcore.NewProver(NewProver(chain, prc, ec.Config().IBCAddress(), ec.Client()), chain.ChainID(), tracer), nil
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
	for hf := range prc.MinimalForkSched {
		switch hf {
		case lcrelay.Altair, lcrelay.Bellatrix, lcrelay.Capella, lcrelay.Deneb, lcrelay.Electra, lcrelay.Fulu, lcrelay.Gloas:
			// OK
		default:
			return fmt.Errorf("config attribute \"minimal_fork_sched\" contains an unknown key: %s", hf)
		}
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
	return lcrelay.IsMainnetPreset(prc.Network)
}

func (prc *ProverConfig) getForkParameters() *lctypes.ForkParameters {
	return lcrelay.GetForkParameters(prc.Network, prc.MinimalForkSched)
}

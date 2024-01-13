package relay

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/hyperledger-labs/yui-relayer/core"
)

const (
	MainnetPreset = "mainnet"
	MinimalPreset = "minimal"
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
	if prc.Preset != MainnetPreset && prc.Preset != MinimalPreset {
		return fmt.Errorf("unknown preset: %v", prc.Preset)
	}
	if prc.ForkParameters == nil || len(prc.ForkParameters) == 0 {
		return fmt.Errorf("config attribute \"fork_parameters\" must not be empty")
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
	return prc.Preset == MainnetPreset
}

func (prc ProverConfig) validateForkParamerters() error {
	if len(prc.ForkParameters) == 0 {
		return fmt.Errorf("config attribute \"fork_parameters\" must not be empty")
	}
	// check if `ForkParameters`` are sorted by epoch in descending order
	for i := 0; i < len(prc.ForkParameters)-1; i++ {
		if prc.ForkParameters[i].Epoch > prc.ForkParameters[i+1].Epoch {
			return fmt.Errorf("config attribute \"fork_parameters\" must be sorted by epoch in descending order: actual=%v", prc.ForkParameters)
		}
		// check if `Version` must be 4 bytes as a hex string
		if bz, err := decodeHex(prc.ForkParameters[i].Version); err != nil {
			return fmt.Errorf("config attribute \"fork_parameters[%v].version\" must be a hex string: actual=%v", i, prc.ForkParameters[i].Version)
		} else if len(bz) != 4 {
			return fmt.Errorf("config attribute \"fork_parameters[%v].version\" must be 4 bytes: actual=%v", i, prc.ForkParameters[i].Version)
		}
	}
	// check if last `ForkParameters` is the genesis fork
	if prc.ForkParameters[len(prc.ForkParameters)-1].Epoch != 0 {
		return fmt.Errorf("genesis fork epoch must be 0: actual=%v", prc.ForkParameters[len(prc.ForkParameters)-1].Epoch)
	}
	return nil
}

func (prc *ProverConfig) getForkParameters() *lctypes.ForkParameters {
	// we assume the followings:
	// 1. `ForkParameters` length is greater than 0
	// 2. `ForkParameters` are sorted by epoch in descending order
	// 3. `Version` must be 4 bytes as a hex string
	// 4. last `ForkParameters` is the genesis fork

	// last fork must be the genesis fork
	genesisFork := prc.ForkParameters[len(prc.ForkParameters)-1]
	if genesisFork.Epoch != 0 {
		panic(fmt.Sprintf("genesis fork epoch must be 0: actual=%v", genesisFork.Epoch))
	}
	var forkParameters lctypes.ForkParameters
	forkParameters.GenesisForkVersion = mustDecodeHex(genesisFork.Version)
	for _, fp := range prc.ForkParameters[:len(prc.ForkParameters)-1] {
		forkParameters.Forks = append(forkParameters.Forks, &lctypes.Fork{
			Version: mustDecodeHex(fp.Version),
			Epoch:   fp.Epoch,
		})
	}
	return &forkParameters
}

func decodeHex(s string) ([]byte, error) {
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	return hex.DecodeString(s)
}

func mustDecodeHex(s string) []byte {
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

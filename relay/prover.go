package relay

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-prover/beacon"
	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"
)

var IBCCommitmentsSlot = common.HexToHash("1ee222554989dda120e26ecacf756fe1235cd8d726706b57517715dde4f0c900")

type Prover struct {
	chain           core.Chain
	config          ProverConfig
	ibcAddress      common.Address
	executionClient *client.ETHClient
	beaconClient    beacon.Client
	codec           codec.ProtoCodecMarshaler
}

func NewProver(chain core.Chain, config ProverConfig, ibcAddress common.Address, executionClient *client.ETHClient) *Prover {
	beaconClient := beacon.NewClient(config.BeaconEndpoint)
	return &Prover{
		chain:           chain,
		config:          config,
		ibcAddress:      ibcAddress,
		executionClient: executionClient,
		beaconClient:    beaconClient,
	}
}

func (pr *Prover) GetLogger() *log.RelayLogger {
	return log.GetLogger().WithChain(pr.chain.ChainID()).WithModule(ModuleName)
}

//--------- Prover implementation ---------//

var _ core.Prover = (*Prover)(nil)

// Init initializes the chain
func (pr *Prover) Init(homePath string, timeout time.Duration, codec codec.ProtoCodecMarshaler, debug bool) error {
	pr.codec = codec
	return nil
}

// SetRelayInfo sets source's path and counterparty's info to the chain
func (pr *Prover) SetRelayInfo(path *core.PathEnd, counterparty *core.ProvableChain, counterpartyPath *core.PathEnd) error {
	return nil
}

// SetupForRelay performs chain-specific setup before starting the relay
func (pr *Prover) SetupForRelay(ctx context.Context) error {
	return nil
}

//--------- LightClient implementation ---------//

type InitialState struct {
	Genesis              beacon.Genesis
	Slot                 uint64
	BlockNumber          uint64
	AccountStorageRoot   [32]byte
	Timestamp            time.Time
	CurrentSyncCommittee lctypes.SyncCommittee
	NextSyncCommittee    lctypes.SyncCommittee
}

// CreateInitialLightClientState returns a pair of ClientState and ConsensusState based on the state of the self chain at `height`.
// These states will be submitted to the counterparty chain as MsgCreateClient.
// If `height` is nil, the latest finalized height is selected automatically.
func (pr *Prover) CreateInitialLightClientState(ctx context.Context, height ibcexported.Height) (ibcexported.ClientState, ibcexported.ConsensusState, error) {
	if height == nil {
		height = pr.newHeight(0)
	}
	initialState, err := pr.buildInitialState(ctx, height.GetRevisionHeight())
	if err != nil {
		return nil, nil, err
	}
	pr.GetLogger().DebugContext(ctx, "InitialState", "initial_state", initialState)
	committeeSize := len(initialState.CurrentSyncCommittee.Pubkeys)
	if pr.config.IsMainnetPreset() {
		if committeeSize != MAINNET_PRESET_SYNC_COMMITTEE_SIZE {
			return nil, nil, fmt.Errorf("the size of current sync committee is not %v: actual=%v", MAINNET_PRESET_SYNC_COMMITTEE_SIZE, committeeSize)
		}
	} else {
		if committeeSize != MINIMAL_PRESET_SYNC_COMMITTEE_SIZE {
			return nil, nil, fmt.Errorf("the size of current sync committee is not %v: actual=%v", MINIMAL_PRESET_SYNC_COMMITTEE_SIZE, committeeSize)
		}
	}
	clientState := pr.buildClientState(
		initialState.Genesis.GenesisValidatorsRoot[:],
		initialState.Genesis.GenesisTimeSeconds,
		initialState.BlockNumber,
	)
	consensusState := &lctypes.ConsensusState{
		Slot:                 initialState.Slot,
		StorageRoot:          initialState.AccountStorageRoot[:],
		Timestamp:            initialState.Timestamp,
		CurrentSyncCommittee: initialState.CurrentSyncCommittee.AggregatePubkey,
		NextSyncCommittee:    initialState.NextSyncCommittee.AggregatePubkey,
	}
	return clientState, consensusState, nil
}

// SetupHeadersForUpdate returns the finalized header and any intermediate headers needed to apply it to the client on the counterparty chain
// The order of the returned header slice should be as: [<intermediate headers>..., <update header>]
// if the header slice's length == 0 and err == nil, the relayer should skip the update-client
func (pr *Prover) SetupHeadersForUpdate(ctx context.Context, counterparty core.FinalityAwareChain, latestFinalizedHeader core.Header) ([]core.Header, error) {
	lfh, ok := latestFinalizedHeader.(*lctypes.Header)
	if !ok {
		return nil, fmt.Errorf("unexpected header type: %T", latestFinalizedHeader)
	}
	if err := lfh.ValidateBasic(); err != nil {
		return nil, err
	}

	latestHeight, err := counterparty.LatestHeight(ctx)
	if err != nil {
		return nil, err
	}

	pr.GetLogger().DebugContext(ctx, "query the latest height of the counterparty chain", "latest_height", latestHeight)

	// retrieve the client state from the counterparty chain
	counterpartyClientRes, err := counterparty.QueryClientState(core.NewQueryContext(ctx, latestHeight))
	if err != nil {
		return nil, err
	}
	var cs ibcexported.ClientState
	if err := pr.codec.UnpackAny(counterpartyClientRes.ClientState, &cs); err != nil {
		return nil, fmt.Errorf("failed to unpack Any into client state: %v", err)
	}

	if cs.GetLatestHeight().GetRevisionHeight() == lfh.ExecutionUpdate.BlockNumber {
		return nil, nil
	} else if cs.GetLatestHeight().GetRevisionHeight() > lfh.ExecutionUpdate.BlockNumber {
		return nil, fmt.Errorf("the latest finalized header is older than the latest height of client state: finalized_block_number=%v client_latest_height=%v", lfh.ExecutionUpdate.BlockNumber, cs.GetLatestHeight().GetRevisionHeight())
	}

	statePeriod, err := pr.getPeriodWithBlockNumber(ctx, cs.GetLatestHeight().GetRevisionHeight())
	if err != nil {
		return nil, fmt.Errorf("failed to get period with block number: block_number=%v %v", cs.GetLatestHeight().GetRevisionHeight(), err)
	}
	latestPeriod := pr.computeSyncCommitteePeriod(pr.computeEpoch(lfh.ConsensusUpdate.SignatureSlot))

	pr.GetLogger().DebugContext(ctx, "try to setup headers for updating the light-client", "lc_latest_height", cs.GetLatestHeight(), "lc_latest_height_period", statePeriod, "latest_period", latestPeriod)

	if statePeriod == latestPeriod {
		latestHeight := cs.GetLatestHeight().(clienttypes.Height)
		res, err := pr.beaconClient.GetLightClientUpdate(ctx, statePeriod)
		if err != nil {
			return nil, fmt.Errorf("failed to get LightClientUpdate: state_period=%v %v", statePeriod, err)
		}
		root, err := res.Data.FinalizedHeader.Beacon.HashTreeRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash tree root: %v", err)
		}
		bootstrapRes, err := pr.beaconClient.GetBootstrap(ctx, root[:])
		if err != nil {
			return nil, fmt.Errorf("failed to get bootstrap: root=%x %v", root, err)
		}
		lfh.TrustedSyncCommittee = &lctypes.TrustedSyncCommittee{
			TrustedHeight: &latestHeight,
			SyncCommittee: bootstrapRes.Data.CurrentSyncCommittee.ToProto(),
			IsNext:        false,
		}
		return []core.Header{lfh}, nil
	} else if statePeriod > latestPeriod {
		return nil, fmt.Errorf("the light-client server's response is old: client_state_period=%v latest_finalized_period=%v", statePeriod, latestPeriod)
	}

	//--------- In case statePeriod < latestPeriod ---------//

	var (
		headers                     []core.Header
		trustedNextSyncCommittee    *lctypes.SyncCommittee
		trustedCurrentSyncCommittee *lctypes.SyncCommittee
		trustedHeight               = cs.GetLatestHeight().(clienttypes.Height)
	)
	pr.GetLogger().DebugContext(ctx, "setup headers for updating the light-client", "state_period", statePeriod, "latest_period", latestPeriod, "client_state_latest_height", cs.GetLatestHeight().GetRevisionHeight())
	res, err := pr.beaconClient.GetLightClientUpdate(ctx, statePeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to get LightClientUpdate: state_period=%v %v", statePeriod, err)
	}
	trustedNextSyncCommittee = res.Data.ToProto().NextSyncCommittee
	for p := statePeriod + 1; p <= latestPeriod; p++ {
		header, err := pr.buildNextSyncCommitteeUpdate(ctx, p, trustedHeight, trustedNextSyncCommittee)
		if err != nil {
			return nil, fmt.Errorf("failed to build next sync committee update for next: period=%v trusted_height=%v %v", p, trustedHeight, err)
		}
		pr.GetLogger().DebugContext(ctx, "setup intermediate header for updating the light-client", "period", p, "trusted_height", header.TrustedSyncCommittee.TrustedHeight, "trusted_sync_committee", fmt.Sprintf("0x%x", header.TrustedSyncCommittee.SyncCommittee.AggregatePubkey), "is_next", header.TrustedSyncCommittee.IsNext, "untrusted_execution_block_number", header.ExecutionUpdate.BlockNumber, "next_sync_committee", fmt.Sprintf("0x%x", header.ConsensusUpdate.NextSyncCommittee.AggregatePubkey))
		trustedHeight = clienttypes.NewHeight(0, header.ExecutionUpdate.BlockNumber)
		trustedCurrentSyncCommittee = trustedNextSyncCommittee
		trustedNextSyncCommittee = header.ConsensusUpdate.NextSyncCommittee
		headers = append(headers, header)
	}
	if trustedCurrentSyncCommittee == nil { // never happen
		panic(fmt.Errorf("trusted current sync committee must not be nil: period=%v", statePeriod))
	}
	if trustedHeight.GT(lfh.GetHeight()) {
		return nil, fmt.Errorf("the latest finalized header is older than the trusted height: finalized_block_number=%v trusted_block_number=%v", lfh.GetHeight().GetRevisionHeight(), trustedHeight.GetRevisionHeight())
	} else if trustedHeight.EQ(lfh.GetHeight()) {
		pr.GetLogger().DebugContext(ctx, "the latest finalized header is the same as the trusted height", "finalized_block_number", lfh.GetHeight().GetRevisionHeight(), "trusted_block_number", trustedHeight.GetRevisionHeight())
		return headers, nil
	}
	lfh.TrustedSyncCommittee = &lctypes.TrustedSyncCommittee{
		TrustedHeight: &trustedHeight,
		SyncCommittee: trustedCurrentSyncCommittee,
		IsNext:        false,
	}
	headers = append(headers, lfh)
	return headers, nil
}

// if `blockNumber` is 0, the latest block number is used
func (pr *Prover) buildInitialState(ctx context.Context, blockNumber uint64) (*InitialState, error) {
	res, err := pr.beaconClient.GetLightClientFinalityUpdate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get light-client finality update: %v", err)
	}
	if eh := &res.Data.FinalizedHeader.Execution; blockNumber == 0 {
		blockNumber = eh.BlockNumber
	} else if eh.BlockNumber < blockNumber {
		return nil, fmt.Errorf("the height is not finalized yet: blockNumber=%v finalized_block_number=%v", blockNumber, eh.BlockNumber)
	}

	timestamp, err := pr.chain.Timestamp(ctx, pr.newHeight(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get timestamp: %v", err)
	}
	if truncatedTm := timestamp.Truncate(time.Second); truncatedTm != timestamp {
		return nil, fmt.Errorf("ethereum timestamp must be truncated to seconds: timestamp=%v truncated_timestamp=%v", timestamp, truncatedTm)
	}

	slot, err := pr.getSlotAtTimestamp(ctx, uint64(timestamp.Unix()))
	if err != nil {
		return nil, fmt.Errorf("failed to compute slot at timestamp: %v", err)
	}

	period := pr.computeSyncCommitteePeriod(pr.computeEpoch(slot))

	pr.GetLogger().InfoContext(ctx, "build initial state", "slot", slot, "block_number", blockNumber, "period", period)
	currentSyncCommittee, err := pr.getBootstrapInPeriod(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get bootstrap in period %v: %v", period, err)
	}
	accountUpdate, err := pr.buildAccountUpdate(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to build account update: %v", err)
	}
	var accountStorageRoot [32]byte
	copy(accountStorageRoot[:], accountUpdate.AccountStorageRoot)
	genesis, err := pr.beaconClient.GetGenesis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get genesis: %v", err)
	}
	res2, err := pr.beaconClient.GetLightClientUpdate(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get LightClientUpdate: period=%v %v", period, err)
	}
	nextSyncCommittee := res2.Data.ToProto().NextSyncCommittee
	return &InitialState{
		Genesis:              *genesis,
		Slot:                 slot,
		BlockNumber:          blockNumber,
		AccountStorageRoot:   accountStorageRoot,
		Timestamp:            timestamp,
		CurrentSyncCommittee: *currentSyncCommittee,
		NextSyncCommittee:    *nextSyncCommittee,
	}, nil
}

// GetLatestFinalizedHeader returns the latest finalized header on this chain
// The returned header is expected to be the latest one of headers that can be verified by the light client
func (pr *Prover) GetLatestFinalizedHeader(ctx context.Context) (headers core.Header, err error) {
	res, err := pr.beaconClient.GetLightClientFinalityUpdate(ctx)
	if err != nil {
		return nil, err
	}
	lcUpdate := res.Data.ToProto()
	executionHeader := &res.Data.FinalizedHeader.Execution
	executionUpdate, err := pr.buildExecutionUpdate(executionHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to build execution update: %v", err)
	}
	executionRoot, err := executionHeader.HashTreeRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate execution root: %v", err)
	}
	if !bytes.Equal(executionRoot[:], lcUpdate.FinalizedExecutionRoot) {
		return nil, fmt.Errorf("execution root mismatch: %X != %X", executionRoot, lcUpdate.FinalizedExecutionRoot)
	}

	accountUpdate, err := pr.buildAccountUpdate(ctx, executionHeader.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to build account update: %v", err)
	}
	pr.GetLogger().InfoContext(ctx, "build latest finalized header", "block_number", executionHeader.BlockNumber, "timestamp", executionHeader.Timestamp, "state_root", hex.EncodeToString(executionHeader.StateRoot))
	return &lctypes.Header{
		ConsensusUpdate: lcUpdate,
		ExecutionUpdate: executionUpdate,
		AccountUpdate:   accountUpdate,
		Timestamp:       executionHeader.Timestamp,
	}, nil
}

func (pr *Prover) CheckRefreshRequired(ctx context.Context, counterparty core.ChainInfoICS02Querier) (bool, error) {
	cpQueryHeight, err := counterparty.LatestHeight(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the counterparty chain: %v", err)
	}
	cpQueryCtx := core.NewQueryContext(ctx, cpQueryHeight)

	resCs, err := counterparty.QueryClientState(cpQueryCtx)
	if err != nil {
		return false, fmt.Errorf("failed to query the client state on the counterparty chain: %v", err)
	}

	var cs ibcexported.ClientState
	if err := pr.codec.UnpackAny(resCs.ClientState, &cs); err != nil {
		return false, fmt.Errorf("failed to unpack Any into tendermint client state: %v", err)
	}

	resCons, err := counterparty.QueryClientConsensusState(cpQueryCtx, cs.GetLatestHeight())
	if err != nil {
		return false, fmt.Errorf("failed to query the consensus state on the counterparty chain: %v", err)
	}

	var cons ibcexported.ConsensusState
	if err := pr.codec.UnpackAny(resCons.ConsensusState, &cons); err != nil {
		return false, fmt.Errorf("failed to unpack Any into tendermint consensus state: %v", err)
	}
	lcLastTimestamp := time.Unix(0, int64(cons.GetTimestamp()))

	selfQueryHeight, err := pr.chain.LatestHeight(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the self chain: %v", err)
	}

	selfTimestamp, err := pr.chain.Timestamp(ctx, selfQueryHeight)
	if err != nil {
		return false, fmt.Errorf("failed to get timestamp of the self chain: %v", err)
	}

	elapsedTime := selfTimestamp.Sub(lcLastTimestamp)

	durationMulByFraction := func(d time.Duration, f *Fraction) time.Duration {
		nsec := d.Nanoseconds() * int64(f.Numerator) / int64(f.Denominator)
		return time.Duration(nsec) * time.Nanosecond
	}
	needsRefresh := elapsedTime > durationMulByFraction(pr.config.GetTrustingPeriod(), pr.config.RefreshThresholdRate)

	return needsRefresh, nil
}

func (pr *Prover) buildClientState(
	genesisValidatorsRoot []byte,
	genesisTime uint64,
	latestExecutionBlockNumber uint64,
) *lctypes.ClientState {
	return &lctypes.ClientState{
		GenesisValidatorsRoot:        genesisValidatorsRoot,
		MinSyncCommitteeParticipants: 1,
		GenesisTime:                  genesisTime,

		ForkParameters:               pr.config.getForkParameters(),
		SecondsPerSlot:               pr.secondsPerSlot(),
		SlotsPerEpoch:                pr.slotsPerEpoch(),
		EpochsPerSyncCommitteePeriod: pr.epochsPerSyncCommitteePeriod(),

		IbcAddress:         pr.ibcAddress.Bytes(),
		IbcCommitmentsSlot: IBCCommitmentsSlot[:],
		TrustLevel: &lctypes.Fraction{
			Numerator:   2,
			Denominator: 3,
		},
		TrustingPeriod: pr.config.GetTrustingPeriod(),
		MaxClockDrift:  pr.config.GetMaxClockDrift(),

		LatestExecutionBlockNumber: latestExecutionBlockNumber,

		FrozenHeight: nil,
	}
}

func (pr *Prover) getBootstrapInPeriod(ctx context.Context, period uint64) (*lctypes.SyncCommittee, error) {
	slotsPerEpoch := pr.slotsPerEpoch()
	startSlot := pr.getPeriodBoundarySlot(period)
	lastSlotInPeriod := pr.getPeriodBoundarySlot(period+1) - 1
	pr.GetLogger().DebugContext(ctx, "get bootstrap in period", "period", period, "start_slot", startSlot, "last_slot_in_period", lastSlotInPeriod, "slots_per_epoch", slotsPerEpoch)
	var errs []error
	for i := startSlot + slotsPerEpoch; i <= lastSlotInPeriod; i += slotsPerEpoch {
		res, err := pr.beaconClient.GetBlockRoot(ctx, i, false)
		if err != nil {
			pr.GetLogger().WarnContext(ctx, "failed to get block root", "slot", i, "err", err)
			errs = append(errs, err)
			return nil, fmt.Errorf("there is no available bootstrap in period: period=%v err=%v", period, errors.Join(errs...))
		}
		bootstrap, err := pr.beaconClient.GetBootstrap(ctx, res.Data.Root[:])
		if err != nil {
			pr.GetLogger().WarnContext(ctx, "failed to get bootstrap", "root", res.Data.Root[:], "err", err)
			errs = append(errs, err)
			continue
		} else {
			return bootstrap.Data.CurrentSyncCommittee.ToProto(), nil
		}
	}
	return nil, fmt.Errorf("failed to get bootstrap in period: period=%v err=%v", period, errors.Join(errs...))
}

func (pr *Prover) buildNextSyncCommitteeUpdate(ctx context.Context, period uint64, trustedHeight clienttypes.Height, trustedNextSyncCommittee *lctypes.SyncCommittee) (*lctypes.Header, error) {
	res, err := pr.beaconClient.GetLightClientUpdate(ctx, period)
	if err != nil {
		return nil, err
	}
	lcUpdate := res.Data.ToProto()
	executionHeader := &res.Data.FinalizedHeader.Execution
	executionUpdate, err := pr.buildExecutionUpdate(executionHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to build execution update: %v", err)
	}
	executionRoot, err := executionHeader.HashTreeRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate execution root: %v", err)
	}
	if !bytes.Equal(executionRoot[:], lcUpdate.FinalizedExecutionRoot) {
		return nil, fmt.Errorf("execution root mismatch: %X != %X", executionRoot, lcUpdate.FinalizedExecutionRoot)
	}

	accountUpdate, err := pr.buildAccountUpdate(ctx, executionHeader.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to build account update: %v", err)
	}
	return &lctypes.Header{
		TrustedSyncCommittee: &lctypes.TrustedSyncCommittee{
			TrustedHeight: &trustedHeight,
			SyncCommittee: trustedNextSyncCommittee,
			IsNext:        true,
		},
		ConsensusUpdate: lcUpdate,
		ExecutionUpdate: executionUpdate,
		AccountUpdate:   accountUpdate,
		Timestamp:       executionHeader.Timestamp,
	}, nil
}

//--------- StateProver implementation ---------//

var _ core.StateProver = (*Prover)(nil)

// ProveState returns the proof of an IBC state specified by `path` and `value`
func (pr *Prover) ProveState(ctx core.QueryContext, path string, value []byte) ([]byte, clienttypes.Height, error) {
	proofHeight := int64(ctx.Height().GetRevisionHeight())
	height := pr.newHeight(proofHeight)
	proof, err := pr.buildStateProof(ctx.Context(), []byte(path), proofHeight)
	return proof, height, err
}

// ProveHostConsensusState returns an existence proof of the consensus state at `height`
// This proof would be ignored in ibc-go, but it is required to `getSelfConsensusState` of ibc-solidity.
func (pr *Prover) ProveHostConsensusState(ctx core.QueryContext, height ibcexported.Height, consensusState ibcexported.ConsensusState) (proof []byte, err error) {
	return clienttypes.MarshalConsensusState(pr.codec, consensusState)
}

func (pr *Prover) newHeight(blockNumber int64) clienttypes.Height {
	return clienttypes.NewHeight(0, uint64(blockNumber))
}

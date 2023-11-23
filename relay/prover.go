package relay

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ethereum-ibc-relay-prover/beacon"
	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/hyperledger-labs/yui-relayer/core"
)

type Prover struct {
	chain           *ethereum.Chain
	config          ProverConfig
	executionClient *client.ETHClient
	beaconClient    beacon.Client
	codec           codec.ProtoCodecMarshaler
}

func NewProver(chain *ethereum.Chain, config ProverConfig) *Prover {
	beaconClient := beacon.NewClient(config.BeaconEndpoint)
	return &Prover{chain: chain, config: config, executionClient: chain.Client(), beaconClient: beaconClient}
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
}

// CreateInitialLightClientState returns a pair of ClientState and ConsensusState based on the state of the self chain at `height`.
// These states will be submitted to the counterparty chain as MsgCreateClient.
// If `height` is nil, the latest finalized height is selected automatically.
func (pr *Prover) CreateInitialLightClientState(height ibcexported.Height) (ibcexported.ClientState, ibcexported.ConsensusState, error) {
	if height == nil {
		height = pr.newHeight(0)
	}
	initialState, err := pr.buildInitialState(height.GetRevisionHeight())
	if err != nil {
		return nil, nil, err
	}
	log.Printf("initial state is %#v", initialState)

	clientState := pr.newClientState()
	clientState.GenesisValidatorsRoot = initialState.Genesis.GenesisValidatorsRoot[:]
	clientState.GenesisTime = initialState.Genesis.GenesisTimeSeconds
	clientState.LatestSlot = uint64(initialState.Slot)
	clientState.LatestExecutionBlockNumber = initialState.BlockNumber

	consensusState := &lctypes.ConsensusState{
		Slot:                 initialState.Slot,
		StorageRoot:          initialState.AccountStorageRoot[:],
		Timestamp:            initialState.Timestamp,
		CurrentSyncCommittee: initialState.CurrentSyncCommittee.AggregatePubkey,
	}

	return clientState, consensusState, nil
}

// SetupHeadersForUpdate returns the finalized header and any intermediate headers needed to apply it to the client on the counterpaty chain
// The order of the returned header slice should be as: [<intermediate headers>..., <update header>]
// if the header slice's length == 0 and err == nil, the relayer should skips the update-client
func (pr *Prover) SetupHeadersForUpdate(counterparty core.FinalityAwareChain, latestFinalizedHeader core.Header) ([]core.Header, error) {
	lfh, ok := latestFinalizedHeader.(*lctypes.Header)
	if !ok {
		return nil, fmt.Errorf("unexpected header type: %T", latestFinalizedHeader)
	}
	if err := lfh.ValidateBasic(); err != nil {
		return nil, err
	}

	latestHeight, err := counterparty.LatestHeight()
	if err != nil {
		return nil, err
	}

	// retrieve the client state from the counterparty chain
	counterpartyClientRes, err := counterparty.QueryClientState(core.NewQueryContext(context.TODO(), latestHeight))
	if err != nil {
		return nil, err
	}
	var cs ibcexported.ClientState
	if err := pr.codec.UnpackAny(counterpartyClientRes.ClientState, &cs); err != nil {
		return nil, err
	}

	if cs.GetLatestHeight().GetRevisionHeight() == lfh.ExecutionUpdate.BlockNumber {
		return nil, nil
	} else if cs.GetLatestHeight().GetRevisionHeight() > lfh.ExecutionUpdate.BlockNumber {
		return nil, fmt.Errorf("the latest finalized header is older than the latest height of client state: finalized_block_number=%v client_latest_height=%v", lfh.ExecutionUpdate.BlockNumber, cs.GetLatestHeight().GetRevisionHeight())
	}

	statePeriod, err := pr.getPeriodWithBlockNumber(cs.GetLatestHeight().GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	latestPeriod := pr.computeSyncCommitteePeriod(pr.computeEpoch(lfh.ConsensusUpdate.FinalizedHeader.Slot))

	log.Printf("try to setup headers for updating the light-client: lc_latest_height=%v lc_latest_height_period=%v latest_period=%v", cs.GetLatestHeight(), statePeriod, latestPeriod)

	if statePeriod == latestPeriod {
		latestHeight := cs.GetLatestHeight().(clienttypes.Height)
		res, err := pr.beaconClient.GetLightClientUpdate(statePeriod)
		if err != nil {
			return nil, err
		}
		root, err := res.Data.FinalizedHeader.Beacon.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		bootstrapRes, err := pr.beaconClient.GetBootstrap(root[:])
		if err != nil {
			return nil, err
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
		headers              []core.Header
		trustedSyncCommittee *lctypes.SyncCommittee
		trustedHeight        = cs.GetLatestHeight().(clienttypes.Height)
	)

	for p := statePeriod; p < latestPeriod; p++ {
		var header *lctypes.Header
		if p == statePeriod {
			header, err = pr.buildNextSyncCommitteeUpdateForCurrent(statePeriod, trustedHeight)
			if err != nil {
				return nil, err
			}
		} else {
			header, err = pr.buildNextSyncCommitteeUpdateForNext(p, trustedHeight)
			if err != nil {
				return nil, err
			}
		}
		headers = append(headers, header)
		trustedHeight = clienttypes.NewHeight(0, header.ExecutionUpdate.BlockNumber)
		trustedSyncCommittee = header.ConsensusUpdate.NextSyncCommittee
	}

	lfh.TrustedSyncCommittee = &lctypes.TrustedSyncCommittee{
		TrustedHeight: &trustedHeight,
		SyncCommittee: trustedSyncCommittee,
		IsNext:        true,
	}
	headers = append(headers, lfh)
	return headers, nil
}

func (pr *Prover) buildInitialState(blockNumber uint64) (*InitialState, error) {
	res, err := pr.beaconClient.GetLightClientFinalityUpdate()
	if err != nil {
		return nil, fmt.Errorf("failed to get light-client finality update: %v", err)
	}
	if eh := &res.Data.FinalizedHeader.Execution; blockNumber == 0 {
		blockNumber = eh.BlockNumber
	} else if eh.BlockNumber < blockNumber {
		return nil, fmt.Errorf("the height is not finalized yet: blockNumber=%v finalized_block_number=%v", blockNumber, eh.BlockNumber)
	}

	timestamp, err := pr.chain.Timestamp(pr.newHeight(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get timestamp: %v", err)
	}
	if truncatedTm := timestamp.Truncate(time.Second); truncatedTm != timestamp {
		return nil, fmt.Errorf("ethereum timestamp must be truncated to seconds: timestamp=%v truncated_timestamp=%v", timestamp, truncatedTm)
	}

	slot, err := pr.getSlotAtTimestamp(uint64(timestamp.Unix()))
	if err != nil {
		return nil, fmt.Errorf("failed to compute slot at timestamp: %v", err)
	}

	var currentSyncCommittee *lctypes.SyncCommittee
	if period := pr.computeSyncCommitteePeriod(pr.computeEpoch(slot)); period == 0 {
		// if period == 0, get the current sync committee from the bootstrap data corresponding to block at slot 1
		res2, err := pr.beaconClient.GetBlockRoot(1, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get block root: %v", err)
		}
		bootstrap, err := pr.beaconClient.GetBootstrap(res2.Data.Root[:])
		if err != nil {
			return nil, fmt.Errorf("failed to get bootstrap: %v", err)
		}
		currentSyncCommittee = bootstrap.Data.CurrentSyncCommittee.ToProto()
	} else {
		// otherwise, get the current sync committee from the light-client update
		update, err := pr.beaconClient.GetLightClientUpdate(period - 1)
		if err != nil {
			return nil, fmt.Errorf("failed to get light-client update: %v", err)
		}
		currentSyncCommittee = update.Data.NextSyncCommittee.ToProto()
	}
	accountUpdate, err := pr.buildAccountUpdate(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to build account update: %v", err)
	}
	var accountStorageRoot [32]byte
	copy(accountStorageRoot[:], accountUpdate.AccountStorageRoot)
	genesis, err := pr.beaconClient.GetGenesis()
	if err != nil {
		return nil, err
	}
	return &InitialState{
		Genesis:              *genesis,
		Slot:                 slot,
		BlockNumber:          blockNumber,
		AccountStorageRoot:   accountStorageRoot,
		Timestamp:            timestamp,
		CurrentSyncCommittee: *currentSyncCommittee,
	}, nil
}

// buildLatestFinalizedHeader returns the latest finalized header on this chain
// The returned header is expected to be the latest one of headers that can be verified by the light client
func (pr *Prover) GetLatestFinalizedHeader() (headers core.Header, err error) {
	res, err := pr.beaconClient.GetLightClientFinalityUpdate()
	if err != nil {
		return nil, err
	}
	lcUpdate := res.Data.ToProto()
	executionHeader := &res.Data.FinalizedHeader.Execution
	executionUpdate, err := pr.buildExecutionUpdate(executionHeader)
	if err != nil {
		return nil, err
	}
	executionRoot, err := executionHeader.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(executionRoot[:], lcUpdate.FinalizedExecutionRoot) {
		return nil, fmt.Errorf("execution root mismatch: %X != %X", executionRoot, lcUpdate.FinalizedExecutionRoot)
	}

	accountUpdate, err := pr.buildAccountUpdate(executionHeader.BlockNumber)
	if err != nil {
		return nil, err
	}
	return &lctypes.Header{
		ConsensusUpdate: lcUpdate,
		ExecutionUpdate: executionUpdate,
		AccountUpdate:   accountUpdate,
		Timestamp:       executionHeader.Timestamp,
	}, nil
}

func (pr *Prover) CheckRefreshRequired(counterparty core.ChainInfoICS02Querier) (bool, error) {
	cpQueryHeight, err := counterparty.LatestHeight()
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the counterparty chain: %v", err)
	}
	cpQueryCtx := core.NewQueryContext(context.TODO(), cpQueryHeight)

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

	selfQueryHeight, err := pr.chain.LatestHeight()
	if err != nil {
		return false, fmt.Errorf("failed to get the latest height of the self chain: %v", err)
	}

	selfTimestamp, err := pr.chain.Timestamp(selfQueryHeight)
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

func (pr *Prover) newClientState() *lctypes.ClientState {
	var commitmentsSlot [32]byte
	ibcAddress := pr.chain.Config().IBCAddress()

	return &lctypes.ClientState{
		ForkParameters:               pr.config.getForkParameters(),
		SecondsPerSlot:               pr.secondsPerSlot(),
		SlotsPerEpoch:                pr.slotsPerEpoch(),
		EpochsPerSyncCommitteePeriod: pr.epochsPerSyncCommitteePeriod(),

		MinSyncCommitteeParticipants: 1,

		IbcAddress:         ibcAddress.Bytes(),
		IbcCommitmentsSlot: commitmentsSlot[:],
		TrustLevel: &lctypes.Fraction{
			Numerator:   2,
			Denominator: 3,
		},
		TrustingPeriod: pr.config.GetTrustingPeriod(),
		MaxClockDrift:  pr.config.GetMaxClockDrift(),
	}
}

func (pr *Prover) buildNextSyncCommitteeUpdateForCurrent(period uint64, trustedHeight clienttypes.Height) (*lctypes.Header, error) {
	res, err := pr.beaconClient.GetLightClientUpdate(period)
	if err != nil {
		return nil, err
	}
	lcUpdate := res.Data.ToProto()
	executionHeader := &res.Data.FinalizedHeader.Execution
	executionUpdate, err := pr.buildExecutionUpdate(executionHeader)
	if err != nil {
		return nil, err
	}
	executionRoot, err := executionHeader.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(executionRoot[:], lcUpdate.FinalizedExecutionRoot) {
		return nil, fmt.Errorf("execution root mismatch: %X != %X", executionRoot, lcUpdate.FinalizedExecutionRoot)
	}

	accountUpdate, err := pr.buildAccountUpdate(executionHeader.BlockNumber)
	if err != nil {
		return nil, err
	}

	root, err := res.Data.FinalizedHeader.Beacon.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	bootstrapRes, err := pr.beaconClient.GetBootstrap(root[:])
	if err != nil {
		return nil, err
	}

	return &lctypes.Header{
		TrustedSyncCommittee: &lctypes.TrustedSyncCommittee{
			TrustedHeight: &trustedHeight,
			SyncCommittee: bootstrapRes.Data.CurrentSyncCommittee.ToProto(),
			IsNext:        false,
		},
		ConsensusUpdate: lcUpdate,
		ExecutionUpdate: executionUpdate,
		AccountUpdate:   accountUpdate,
		Timestamp:       executionHeader.Timestamp,
	}, nil
}

func (pr *Prover) buildNextSyncCommitteeUpdateForNext(period uint64, trustedHeight clienttypes.Height) (*lctypes.Header, error) {
	res, err := pr.beaconClient.GetLightClientUpdate(period)
	if err != nil {
		return nil, err
	}
	lcUpdate := res.Data.ToProto()
	executionHeader := &res.Data.FinalizedHeader.Execution
	executionUpdate, err := pr.buildExecutionUpdate(executionHeader)
	if err != nil {
		return nil, err
	}
	executionRoot, err := executionHeader.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(executionRoot[:], lcUpdate.FinalizedExecutionRoot) {
		return nil, fmt.Errorf("execution root mismatch: %X != %X", executionRoot, lcUpdate.FinalizedExecutionRoot)
	}

	accountUpdate, err := pr.buildAccountUpdate(executionHeader.BlockNumber)
	if err != nil {
		return nil, err
	}
	return &lctypes.Header{
		TrustedSyncCommittee: &lctypes.TrustedSyncCommittee{
			TrustedHeight: &trustedHeight,
			SyncCommittee: lcUpdate.NextSyncCommittee,
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
	proof, err := pr.buildStateProof([]byte(path), proofHeight)
	return proof, height, err
}

func (pr *Prover) newHeight(blockNumber int64) clienttypes.Height {
	return clienttypes.NewHeight(0, uint64(blockNumber))
}

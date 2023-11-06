package relay

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

var _ core.LightClient = (*Prover)(nil)

// CreateMsgCreateClient creates a MsgCreateClient for the counterparty chain
func (pr *Prover) CreateMsgCreateClient(_ string, selfHeader core.Header, signer sdk.AccAddress) (*clienttypes.MsgCreateClient, error) {
	header, ok := selfHeader.(*lctypes.Header)
	if !ok {
		return nil, fmt.Errorf("unexpected header type: %T", selfHeader)
	}

	genesis, err := pr.beaconClient.GetGenesis()
	if err != nil {
		return nil, err
	}
	clientState := pr.newClientState()
	clientState.GenesisValidatorsRoot = genesis.GenesisValidatorsRoot[:]
	clientState.GenesisTime = genesis.GenesisTimeSeconds
	clientState.LatestSlot = uint64(header.ConsensusUpdate.FinalizedHeader.Slot)
	clientState.LatestExecutionBlockNumber = header.ExecutionUpdate.BlockNumber

	res, err := pr.beaconClient.GetLightClientUpdate(pr.computeSyncCommitteePeriod(pr.computeEpoch(header.ConsensusUpdate.FinalizedHeader.Slot)))
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
	var currentSyncCommitteePubKey []byte
	if header.TrustedSyncCommittee != nil && !bytes.Equal(bootstrapRes.Data.CurrentSyncCommittee.AggregatePubKey, header.TrustedSyncCommittee.SyncCommittee.AggregatePubkey) {
		return nil, fmt.Errorf("the trusted sync committee is not matched: trusted=%X current=%X", header.TrustedSyncCommittee.SyncCommittee.AggregatePubkey, bootstrapRes.Data.CurrentSyncCommittee.AggregatePubKey)
	} else {
		currentSyncCommitteePubKey = bootstrapRes.Data.CurrentSyncCommittee.AggregatePubKey
	}
	accountUpdate, err := pr.buildAccountUpdate(header.ExecutionUpdate.BlockNumber)
	if err != nil {
		return nil, err
	}

	consensusState := &lctypes.ConsensusState{
		Slot:                 clientState.LatestSlot,
		StorageRoot:          accountUpdate.AccountStorageRoot,
		Timestamp:            time.Unix(int64(header.Timestamp), 0),
		CurrentSyncCommittee: currentSyncCommitteePubKey,
	}

	return clienttypes.NewMsgCreateClient(clientState, consensusState, signer.String())
}

// SetupHeadersForUpdate returns the finalized header and any intermediate headers needed to apply it to the client on the counterpaty chain
// The order of the returned header slice should be as: [<intermediate headers>..., <update header>]
// if the header slice's length == 0 and err == nil, the relayer should skips the update-client
func (pr *Prover) SetupHeadersForUpdate(counterparty core.FinalityAwareChain, latestFinalizedHeader core.Header) ([]core.Header, error) {
	lfh := latestFinalizedHeader.(*lctypes.Header)

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

	statePeriod, err := pr.findPeriodByBlockNumber(cs.GetLatestHeight().GetRevisionHeight(), lfh.ConsensusUpdate.FinalizedHeader.Slot)
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

// GetLatestFinalizedHeader returns the latest finalized header on this chain
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
	executionHeader := res.Data.FinalizedHeader.Execution
	executionUpdate, err := pr.buildExecutionUpdate(&executionHeader)
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

package beacon

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hyperledger-labs/yui-relayer/log"
)

type Client struct {
	endpoint string
}

func NewClient(endpoint string) Client {
	return Client{endpoint: endpoint}
}

func (Client) SupportedVersion() string {
	return "capella"
}

func (cl Client) GetGenesis() (*Genesis, error) {
	var res GenesisResponse
	if err := cl.get("/eth/v1/beacon/genesis", &res); err != nil {
		return nil, err
	}
	return ToGenesis(res)
}

func (cl Client) GetBlockRoot(slot uint64, allowOptimistic bool) (*BlockRootResponse, error) {
	var res BlockRootResponse
	if err := cl.get(fmt.Sprintf("/eth/v1/beacon/blocks/%v/root", slot), &res); err != nil {
		return nil, err
	}
	if !allowOptimistic && res.ExecutionOptimistic {
		return nil, fmt.Errorf("optimistic execution not allowed")
	}
	return &res, nil
}

func (cl Client) GetFinalityCheckpoints() (*StateFinalityCheckpoints, error) {
	var res StateFinalityCheckpointResponse
	if err := cl.get("/eth/v1/beacon/states/head/finality_checkpoints", &res); err != nil {
		return nil, err
	}
	return ToStateFinalityCheckpoints(res)
}

func (cl Client) GetBootstrap(finalizedRoot []byte) (*LightClientBootstrapResponse, error) {
	if len(finalizedRoot) != 32 {
		return nil, fmt.Errorf("finalizedRoot length must be 32: actual=%v", finalizedRoot)
	}
	var res LightClientBootstrapResponse
	if err := cl.get(fmt.Sprintf("/eth/v1/beacon/light_client/bootstrap/0x%v", hex.EncodeToString(finalizedRoot[:])), &res); err != nil {
		return nil, err
	}
	if res.Version != cl.SupportedVersion() {
		return nil, fmt.Errorf("unsupported version: %v", res.Version)
	}
	return &res, nil
}

func (cl Client) GetLightClientUpdates(period uint64, count uint64) (LightClientUpdatesResponse, error) {
	var res LightClientUpdatesResponse
	if err := cl.get(fmt.Sprintf("/eth/v1/beacon/light_client/updates?start_period=%v&count=%v", period, count), &res); err != nil {
		return nil, err
	}
	if len(res) != int(count) {
		return nil, fmt.Errorf("unexpected response length: expected=%v actual=%v", count, len(res))
	}
	for i := range res {
		if res[i].Version != cl.SupportedVersion() {
			return nil, fmt.Errorf("unsupported version: %v", res[i].Version)
		}
	}
	return res, nil
}

func (cl Client) GetLightClientUpdate(period uint64) (*LightClientUpdateResponse, error) {
	res, err := cl.GetLightClientUpdates(period, 1)
	if err != nil {
		return nil, err
	}
	return &res[0], nil
}

func (cl Client) GetLightClientFinalityUpdate() (*LightClientFinalityUpdateResponse, error) {
	var res LightClientFinalityUpdateResponse
	if err := cl.get("/eth/v1/beacon/light_client/finality_update", &res); err != nil {
		return nil, err
	}
	if res.Version != cl.SupportedVersion() {
		return nil, fmt.Errorf("unsupported version: %v", res.Version)
	}
	return &res, nil
}

func (cl Client) get(path string, res any) error {
	log.GetLogger().Debug("Beacon API request", "endpoint", cl.endpoint+path)
	r, err := http.Get(cl.endpoint + path)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	bz, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, &res)
}

package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/pkg/errors"
	"strings"
)

// LogByTx - get logs transaction by id.
func (p *Params) LogByTx(id string) (log *Log, err error) {

	log = new(Log)

	switch p.platform {
	case pbspot.Platform_ETHEREUM:
		p.query = []string{"-X", "POST", "--data", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%v"],"id":1}`, id), p.rpc}
	case pbspot.Platform_TRON:
		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/gettransactioninfobyid", p.rpc), "-d", fmt.Sprintf(`{"value": "%v"}`, id)}
	default:
		return log, errors.New("method not found!...")
	}

	if err := p.commit(); err != nil {
		return log, err
	}

	return p.log()
}

// log - log to new struct compare.
func (p *Params) log() (log *Log, err error) {

	log = new(Log)

	switch p.platform {
	case pbspot.Platform_ETHEREUM:

		if result, ok := p.response["result"].(map[string]interface{}); ok {

			if vlog, ok := result["logs"]; ok {

				if logs, ok := vlog.([]interface{}); ok {
					for i := 0; i < len(logs); i++ {

						if data, ok := logs[i].(map[string]interface{})["data"]; ok {

							data, err := hex.DecodeString(strings.TrimPrefix(data.(string), "0x"))
							if err != nil {
								return log, nil
							}

							log.Data = data
						}

						if topics, ok := logs[i].(map[string]interface{})["topics"]; ok {
							log.Topics = topics.([]interface{})
						}
					}
				}
			}
		}

	case pbspot.Platform_TRON:

		if err, ok := p.response["Error"]; ok || err != nil {
			return log, errors.New(err.(string))
		}

		if vlog, ok := p.response["log"]; ok {

			if logs, ok := vlog.([]interface{}); ok {
				for i := 0; i < len(logs); i++ {

					if data, ok := logs[i].(map[string]interface{})["data"]; ok {

						data, err := hex.DecodeString(data.(string))
						if err != nil {
							return log, nil
						}

						log.Data = data
					}

					if topics, ok := logs[i].(map[string]interface{})["topics"]; ok {
						log.Topics = topics.([]interface{})
					}
				}
			}
		}
	}

	return log, nil
}

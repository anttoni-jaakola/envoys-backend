package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/pkg/errors"
	"strings"
)

// BlockByNumber - get block by number.
func (p *Params) BlockByNumber(number int64) (block *Block, err error) {

	block = new(Block)

	switch p.platform {
	case pbspot.Platform_ETHEREUM:
		p.query = []string{"-X", "POST", "--data", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":[%d, true],"id":1}`, number), p.rpc}
	case pbspot.Platform_TRON:
		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/getblockbynum", p.rpc), "-d", fmt.Sprintf(`{"num": %d}`, number)}
	default:
		return block, errors.New("method not found!...")
	}

	if err := p.commit(); err != nil {
		return block, err
	}

	return p.block()
}

// block - block to new struct compare.
func (p *Params) block() (block *Block, err error) {

	block = new(Block)

	switch p.platform {
	case pbspot.Platform_ETHEREUM:

		if result, ok := p.response["result"].(map[string]interface{}); ok {

			if transactions, ok := result["transactions"].([]interface{}); ok {
				for i := 0; i < len(transactions); i++ {
					if input, ok := transactions[i].(map[string]interface{})["input"]; ok {
						if strings.Contains(input.(string), "0xa9059cbb") {
							transactions[i].(map[string]interface{})["type"] = TypeContract
						} else {
							transactions[i].(map[string]interface{})["type"] = TypeInternal
						}
					}
				}
			}

			serialize, err := json.Marshal(result)
			if err != nil {
				return block, err
			}

			if err := json.Unmarshal(serialize, &block); err != nil {
				return block, err
			}

		} else {
			return block, errors.New("block not found!...")
		}

	case pbspot.Platform_TRON:

		if err, ok := p.response["Error"]; ok || err != nil {
			return block, errors.New(err.(string))
		}

		if hash, ok := p.response["blockID"]; ok {
			block.Hash = hash.(string)
		}

		if _, ok := p.response["block_header"]; ok {
			if raws, ok := p.response["block_header"].(map[string]interface{})["raw_data"].(map[string]interface{}); ok {
				block.TransactionsRoot = raws["txTrieRoot"].(string)
				block.ParentHash = raws["parentHash"].(string)
			}
		}

		if transactions, ok := p.response["transactions"].([]interface{}); ok {

			for i := 0; i < len(transactions); i++ {

				var (
					column Transaction
				)

				if transaction, ok := transactions[i].(map[string]interface{})["raw_data"].(map[string]interface{})["contract"].([]interface{}); ok {

					for i := 0; i < len(transaction); i++ {

						if value, ok := transaction[i].(map[string]interface{})["parameter"].(map[string]interface{})["value"].(map[string]interface{}); ok {

							if amount, ok := value["amount"]; ok {
								column.Value = fmt.Sprintf("%v", amount.(float64))
							}
							if from, ok := value["owner_address"]; ok {
								column.From = from.(string)
							}
							if to, ok := value["to_address"]; ok {
								column.To = to.(string)
							} else if to, ok = value["contract_address"]; ok {
								column.To = to.(string)
							}
							if data, ok := value["data"]; ok {

								data, err := hex.DecodeString(data.(string))
								if err != nil {
									return block, nil
								}

								column.Data = data
							}
						}

						if types, ok := transaction[i].(map[string]interface{})["type"]; ok {

							switch types.(string) {
							case "TransferContract":
								column.Type = TypeInternal
							case "TriggerSmartContract":
								column.Type = TypeContract
							}

						}
					}
				}

				if hash, ok := transactions[i].(map[string]interface{}); ok {
					column.Hash = hash["txID"].(string)
				}

				if column.Type == TypeInternal || column.Type == TypeContract {
					block.Transactions = append(block.Transactions, &column)
				}
			}
		}
	}

	return block, nil
}

// Status - Transaction status.
func (p *Params) Status(tx string) (success bool) {

	switch p.platform {
	case pbspot.Platform_ETHEREUM:
		p.query = []string{"-X", "POST", "--data", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%s"]}`, tx), p.rpc}
	case pbspot.Platform_TRON:

		request := struct {
			Value string `json:"value"`
		}{
			Value: tx,
		}

		marshal, err := json.Marshal(request)
		if err != nil {
			return success
		}

		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/gettransactioninfobyid", p.rpc), "-d", string(marshal)}
	}

	if err := p.commit(); err != nil {
		return success
	}

	if result, ok := p.response["result"]; ok {

		if maps, ok := result.(map[string]interface{}); ok {

			// ETHEREUM status: QUANTITY either 1 (success) or 0 (failure).
			if maps["status"].(string) == "0x0" {
				return success
			}
		} else {

			// TRON status: SUCCESS (success) or FAILED (failure).
			if result.(string) == "FAILED" {
				return success
			}
		}
	}

	return true
}

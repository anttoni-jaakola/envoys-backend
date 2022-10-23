package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/address"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

func (p *Params) Private(private *ecdsa.PrivateKey) {
	p.private = private
}

// Transfer - new transfer to.
func (p *Params) Transfer(tx *Transfer) (hash string, err error) {

	public, ok := p.private.Public().(*ecdsa.PublicKey)
	if !ok {
		return hash, errors.New("error casting public key to ECDSA")
	}
	owner := crypto.PubkeyToAddress(*public)

	switch p.platform {
	case RpcEthereum:
		p.query = []string{"-X", "POST", "--data", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendTransaction","params":["%v"],"id":1}`, "id"), p.rpc}
	case RpcTron:

		if len(tx.Contract) > 0 {

			request := struct {
				ContractAddress  string `json:"contract_address"`
				FunctionSelector string `json:"function_selector"`
				Parameter        string `json:"parameter"`
				FeeLimit         int    `json:"fee_limit"`
				OwnerAddress     string `json:"owner_address"`
			}{
				ContractAddress:  tx.Contract,
				FunctionSelector: "transfer(address,uint256)",
				Parameter:        strings.TrimPrefix(hexutil.Encode(tx.Data), "0x"),
				FeeLimit:         tx.Gas,
				OwnerAddress:     address.New(owner.String()).Hex(true),
			}

			marshal, err := json.Marshal(request)
			if err != nil {
				return "", err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/triggersmartcontract", p.rpc), "-d", string(marshal)}
		} else {

			request := struct {
				ToAddress    string   `json:"to_address"`
				OwnerAddress string   `json:"owner_address"`
				Amount       *big.Int `json:"amount"`
			}{
				ToAddress:    tx.To,
				OwnerAddress: address.New(owner.String()).Hex(true),
				Amount:       tx.Value,
			}

			marshal, err := json.Marshal(request)
			if err != nil {
				return "", err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/createtransaction", p.rpc), "-d", string(marshal)}
		}

	default:
		return hash, errors.New("method not found!...")
	}

	if err := p.commit(); err != nil {
		return hash, err
	}

	return p.signature()
}

// AccountResource - account resource.
func (p *Params) AccountResource() (energy int64, err error) {

	public, ok := p.private.Public().(*ecdsa.PublicKey)
	if !ok {
		return energy, errors.New("error casting public key to ECDSA")
	}
	owner := crypto.PubkeyToAddress(*public)

	request := struct {
		Address string `json:"address"`
	}{
		Address: address.New(owner.String()).Hex(true),
	}

	marshal, err := json.Marshal(request)
	if err != nil {
		return energy, err
	}

	p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/getaccountresource", p.rpc), "-d", string(marshal)}

	resource, err := p.get()
	if err != nil {
		return energy, err
	}

	if freeNetLimit, ok := resource["freeNetLimit"]; ok {

		freeNetUsed, ok := resource["freeNetUsed"]
		if !ok {
			freeNetUsed = float64(0)
		}
		netLimit, ok := resource["NetLimit"]
		if !ok {
			netLimit = float64(0)
		}
		netUsed, ok := resource["NetUsed"]
		if !ok {
			netUsed = float64(0)
		}

		return int64(freeNetLimit.(float64) + netLimit.(float64) - freeNetUsed.(float64) - netUsed.(float64)), nil
	}

	return energy, nil
}

// EstimateGas - use fee transaction.
func (p *Params) EstimateGas(tx *Transfer) (fee int64, err error) {

	var (
		gas int
	)

	public, ok := p.private.Public().(*ecdsa.PublicKey)
	if !ok {
		return fee, errors.New("error casting public key to ECDSA")
	}
	owner := crypto.PubkeyToAddress(*public)

	switch p.platform {
	case RpcEthereum:
	case RpcTron:

		if len(tx.Contract) > 0 {

			request := struct {
				ContractAddress  string `json:"contract_address"`
				FunctionSelector string `json:"function_selector"`
				Parameter        string `json:"parameter"`
				FeeLimit         int    `json:"fee_limit"`
				OwnerAddress     string `json:"owner_address"`
			}{
				ContractAddress:  tx.Contract,
				FunctionSelector: "transfer(address,uint256)",
				Parameter:        strings.TrimPrefix(hexutil.Encode(tx.Data), "0x"),
				FeeLimit:         tx.Gas,
				OwnerAddress:     address.New(owner.String()).Hex(true),
			}

			marshal, err := json.Marshal(request)
			if err != nil {
				return fee, err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/triggerconstantcontract", p.rpc), "-d", string(marshal)}

		} else {

			request := struct {
				ToAddress    string   `json:"to_address"`
				OwnerAddress string   `json:"owner_address"`
				Amount       *big.Int `json:"amount"`
			}{
				ToAddress:    tx.To,
				OwnerAddress: address.New(owner.String()).Hex(true),
				Amount:       tx.Value,
			}

			marshal, err := json.Marshal(request)
			if err != nil {
				return fee, err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/createtransaction", p.rpc), "-d", string(marshal)}
		}

	default:
		return fee, errors.New("method not found!...")
	}

	constant, err := p.get()
	if err != nil {
		return fee, err
	}

	if raw, ok := constant["raw_data_hex"]; ok {
		price, err := p.AccountResource()
		if err != nil {
			return fee, err
		}

		fee = decimal.FromInt(int64(len(raw.(string)))).Mul(decimal.FromInt(10)).Int64()
		if price >= decimal.FromInt(fee).Div(decimal.FromInt(10)).Int64() {
			fee = 0
		}

		return fee, nil
	}

	if transaction, ok := constant["transaction"]; ok {
		if transaction, ok := transaction.(map[string]interface{}); ok {

			if _, ok := transaction["txID"]; ok {

				price, err := p.AccountResource()
				if err != nil {
					return fee, err
				}

				signature, err := crypto.Sign(common.Hex2Bytes(transaction["txID"].(string)), p.private)
				if err != nil {
					return fee, err
				}

				if _, ok := transaction["raw_data_hex"]; ok {

					raw, err := hex.DecodeString(transaction["raw_data_hex"].(string))
					if err != nil {
						return fee, err
					}

					gas += len(raw)
				}
				gas += len(signature)

				if energy, ok := constant["energy_used"]; ok {

					fee = decimal.FromInt(9 + 60 + int64(energy.(float64)*10) + int64(gas)).Mul(decimal.FromInt(10)).Int64()
					bandwidth := decimal.FromInt(fee).Sub(decimal.FromInt(int64(energy.(float64) * 100))).Int64()

					if price >= decimal.FromInt(bandwidth).Div(decimal.FromInt(10)).Int64() {
						fee = decimal.FromInt(fee).Sub(decimal.FromInt(bandwidth)).Int64()
					}

					return fee, nil
				}
			}
		}
	}

	return fee, errors.New("constant contract not found!...")
}

// status - transaction status.
func (p *Params) status(tx string) error {

	switch p.platform {
	case RpcEthereum:
	case RpcTron:

		request := struct {
			Value string `json:"value"`
		}{
			Value: tx,
		}

		marshal, err := json.Marshal(request)
		if err != nil {
			return err
		}

		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/gettransactioninfobyid", p.rpc), "-d", string(marshal)}
	}

	if err := p.commit(); err != nil {
		return err
	}

	if result, ok := p.response["result"]; ok {
		if result.(string) == "FAILED" {
			message, err := hex.DecodeString(p.response["resMessage"].(string))
			if err != nil {
				return err
			}
			return errors.New(fmt.Sprintf("[%v] - %v", result.(string), string(message)))
		}
	}

	return nil
}

// signature - signing the transfer with a private key.
func (p *Params) signature() (txID string, err error) {

	// Signing the transfer with a private key.
	switch p.platform {
	case RpcEthereum:
		return txID, nil
	case RpcTron:

		if err, ok := p.response["Error"]; ok {
			return txID, errors.New(err.(string))
		}

		if _, ok := p.response["result"]; ok {
			if transaction, ok := p.response["transaction"]; ok {
				p.response = transaction.(map[string]interface{})
			}
		} else if _, ok = p.response["txID"]; !ok {
			return txID, errors.New("map[string]interface{} not recognized")
		}

		if _, ok := p.response["txID"]; ok {
			signature, err := crypto.Sign(common.Hex2Bytes(p.response["txID"].(string)), p.private)
			if err != nil {
				return txID, err
			}
			p.response["signature"] = []string{common.Bytes2Hex(signature)}
			p.success = true

			return p.response["txID"].(string), nil
		}
	}

	return txID, errors.New("invalid characters encountered in Hex string")
}

// Transaction - new transaction to blockchain.
func (p *Params) Transaction() error {

	if !p.success {
		return errors.New("transfer function has not been initialized")
	}

	switch p.platform {
	case RpcEthereum:
	case RpcTron:

		if transaction, ok := p.response["transaction"]; ok {
			p.response = transaction.(map[string]interface{})
		} else if _, ok = p.response["txID"]; !ok {
			return errors.New("map[string]interface{} not recognized")
		}

		serialize, err := json.Marshal(p.response)
		if err != nil {
			return err
		}

		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/broadcasttransaction", p.rpc), "-d", string(serialize)}
	}

	if err := p.commit(); err != nil {
		return err
	}
	p.success = false

	if code, ok := p.response["code"]; ok {
		message, err := hex.DecodeString(p.response["message"].(string))
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("[%v], %v", code.(string), string(message)))
	}

	return nil
}

// Data - data transfer.
func (p *Params) Data(to string, amount []byte) (data []byte, err error) {

	switch p.platform {
	case RpcEthereum:

		decode, err := hex.DecodeString(address.New(to).Hex())
		if err != nil {
			return nil, err
		}

		data = append(data, help.SignatureKeccak256([]byte("transfer(address,uint256)"))...)
		data = append(data, common.LeftPadBytes(decode, 32)...)
		data = append(data, common.LeftPadBytes(amount, 32)...)

	case RpcTron:

		decode, err := hex.DecodeString(address.New(to, true).Hex(true)[2:])
		if err != nil {
			return nil, err
		}

		data = append(data, common.LeftPadBytes(decode, 32)...)
		data = append(data, common.LeftPadBytes(amount, 32)...)

	}

	return data, nil
}

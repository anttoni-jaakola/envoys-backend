package blockchain

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/pkg/errors"
	"math/big"
	"os/exec"
)

const (
	TypeInternal = 0x1000a
	TypeContract = 0x2000a
)

type Transfer struct {
	Hash     string
	Contract string
	Nonce    int
	From     string
	To       string
	Value    *big.Int
	Gas      int
	GasPrice int
	Data     []byte
}

type Block struct {
	TransactionsRoot string
	Hash             string
	ParentHash       string
	Transactions     []*Transaction
}

type Log struct {
	Data   []byte
	Topics []interface{}
}

type Transaction struct {
	From  string
	To    string
	Hash  string
	Value string
	Type  int
	Data  []byte
}

type Params struct {
	rpc      string
	platform pbspot.Platform
	response map[string]interface{}
	query    []string
	private  *ecdsa.PrivateKey
	success  bool
}

// Dial - connect to blockchain.
func Dial(rpc string, platform pbspot.Platform) *Params {
	return &Params{
		rpc:      rpc,
		platform: platform,
	}
}

// commit - new request to blockchain.
func (p *Params) commit() error {

	response, err := p.get()
	if err != nil {
		return err
	}
	p.response = response

	return nil
}

// get - new request.
func (p *Params) get() (response map[string]interface{}, err error) {

	cmd, err := exec.Command("curl", p.query...).Output()
	if err != nil {
		return response, err
	}

	if err = json.Unmarshal(cmd, &response); err != nil {
		return response, err
	}

	if len(response) == 0 {
		return response, errors.New("map[] was not found!...")
	}

	return response, nil
}

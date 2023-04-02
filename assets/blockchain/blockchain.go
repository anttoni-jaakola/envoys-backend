package blockchain

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/pkg/errors"
	"math/big"
	"os/exec"
)

// The purpose of these constants is to provide a way to distinguish between different types of data. By assigning each
// type a unique numerical value, it makes it easier to identify and organize the data. For example, if you have a system
// that stores information about both internal and contract data, you can use these constants to differentiate between them.
const (
	TypeInternal = 0x1000a
	TypeContract = 0x2000a
)

// Transfer - The purpose of the above struct is to store details of a transaction that has occurred on the Ethereum blockchain. The
// struct is used to track the hash of the transaction, the contract involved, the sender and receiver addresses, the
// value of the transaction in BigInt, the amount of gas used, the gas price, and any additional data associated with the transaction.
type Transfer struct {
	Hash     string
	Contract string
	From     string
	To       string
	Value    *big.Int
	Gas      int
	GasPrice int
	Data     []byte
}

// Block - The purpose of the following block struct is to provide a structure for storing information about a block in a
// blockchain. It contains the transactions root, hash, parent hash, and a list of transactions associated with the
// block. This information can then be used to validate the data stored in the block, as well as to ensure that the block
// is properly linked to the rest of the blockchain.
type Block struct {
	TransactionsRoot string
	Hash             string
	ParentHash       string
	Transactions     []*Transaction
}

// The Log struct is used to represent a log in the Ethereum/Tron blockchain. It contains two fields: Data, which is an array
// of bytes that holds the log data, and Topics, which is an array of interfaces that are used to filter and categorize
// logs.
type Log struct {
	Data   []byte
	Topics []interface{}
}

// Transaction - The Transaction struct is used to represent a single transaction on a blockchain. It contains information such as the
// sender and receiver of the transaction, the hash of the transaction, the value of the transaction, the type of the
// transaction, and any additional data associated with the transaction. This information is used to track and validate
// transactions on the blockchain.
type Transaction struct {
	From  string
	To    string
	Hash  string
	Value string
	Type  int
	Data  []byte
}

// Params - This is a struct used to store data related to a specific function. It is used to store data that will be used in the
// function, as well as the results of the function. The data stored includes a rpc string, a platform, a response map, a
// query array, a private key, a network, a success boolean and a stop boolean.
type Params struct {
	rpc      string
	platform string
	response map[string]interface{}
	query    []string
	private  *ecdsa.PrivateKey
	network  *big.Int
	gas      uint64
	success  bool
	stop     bool
}

// Dial - The purpose of the code is to test the connection to the blockchain and then create a Params struct using the given
// parameters and return a pointer to the struct along with a nil error. This allows the code to make use of the struct
// while also ensuring that the operation was successful.
func Dial(rpc, platform string) (*Params, error) {

	// The code is used to test if the connection to the blockchain is successful. The help.Ping() function is used to test
	// the connection and the ok variable stores the result of the function. If ok is false, an error message is returned
	// indicating a connection error to the blockchain.
	if ok := help.Ping(rpc); !ok {
		return nil, errors.New("connect error to blockchain")
	}

	// This code is used to create a Params struct using the given rpc and platform parameters and return a pointer to the
	// struct along with a nil error. This allows the code to make use of the struct while also ensuring that the operation
	// was successful.
	return &Params{
		rpc:      rpc,
		platform: platform,
	}, nil
}

// commit - The purpose of this code is to make a get request with the p object, check for any errors, and store the response in
// the p object if no errors are encountered.
func (p *Params) commit() error {

	// This code is used to check for errors when making a get request with the p object. If an error is encountered, it is
	// returned and the response is stored in the p object.
	response, err := p.get()
	if err != nil {
		return err
	}
	p.response = response

	return nil
}

// get -The purpose of this code is to execute a command using the curl command, capture the output of that command, and
// unmarshal a JSON string into a response object. It also includes a defer statement that handles any panic that may
// occur in the function, and an if statement that checks for an empty response and returns an error if one is found.
func (p *Params) get() (response map[string]interface{}, err error) {

	// The purpose of this code is to handle any panic that may occur in the function. The defer statement will execute a
	// function after the surrounding function returns, in this case if a panic occurs, the recover() function will catch it
	// and return nil. This allows for the panic to be handled and the program to continue running instead of crashing.
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	// This code is used to execute a command using the curl command and pass the values stored in the variable p.query as
	// arguments. If the command fails, it returns an error.
	cmd := exec.Command("curl", p.query...)
	if cmd.Err != nil {
		return nil, cmd.Err
	}

	// The purpose of this code is to execute a command and capture the output of that command. The output variable will
	// hold the output of the command, and err will hold any error that may occur from executing the command. If there is an
	// error, the function will return the response and the error.
	output, err := cmd.Output()
	if err != nil {
		return response, err
	}

	// This code is used to unmarshal a JSON string into a response object. It is used to convert the JSON string into a Go
	// data structure so that it can be used in the application. The if statement checks for errors during the unmarshal
	// process, and if an error is found, it returns the response object and the error.
	if err = json.Unmarshal(output, &response); err != nil {
		return response, err
	}

	// The purpose of this code is to check if the response from a function is empty. If it is, an error is returned
	// indicating that a "map[] was not found."
	if len(response) == 0 {
		return response, errors.New("map[] was not found!...")
	}

	return response, nil
}

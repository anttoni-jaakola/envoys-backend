package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/address"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	core "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
	"strings"
)

// Private - This function sets the private key on the Params struct, which is a type of struct used to store data related to an
// operation. The function takes an ecdsa.PrivateKey as an argument and sets it to the p.private field of the Params
// struct. This is used to ensure that the private key associated with the operation is secure and accessible only by the user.
func (p *Params) Private(private *ecdsa.PrivateKey) {
	p.private = private
}

// Network - This function sets the "network" field of the Params struct to a big integer value, based on the given id int64. This
// function is used to set the network field of the Params struct so that it can be used to identify the network that the
// struct is associated with.
func (p *Params) Network(id int64) {
	p.network = big.NewInt(id)
}

// gasUsed - This function is used to return the amount of gas used by a certain platform. Depending on the platform, this amount
// can vary, and the boolean parameter is used to determine if the amount of gas used should be 65000 or 21000 for
// Ethereum. For Tron, the amount of gas used is always 10000000. If the platform is none of the listed, 0 is returned.
func (p *Params) gasUsed(c bool) uint64 {
	switch p.platform {

	case types.PlatformEthereum:
		if c {
			return 65000
		} else {
			return 21000
		}
	case types.PlatformTron:
		return 10000000
	}

	return 0
}

// Transfer - The purpose of this code is to transfer funds from one user to another on a blockchain platform. It checks the
// platform being used, creates and signs a transaction, and then sends it for processing. It also stores the result of
// the transaction in the p.response object, and then returns the transaction's hash and any error that may have occurred during the process.
func (p *Params) Transfer(tx *Transfer) (hash string, err error) {

	// This code is attempting to check if the `p.private.Public()` call returns a valid *ecdsa.PublicKey. If it does, it is
	// stored in the `public` variable, and the code can proceed. If it does not, an error is returned.
	public, ok := p.private.Public().(*ecdsa.PublicKey)
	if !ok {
		return hash, errors.New("error casting public key to ECDSA")
	}

	// Is used to convert a public key to an address. It takes the variable public
	// (which is a pointer to a public key) as an argument and returns an address. This can be used to identify a user on
	// the blockchain.
	owner := crypto.PubkeyToAddress(*public)

	// The switch statement is a control structure used to execute different blocks of code depending on the value of a
	// variable. In this example, the variable being tested is "p.platform", and depending on its value, a different block of code will be executed.
	switch p.platform {
	case types.PlatformEthereum:

		// This code is attempting to get the gas price from the 'p' object. If there is an error, the hash and error are returned.
		gasPrice, err := p.gasPrice()
		if err != nil {
			return hash, err
		}

		// The purpose of this code is to get the nonce associated with the owner (which is a String) and then check for any
		// errors. If there is an error, it will return the hash and the error.
		nonce, err := p.getNonce(owner.String())
		if err != nil {
			return hash, err
		}

		// This if statement checks the length of the tx.Contract value. If it is greater than 0, the code within the statement
		// will be executed. This statement is used to determine if there is any data in the tx.Contract value, and if so,
		// execute a certain set of operations.
		if len(tx.Contract) > 0 {

			// The purpose of the following is to convert a hexadecimal string into a common address using the HexToAddress()
			// function. The tx.Contract data is converted into a common address that can be used in a contract.
			to := common.HexToAddress(tx.Contract)

			// This code is creating and signing a new Ethereum transaction with the given parameters. The parameters include the
			// address to send the transaction to, the amount of gas to use, the gas price to use, and the transaction data. Once
			// the transaction is created and signed, it is sent for processing.
			transfer, err := core.SignNewTx(p.private, core.NewEIP155Signer(p.network), &core.LegacyTx{
				Nonce:    nonce.Uint64(),
				To:       &to,
				Value:    big.NewInt(0),
				Gas:      p.gasUsed(true),
				GasPrice: big.NewInt(gasPrice),
				Data:     tx.Data,
			})
			if err != nil {
				return hash, err
			}

			// MarshalBinary() is a method that encodes a transfer object into a binary form. The purpose of the code above is to
			// call the MarshalBinary() method, which will return a hash and an error if it fails. If there is an error, the
			// function will return the hash and the error.
			marshal, err := transfer.MarshalBinary()
			if err != nil {
				return hash, err
			}

			// This code is assigning a key-value pair to the p.response object. The key is "result" and the value is an array
			// with two elements: the hex-encoded version of the marshal object and the string representation of the transfer
			// hash. This is likely being used to store or return the result of a calculation or set of operations.
			p.response = map[string]interface{}{
				"result": []string{
					hexutil.Encode(marshal),
					transfer.Hash().String(),
				},
			}

		} else {

			// The purpose of this code is to convert a hexadecimal value (represented as a string) to an address using the
			// common.HexToAddress() function. This is typically used when a transaction (tx) is taking place on a blockchain, as
			// the address is needed for the transaction to be successful.
			to := common.HexToAddress(tx.To)

			// This code is signing a new transaction (transfer) with the private key of the user (p.private). It is also creating
			// a new EIP155 signer (types.NewEIP155Signer(p.network)), and setting up the nonce, the address to transfer to (to),
			// the amount to transfer (tx.Value), the amount of gas to pay (tx.Gas) and the gas price to pay (gasPrice). If an
			// error occurs during the signing process, the code returns the hash and an error.
			transfer, err := core.SignNewTx(p.private, core.NewEIP155Signer(p.network), &core.LegacyTx{
				Nonce:    nonce.Uint64(),
				To:       &to,
				Value:    tx.Value,
				Gas:      p.gasUsed(false),
				GasPrice: big.NewInt(gasPrice),
			})
			if err != nil {
				return hash, err
			}

			// MarshalBinary() is a method that encodes a transfer object into a binary form. The purpose of the code above is to
			// call the MarshalBinary() method, which will return a hash and an error if it fails. If there is an error, the
			// function will return the hash and the error.
			marshal, err := transfer.MarshalBinary()
			if err != nil {
				return hash, err
			}

			// This code is assigning a key-value pair to the p.response object. The key is "result" and the value is an array
			// with two elements: the hex-encoded version of the marshal object and the string representation of the transfer
			// hash. This is likely being used to store or return the result of a calculation or set of operations.
			p.response = map[string]interface{}{
				"result": []string{
					hexutil.Encode(marshal),
					transfer.Hash().String(),
				},
			}
		}

		p.stop = true

	case types.PlatformTron:

		// This if statement checks the length of the tx.Contract value. If it is greater than 0, the code within the statement
		// will be executed. This statement is used to determine if there is any data in the tx.Contract value, and if so,
		// execute a certain set of operations.
		if len(tx.Contract) > 0 {

			// This code defines a struct called 'request' that is used to package the data needed to execute a "transfer"
			// function in a blockchain network. The struct contains the following fields: ContractAddress, FunctionSelector,
			// Parameter, FeeLimit, and OwnerAddress. This data is then used to make the request to the blockchain network.
			request := struct {
				ContractAddress  string `json:"contract_address"`
				FunctionSelector string `json:"function_selector"`
				Parameter        string `json:"parameter"`
				FeeLimit         uint64 `json:"fee_limit"`
				OwnerAddress     string `json:"owner_address"`
			}{
				ContractAddress:  address.New(tx.Contract).Hex(true),
				FunctionSelector: "transfer(address,uint256)",
				Parameter:        strings.TrimPrefix(hexutil.Encode(tx.Data), "0x"),
				FeeLimit:         p.gasUsed(true),
				OwnerAddress:     address.New(owner.String()).Hex(true),
			}

			// This code is used to convert a struct object (request) into a JSON string. The json.Marshal() function is used to
			// take an input object and convert it into a JSON string. The if statement checks for errors when marshalling the
			// data. If an error is encountered, the function returns an empty hash and the error encountered.
			marshal, err := json.Marshal(request)
			if err != nil {
				return hash, err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/triggersmartcontract", p.rpc), "-d", string(marshal)}
		} else {

			// This request creates a data structure in the form of a struct that holds three fields - ToAddress, OwnerAddress and
			// Amount. It assigns the value of the ToAddress field to the value of the tx.To variable, assigns the OwnerAddress
			// field the hexadecimal value of the owner.String() variable and assigns the Amount field the value of the tx.Value variable. This struct is then used to create a JSON object.
			// This data is then used to make the request to the blockchain network.
			request := struct {
				ToAddress    string   `json:"to_address"`
				OwnerAddress string   `json:"owner_address"`
				Amount       *big.Int `json:"amount"`
			}{
				ToAddress:    address.New(tx.To).Hex(true),
				OwnerAddress: address.New(owner.String()).Hex(true),
				Amount:       tx.Value,
			}

			// This code is used to convert a struct object (request) into a JSON string. The json.Marshal() function is used to
			// take an input object and convert it into a JSON string. The if statement checks for errors when marshalling the
			// data. If an error is encountered, the function returns an empty hash and the error encountered.
			marshal, err := json.Marshal(request)
			if err != nil {
				return hash, err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/createtransaction", p.rpc), "-d", string(marshal)}
		}

	default:
		return hash, errors.New("method not found!...")
	}

	// This code is checking if the variable p.stop is false and, if it is, executing the p.commit() method. If p.commit()
	// returns an error, the code returns the hash of the result and the error.
	if !p.stop {
		if err := p.commit(); err != nil {
			return hash, err
		}
	}

	// The purpose of the return p.buildTransaction() statement is to return the value of the buildTransaction() method,
	// which is a part of the p object. This method is likely used to build a transaction and the value it returns may be
	// used for further operations in the program.
	return p.buildTransaction()
}

// GasPrice - The purpose of this function is to get the gas price from the Params struct. It first builds a query from the Params
// struct and then attempts to get a resource using the p.get() function. The code then parses the result from the
// resource map, strips the prefix "0x" from it, and parses it into an int64. Finally, it returns the gas price and any errors encountered.
func (p *Params) gasPrice() (gas int64, err error) {

	p.query = []string{"-X", "POST", "-H", "Content-Type:application/json", "-H", "Accept: application/json", "-d", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}`), p.rpc}

	// This code is checking for an error when attempting to get a resource from the p.get() function. If there is an error,
	// the code returns the error and stops the rest of the code from running.
	resource, err := p.get()
	if err != nil {
		return gas, err
	}

	// This code is used to parse the result from a resource map. It finds the result in the resource map, strips the prefix
	// "0x" from it, and then parses it into an int64. If an error occurs, it returns the error.
	if result, ok := resource["result"]; ok {
		gas, err = strconv.ParseInt(strings.TrimPrefix(result.(string), "0x"), 16, 64)
		if err != nil {
			return gas, err
		}
	}

	return gas, nil
}

// nonce - This function is used to get a nonce (transaction counter) for a given Ethereum address. It sets up an HTTP request to
// query an Ethereum node, saves the response, and checks for errors. If no errors are found, it extracts the nonce from
// the response and returns it as a big integer. If an error is encountered, it is returned instead of the nonce.
func (p *Params) getNonce(address string) (nonce *big.Int, err error) {

	p.query = []string{"-X", "POST", "-H", "Content-Type:application/json", "-H", "Accept: application/json", "-d", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["%v", "latest"],"id":1}`, address), p.rpc}

	// This code is checking for an error when attempting to get a resource from the p.get() function. If there is an error,
	// the code returns the error and stops the rest of the code from running.
	resource, err := p.get()
	if err != nil {
		return nonce, err
	}

	// This code checks if the key "result" is present in the resource map. If it is, it attempts to decode the value stored
	// in the key as a hexadecimal number and returns it as a big integer. If either the key is not present or the decoding
	// fails, an error is returned.
	if result, ok := resource["result"]; ok {
		decodeBig, err := hexutil.DecodeBig(result.(string))
		if err != nil {
			return nonce, err
		}
		return decodeBig, nil
	}

	return nonce, errors.New("nonce not found")
}

// getResource - The purpose of this code is to get the total amount of energy available to a user from the blockchain. It does this by
// casting the public key of a private key to an ECDSA public key, converting the public key to an address, creating a
// request object in the form of a struct, converting the Go data structure into a JSON string, sending the request to a
// server, and then checking for an error when attempting to get a resource from the server. It then checks for the keys
// "freeNetLimit", "freeNetUsed", "NetLimit", and "NetUsed" in the resource map and assigns the associated values to
// variables. The code then returns the total amount of energy available to the user.
func (p *Params) getResource() (energy int64, err error) {

	// This code is attempting to cast the public key of a private key of type "p.private" to an ECDSA public key type. The
	// code will check if the cast is successful with the "ok" variable, and if not, it will return an error.
	public, ok := p.private.Public().(*ecdsa.PublicKey)
	if !ok {
		return energy, errors.New("error casting public key to ECDSA")
	}

	// Is used to convert a public key to an address. It takes the variable public
	// (which is a pointer to a public key) as an argument and returns an address. This can be used to identify a user on
	// the blockchain.
	owner := crypto.PubkeyToAddress(*public)

	// This code creates a request object in the form of a struct with a single field: Address, which is a string type. The
	// value of this field is set to the hexadecimal representation of the owner's string. This is likely part of an API
	// request which will be sent to a server.
	request := struct {
		Address string `json:"address"`
	}{
		Address: address.New(owner.String()).Hex(true),
	}

	// The purpose of this code is to convert a Go data structure into a JSON string. The Marshal function from the json package is used to convert the request into a JSON string. If there is an error during the conversion, the function will return an error.
	marshal, err := json.Marshal(request)
	if err != nil {
		return energy, err
	}

	p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/getaccountresource", p.rpc), "-d", string(marshal)}

	// This code is checking for an error when attempting to get a resource from the p.get() function. If there is an error,
	// the code returns the error and stops the rest of the code from running.
	resource, err := p.get()
	if err != nil {
		return energy, err
	}

	// This is an if statement that is attempting to assign a value (freeNetLimit) to a variable from a map (resource) and
	// then check if the assignment was successful (ok). The ok variable will be set to true if the key "freeNetLimit"
	// exists in the map and the assignment was successful, or false otherwise.
	if freeNetLimit, ok := resource["freeNetLimit"]; ok {

		// This code is checking if the key "freeNetUsed" is present in the resource map. If it is not present, then the value
		// of freeNetUsed is set to 0. This is useful for ensuring that the program does not panic if the key is not present.
		freeNetUsed, ok := resource["freeNetUsed"]
		if !ok {
			freeNetUsed = float64(0)
		}

		// The purpose of this code is to check if the map "resource" contains the key "NetLimit" and if it does, assign the
		// associated value to the variable netLimit. If the key is not found, the variable netLimit is assigned the value 0.
		netLimit, ok := resource["NetLimit"]
		if !ok {
			netLimit = float64(0)
		}

		// This code is checking to see if the key "NetUsed" exists in the resource map. If it does not exist, the code sets
		// the netUsed variable to 0. This code can be used to set a default value to the netUsed variable if the key does not exist.
		netUsed, ok := resource["NetUsed"]
		if !ok {
			netUsed = float64(0)
		}

		return int64(freeNetLimit.(float64) + netLimit.(float64) - freeNetUsed.(float64) - netUsed.(float64)), nil
	}

	return energy, errors.New("account resource not found")
}

// EstimateGas - This function is used to estimate the amount of gas needed to complete a Transfer transaction. It takes a pointer to a
// Params type and a pointer to a Transfer type as arguments, and returns an int64 representing the estimated amount of
// gas and an error if there is one. It is used to determine the amount of gas that will be required for a transaction
// before it is sent, so that the user can adjust the gas fee accordingly.
func (p *Params) EstimateGas(tx *Transfer) (fee int64, err error) {

	var (
		gas int
	)

	// This code is attempting to cast the public key of a private key of type "p.private" to an ECDSA public key type. The
	// code will check if the cast is successful with the "ok" variable, and if not, it will return an error.
	public, ok := p.private.Public().(*ecdsa.PublicKey)
	if !ok {
		return fee, errors.New("error casting public key to ECDSA")
	}

	// Is used to convert a public key to an address. It takes the variable public
	// (which is a pointer to a public key) as an argument and returns an address. This can be used to identify a user on
	// the blockchain.
	owner := crypto.PubkeyToAddress(*public)

	// This switch statement is used to evaluate the value of the p.platform variable and take the appropriate action based
	// on the value of the variable. This is a common way to control the flow of a program based on the value of a variable.
	switch p.platform {
	case types.PlatformEthereum:

		var (
			marshal []byte
		)

		// This if statement checks the length of the tx.Contract value. If it is greater than 0, the code within the statement
		// will be executed. This statement is used to determine if there is any data in the tx.Contract value, and if so,
		// execute a certain set of operations.
		if len(tx.Contract) > 0 {

			// This request is creating a JSON object with fields for "to", "from", and "data". The "to" field is set to the value
			// of the "tx.Contract" variable, the "from" field is set to the value of the "owner.String()" variable, and the
			// "data" field is set to the value of the "hexutil.Encode(tx.Data)" variable. This request is likely being used to send data to a web service.
			request := struct {
				To   string `json:"to"`
				From string `json:"from"`
				Data string `json:"data"`
			}{
				From: owner.String(),
				To:   tx.Contract,
				Data: hexutil.Encode(tx.Data),
			}

			// This code is used to convert a struct object (request) into a JSON string. The json.Marshal() function is used to
			// take an input object and convert it into a JSON string. The if statement checks for errors when marshalling the
			// data. If an error is encountered, the function returns an empty hash and the error encountered.
			marshal, err = json.Marshal(request)
			if err != nil {
				return fee, err
			}

		} else {

			// This code snippet is defining a structure called "request" which is used to send a transaction from one address to
			// another. The structure contains the sender address, the receiver address, and the amount of the transaction. The
			// json tags allow the structure to be encoded into JSON format for communication over the network.
			request := struct {
				To    string `json:"to"`
				From  string `json:"from"`
				Value string `json:"value"`
			}{
				From:  owner.String(),
				To:    tx.To,
				Value: hexutil.EncodeBig(tx.Value),
			}

			// This code is used to convert a struct object (request) into a JSON string. The json.Marshal() function is used to
			// take an input object and convert it into a JSON string. The if statement checks for errors when marshalling the
			// data. If an error is encountered, the function returns an empty hash and the error encountered.
			marshal, err = json.Marshal(request)
			if err != nil {
				return fee, err
			}
		}

		p.query = []string{"-X", "POST", "-H", "Content-Type:application/json", "-H", "Accept: application/json", "-d", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_estimateGas","params":[%v],"id":1}`, string(marshal)), p.rpc}

	case types.PlatformTron:

		// This if statement checks the length of the tx.Contract value. If it is greater than 0, the code within the statement
		// will be executed. This statement is used to determine if there is any data in the tx.Contract value, and if so,
		// execute a certain set of operations.
		if len(tx.Contract) > 0 {

			// This request defines the data that will be sent to a smart contract. It contains the address of the smart contract,
			// the function selector for the function that will be called, the parameters used to call the function, the fee limit
			// for the call, and the address of the account making the call.
			request := struct {
				ContractAddress  string `json:"contract_address"`
				FunctionSelector string `json:"function_selector"`
				Parameter        string `json:"parameter"`
				FeeLimit         uint64 `json:"fee_limit"`
				OwnerAddress     string `json:"owner_address"`
			}{
				ContractAddress:  address.New(tx.Contract).Hex(true),
				FunctionSelector: "transfer(address,uint256)",
				Parameter:        strings.TrimPrefix(hexutil.Encode(tx.Data), "0x"),
				FeeLimit:         p.gasUsed(true),
				OwnerAddress:     address.New(owner.String()).Hex(true),
			}

			// This code is used to convert a struct object (request) into a JSON string. The json.Marshal() function is used to
			// take an input object and convert it into a JSON string. The if statement checks for errors when marshalling the
			// data. If an error is encountered, the function returns an empty hash and the error encountered.
			marshal, err := json.Marshal(request)
			if err != nil {
				return fee, err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/triggerconstantcontract", p.rpc), "-d", string(marshal)}

		} else {

			// This code creates a struct (request) that holds three variables: ToAddress, OwnerAddress, and Amount. The values of
			// the three variables are set to tx.To, address.New(owner.String()).Hex(true), and tx.Value respectively. This struct
			// is used to store data in a structured way so that it can be retrieved, manipulated, and used in other parts of the program.
			request := struct {
				ToAddress    string   `json:"to_address"`
				OwnerAddress string   `json:"owner_address"`
				Amount       *big.Int `json:"amount"`
			}{
				ToAddress:    address.New(tx.To).Hex(true),
				OwnerAddress: address.New(owner.String()).Hex(true),
				Amount:       tx.Value,
			}

			// This code is used to convert a struct object (request) into a JSON string. The json.Marshal() function is used to
			// take an input object and convert it into a JSON string. The if statement checks for errors when marshalling the
			// data. If an error is encountered, the function returns an empty hash and the error encountered.
			marshal, err := json.Marshal(request)
			if err != nil {
				return fee, err
			}

			p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/createtransaction", p.rpc), "-d", string(marshal)}
		}

	default:
		return fee, errors.New("method not found!...")
	}

	// This code is checking for an error when attempting to get a resource from the p.get() function. If there is an error,
	// the code returns the error and stops the rest of the code from running.
	constant, err := p.get()
	if err != nil {
		return fee, err
	}

	// This code is used to check if there is an error in the response of a given operation. If there is an error present in
	// the response, it will return a new error with the message associated with the error.
	if err, ok := constant["error"]; ok {
		return fee, errors.New(err.(map[string]interface{})["message"].(string))
	}

	// This if statement is checking to see if the key "Error" exists in the constant map. If it does, then it is returning
	// a fee and creating a new error using the value of the key. This is likely being used to handle an error condition.
	if err, ok := constant["Error"]; ok {
		return fee, errors.New(err.(string))
	}

	// This code is using the comma ok idiom to check if the key "result" is present in the constant map. If the key is
	// present, it assigns the value of the key to the variable result, and sets ok to true. Otherwise, ok is set to false.
	if result, ok := constant["result"]; ok {

		// The switch result.(type) statement is used to check the type of the value stored in the result variable. The switch
		// statement will match the type of the result variable to the specified type and execute the corresponding code block
		// if the types match. This can be used to handle different types of data or to take different actions depending on the type of the variable.
		switch result.(type) {
		case string:

			// This code is decoding a hexadecimal string and converting it into a big integer. The decodeBig variable is used to
			// store the result of the conversion. If there is an error in the conversion, the error variable will be set to a
			// non-nil value and the function will return the fee and the error.
			decodeBig, err := hexutil.DecodeBig(result.(string))
			if err != nil {
				return fee, err
			}

			// This code is used to get the gas price required to complete a transaction on the blockchain. It uses the gasPrice()
			// function from the p package to get the gas price, and if an error is thrown, it returns it and stops the execution of the code.
			gasPrice, err := p.gasPrice()
			if err != nil {
				return fee, err
			}

			return gasPrice * int64(decodeBig.Uint64()), nil
		}
	}

	// This if statement is checking to see if the constant map has a key called "raw_data_hex", and if it does, it sets the
	// variable raw to the value associated with that key. The ok variable is a boolean that is true if the key exists in
	// the map and false if it does not.
	if raw, ok := constant["raw_data_hex"]; ok {

		// This code is checking if the variable "err" is nil (has no value) after the method call "p.getResource()". If it is
		// not nil (it has a value), then the function will return the variable "fee" and the error.
		price, err := p.getResource()
		if err != nil {
			return fee, err
		}

		// This code is used to calculate a fee. The first line calculates the fee as 10 times the length of the string in the
		// raw variable. The second line checks if the price is greater than or equal to the fee divided by 10. If it is, then the fee is set to 0.
		fee = decimal.New(int64(len(raw.(string)))).Mul(10).Int64()

		// This code checks to see if the variable 'energies' is true. If it is, then it checks to see if the variable 'price'
		// is greater than or equal to 10 times the variable 'fee'. If it is, then the variable 'fee' is set to 0.
		if price >= decimal.New(fee).Div(10).Int64() {
			fee = 0
		}

		return fee, nil
	}

	// This if statement is used to check if the key "transaction" is present in the map constant. If it is present, the
	// value associated with it is assigned to the variable transaction and the boolean value ok is set to true. Otherwise, ok is set to false.
	if transaction, ok := constant["transaction"]; ok {

		// This is an example of type assertion, which is a form of conversion. It is used to check if the type of variable
		// is the same as the type specified in the assertion. In this case, it is checking if the variable "transaction" is of
		// type map[string]interface{}. If it is, the "ok" variable will be set to true, otherwise it will be set to false.
		if transaction, ok := transaction.(map[string]interface{}); ok {

			// This code checks if the map "transaction" contains a key "txID". The "_" is used in a Go language statement to
			// ignore the value associated with this key if it exists. The "ok" variable will be true if the key exists and false
			// if it does not.
			if _, ok := transaction["txID"]; ok {

				// This code is checking if the variable "err" is nil (has no value) after the method call "p.getResource()". If it is
				// not nil (it has a value), then the function will return the variable "fee" and the error.
				price, err := p.getResource()
				if err != nil {
					return fee, err
				}

				// This code is signing a transaction with the private key of the user. The purpose of signing a transaction is to
				// create a digital signature that is used to authenticate the sender of the transaction and to ensure that the
				// transaction has not been modified. The signature is then sent along with the transaction to be verified by the
				// network. If an error occurs during the signing process, the code returns an error.
				signature, err := crypto.Sign(common.Hex2Bytes(transaction["txID"].(string)), p.private)
				if err != nil {
					return fee, err
				}

				// This if statement is checking to see if the constant map has a key called "raw_data_hex", and if it does, it sets the
				// variable raw to the value associated with that key. The ok variable is a boolean that is true if the key exists in
				// the map and false if it does not.
				if _, ok := transaction["raw_data_hex"]; ok {

					// This code is used to decode the data stored in a hexadecimal format. The line "raw, err :=
					// hex.DecodeString(transaction["raw_data_hex"].(string))" takes the data stored in the variable "transaction" and
					// tries to decode it from hexadecimal into a byte array. The "if err != nil" line checks if an error occurred
					// during the decoding process, and if so, it returns an error and the fee.
					raw, err := hex.DecodeString(transaction["raw_data_hex"].(string))
					if err != nil {
						return fee, err
					}

					// This statement is adding the length of the raw variable to the gas variable. The purpose of this statement is to
					// increase the value of the gas variable by the length of the raw variable.
					gas += len(raw)
				}
				gas += len(signature)

				// The code checks if the key "energy_used" is present in the constant map. If it is, the value corresponding to that
				// key is stored in the variable energy. The ok variable is a boolean which is set to true if the key is present in
				// the constant map, and false if it is not.
				if energy, ok := constant["energy_used"]; ok {

					// The purpose of this code is to calculate a fee from a given energy and gas value. It takes the energy and gas
					// values as float64, adds 9, 60, and the product of the energy and 10, and multiplies the sum by 10 to calculate
					// the fee. The result is an int64.
					fee = decimal.New(9 + 60 + int64(energy.(float64)*10) + int64(gas)).Mul(10).Int64()

					// This code is used to calculate the bandwidth of a network. The variable 'fee' represents the fee associated with the network, while the variable 'energy' represents the energy consumed by the network.
					// The code calculates the bandwidth by subtracting the energy consumed (multiplied by 100) from the fee. The result is then stored as an integer in the variable 'bandwidth'.
					bandwidth := decimal.New(fee).Sub(energy.(float64) * 100).Int64()

					// This code is intended to calculate a fee based on a price and a bandwidth.  If the price is greater than or equal
					// to 1/10 of the bandwidth, then the fee is reduced by the bandwidth value.
					if price >= decimal.New(bandwidth).Div(10).Int64() {
						fee = decimal.New(fee).Sub(float64(bandwidth)).Int64()
					}

					return fee, nil
				}
			}
		}
	}

	return fee, errors.New("constant fee calculate not found!...")
}

// buildTransaction - The purpose of this code is to build a transaction for a given platform, such as Ethereum or Tron. It checks for the
// presence of a "result" or "Error" key in the "response" map of the object "p". Depending on the platform, it assigns
// the transaction ID and a signature to the "transaction" and "signature" fields of the "p.response" map. It also sets
// the "success" property of the "p" object to true to indicate that the process was successful. Finally, it returns the
// transaction ID and either a nil or an error object.
func (p *Params) buildTransaction() (txID string, err error) {

	// The switch statement is a control flow statement that evaluates an expression, matches the expression's value to a
	// case label, and executes statements associated with that case label. In this example, the switch statement is
	// evaluating the value of the expression p.platform. Depending on the value of p.platform, different statements will be executed.
	switch p.platform {
	case types.PlatformEthereum:

		// This code checks whether the "result" key exists in the map p.response. If it does, it assigns the value of the key
		// to the variable result and assigns true to the variable ok. The if statement then evaluates whether ok is true. If
		// ok is true, any code within the if statement will be executed.
		if result, ok := p.response["result"]; ok {

			// This code is used to assign the first element in the result slice to the "transaction" field of the p.response map.
			// The purpose of this code is to store the result of a transaction in the response field of the p object.
			p.response["transaction"] = result.([]string)[0]

			// txID is a variable used to store the second element in the result array of strings. The purpose of this variable is
			// to access and store the transaction ID associated with the result.
			txID = result.([]string)[1]

			// This statement sets the value of the "success" property of the "p" object to true. The purpose of this statement
			// may be to indicate that an operation or process has been successful.
			p.success = true

			return txID, nil
		}

	case types.PlatformTron:

		// The purpose of this code is to check if there is an "Error" key in the "response" map of the object "p". If so, it
		// will return the transaction ID (txID) and an error object that contains the description of the error.
		if err, ok := p.response["Error"]; ok {
			return txID, errors.New(err.(string))
		}

		// The statement is used to check if the key "result" is present in the map p.response. The statement assigns the value
		// associated with the key to the variable _ and the boolean value ok to the variable ok. If the key is present, ok
		// will be true, otherwise it will be false.
		if _, ok := p.response["result"]; ok {

			// This code is used to check if the map "p.response" contains a key named "transaction", and if it does, it sets the
			// map "p.response" equal to the value associated with the "transaction" key. The expression
			// "transaction.(map[string]interface{})" is used to cast the value of "transaction" into the correct data type, a
			// map[string]interface{}.
			if transaction, ok := p.response["transaction"]; ok {
				p.response = transaction.(map[string]interface{})
			}

			//	This if statement is checking to see if the value of the txID key is present in the response map. The "_" is a
			//	blank identifier which is used when there is no need to use the value of the variable. If the txID key is not
			//	present, the statement returns an error.
		} else if _, ok = p.response["txID"]; !ok {
			return txID, errors.New("map[string]interface{} not recognized")
		}

		// This statement is checking if the key "txID" exists in the map p.response. The if statement is checking for the
		// existence of the key "txID" in the map p.response. If the key exists, the statement returns true, otherwise it returns false.
		if _, ok := p.response["txID"]; ok {

			// The purpose of this code is to sign a transaction using a private key. The transaction ID is obtained from a
			// response object and is converted to bytes using the common.Hex2Bytes() function. The crypto.Sign() function is then
			// used to sign the transaction using the private key. If an error occurs during the signing process, an error is returned.
			signature, err := crypto.Sign(common.Hex2Bytes(p.response["txID"].(string)), p.private)
			if err != nil {
				return txID, err
			}

			// The purpose of the code is to assign a signature in the form of a hexadecimal string to the "signature" property of
			// the response object. The signature is obtained by converting the bytes of the signature to hexadecimal form using the common.Bytes2Hex() function.
			p.response["signature"] = []string{common.Bytes2Hex(signature)}

			// The purpose of this statement is to set the value of the variable 'p.success' to 'true'. This can be used to indicate the success of an action or check if something has been successful.
			p.success = true

			return p.response["txID"].(string), nil
		}
	}

	return txID, errors.New("invalid transactions map")
}

// Transaction - This function is used to perform a transaction on either the Ethereum or Tron blockchain, depending on the platform
// selected. It checks for a successful initialization of the transfer function and formats the query accordingly. If the
// transaction fails, it returns an error with a code and message.
func (p *Params) Transaction() error {

	// The purpose of this code is to check if the boolean value of the variable p.success is false. If it is, it will
	// return an error message stating that the "transfer function has not been initialized". This is used as a way to alert
	// the user that the transfer function has not been initialized and to prompt the user to take appropriate action.
	if !p.success {
		return errors.New("transfer function has not been initialized")
	}

	// The switch statement is a control flow statement that evaluates an expression, matches the expression's value to a
	// case label, and executes statements associated with that case label. In this example, the switch statement is
	// evaluating the value of the expression p.platform. Depending on the value of p.platform, different statements will be executed.
	switch p.platform {
	case types.PlatformEthereum:

		p.query = []string{"-X", "POST", "-H", "Content-Type:application/json", "-H", "Accept: application/json", "-d", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["%v"],"id":1}`, p.response["transaction"]), p.rpc}

	case types.PlatformTron:

		// The code is checking if the key "transaction" exists in the map "p.response", and if it does, assign the value of
		// that key to the variable "transaction" and set the variable "ok" to true.
		if transaction, ok := p.response["transaction"]; ok {

			// The purpose of this code is to assign the response from a transaction to a variable called "p.response" as a
			// map[string]interface{}. The map[string]interface{} is a type of data structure in which data can be stored as
			// key-value pairs of string-interface{}, respectively.
			p.response = transaction.(map[string]interface{})

			//	This code is used to check if the key "txID" exists in the map p.response. If it does not exist, it will return an
			//	error letting the user know that the map[string]interface{} was not recognized.
		} else if _, ok = p.response["txID"]; !ok {
			return errors.New("map[string]interface{} not recognized")
		}

		// This code is used to convert a struct object (request) into a JSON string. The json.Marshal() function is used to
		// take an input object and convert it into a JSON string. The if statement checks for errors when marshalling the
		// data. If an error is encountered, the function returns an empty hash and the error encountered.
		serialize, err := json.Marshal(p.response)
		if err != nil {
			return err
		}

		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/broadcasttransaction", p.rpc), "-d", string(serialize)}
	}

	// The purpose of this code is to call the commit() function on the object 'p', and if it returns an error (err != nil)
	// the function returns that error to the caller.
	if err := p.commit(); err != nil {
		return err
	}

	// The purpose of this code is to set the value of the 'success' property of the 'p' object to false.
	p.success = false

	// This is a conditional statement that is used to check if the response map contains a key named "code". If the key
	// exists, the value associated with it is assigned to the variable "code" and the boolean value "ok" is set to true. If
	// the key does not exist in the map, the boolean "ok" will be false. This can be used to ensure that the response
	// contains the necessary key before the value is used.
	if code, ok := p.response["code"]; ok {

		// This code is decoding a hexadecimal string stored in the response field of the p object. This is done so that the
		// code can use the data stored in the response field in its original form, which is a byte array. The DecodeString()
		// function will convert the hexadecimal string into a byte array so that the code can use it.
		message, err := hex.DecodeString(p.response["message"].(string))
		if err != nil {
			return err
		}

		return errors.New(fmt.Sprintf("[%v], %v", code.(string), string(message)))
	}

	// This code is used to check if there is an error in the response of a given operation. If there is an error present in
	// the response, it will return a new error with the message associated with the error.
	if err, ok := p.response["error"]; ok {
		return errors.New(err.(map[string]interface{})["message"].(string))
	}

	return nil
}

// Data - This function is used to construct a data payload for a transaction. It takes in a "to" address and an "amount" in
// bytes and returns a data payload encoded in bytes. It also checks the platform to determine which protocol is being
// used and then formats the data payload accordingly.
func (p *Params) Data(to string, amount []byte) (data []byte, err error) {

	// The switch statement is a control flow statement that evaluates an expression, matches the expression's value to a
	// case label, and executes statements associated with that case label. In this example, the switch statement is
	// evaluating the value of the expression p.platform. Depending on the value of p.platform, different statements will be executed.
	switch p.platform {
	case types.PlatformEthereum:

		// This code is decoding a hexadecimal string. To decode variable is set to the result of the hex.DecodeString()
		// function, which decodes the string passed to it into its byte representation. The strings.TrimPrefix() function is
		// used to remove the "0x" prefix from the string before it is passed to the hex.DecodeString() function. Finally, an
		// error check is performed to make sure that no errors occurred during the decoding process. If an error occurred, the
		// data and the error are returned.
		decode, err := hex.DecodeString(strings.TrimPrefix(to, "0x"))
		if err != nil {
			return data, err
		}

		//This code is used to construct a call data for a function call in Ethereum. The call data is a byte array that is
		//used to invoke a contract method on the Ethereum blockchain.
		// The first line appends the method signature calculated using Keccak-256 hash to the data array.
		// The second line appends the decoded address (address argument of the method) that is left-padded with zeroes to the data array.
		// The third line appends the amount (uint256 argument of the method) that is left-padded with zeroes to the data array.
		// The purpose of the code is to construct the call data for the given method signature and arguments.
		data = append(data, help.SignatureKeccak256([]byte("transfer(address,uint256)"))...)
		data = append(data, common.LeftPadBytes(decode, 32)...)
		data = append(data, common.LeftPadBytes(amount, 32)...)

	case types.PlatformTron:

		// This code is used to decode an address string into a byte slice. The address.New(to, true).Hex(true)[2:] is used to
		// get the hex-encoded address from the two variable. The hex.DecodeString() is then used to decode the hex-encoded
		// address into a byte slice and store it in the decode variable. If an error occurs, the function returns the data and the error.
		decode, err := hex.DecodeString(address.New(to, true).Hex(true)[2:])
		if err != nil {
			return data, err
		}

		// This code is appending two byte slices, decode and amount, to a larger byte slice data. The function LeftPadBytes is
		// used to ensure that both byte slices have a length of 32 bytes, which is likely to be necessary for some type of
		// computation.
		data = append(data, common.LeftPadBytes(decode, 32)...)
		data = append(data, common.LeftPadBytes(amount, 32)...)
	}

	return data, nil
}

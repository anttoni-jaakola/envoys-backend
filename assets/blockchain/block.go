package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/pkg/errors"
	"strings"
)

// BlockByNumber - This function is a method for the Params object in the pbspot package. It is used to retrieve a Block object by its
// number using either an Ethereum or Tron API. It will set the query parameters accordingly based on the platform, and
// then call the commit() method on the Params object. If an error is returned, it will return the "block" and the error.
func (p *Params) BlockByNumber(number int64) (block *Block, err error) {

	// This line of code is creating a new instance of a class called Block. This is known as instantiation, which is the
	// process of creating an instance of a class. This particular line of code is creating a new block object and assigning
	// it to the variable 'block'.
	block = new(Block)

	// The purpose of the switch statement is to determine which type of platform is being used and then set the query
	// accordingly. In this example, if the platform is Ethereum, the query will include parameters for an
	// eth_getBlockByNumber request, whereas if the platform is Tron, the query will include parameters for a getblockbynum
	// request. The default case will return an error if the platform is not supported.
	switch p.platform {
	case types.PlatformEthereum:
		p.query = []string{"-X", "POST", "-H", "Content-Type:application/json", "-H", "Accept: application/json", "-d", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":1}`, number), p.rpc}
	case types.PlatformTron:
		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/getblockbynum", p.rpc), "-d", fmt.Sprintf(`{"num": %d}`, number)}
	default:
		return block, errors.New("method not found!...")
	}

	// This code is checking for an error when calling the commit() method on the 'p' object. If an error is returned, it
	// will return the "block" and the error. This code is used to handle potential errors when calling the commit() method.
	if err := p.commit(); err != nil {
		return block, err
	}

	return p.block()
}

// block - This function is used to process and parse the data received from a blockchain platform, such as Ethereum or Tron. It
// uses a switch statement to determine which platform the data is from and then uses type assertions, type switches, and
// looping to process the data accordingly. It then serializes the data into a byte array and converts it back into a
// data structure. Finally, it checks the type of the data and appends it to a list of block transactions if it is either
// a TypeInternal or TypeContract transaction.
func (p *Params) block() (block *Block, err error) {

	// This line of code is creating a new instance of a class called Block. This is known as instantiation, which is the
	// process of creating an instance of a class. This particular line of code is creating a new block object and assigning
	// it to the variable 'block'.
	block = new(Block)

	// The switch statement is used here to determine what platform the p object is using. If the p object is using the
	// Ethereum platform, then the code within the case statement will be executed.
	switch p.platform {
	case types.PlatformEthereum:

		// The if statement is a type assertion that checks if the value stored in the response["result"] key of the p map is
		// an interface containing a map[string]interface{}. If the assertion succeeds, the variable result will be set to the
		// value stored in the response["result"] key, and the variable ok will be set to true. Otherwise, ok will be set to false.
		if result, ok := p.response["result"].(map[string]interface{}); ok {

			// This code is testing the type of the value stored in the map at the key "transactions". If the value is an array of
			// interfaces, the code will assign it to the variable `transactions`. Otherwise, the ok variable will be false. This
			// allows the code to safely access the value without causing a runtime panic.
			if transactions, ok := result["transactions"].([]interface{}); ok {

				// This for loop is iterating through a list of transactions and checking for an input key in each transaction. If an input key is found, it will do something.
				for i := 0; i < len(transactions); i++ {
					if input, ok := transactions[i].(map[string]interface{})["input"]; ok {

						// This code is used to determine the type of transaction based on the input string. If the input string contains
						// "0xa9059cbb", the code sets the "type" of the transaction to "TypeContract", otherwise it sets the "type" to
						// "TypeInternal".
						if strings.Contains(input.(string), "0xa9059cbb") {
							transactions[i].(map[string]interface{})["type"] = TypeContract
						} else {
							transactions[i].(map[string]interface{})["type"] = TypeInternal
						}

					}
				}

			}

			// This code is used to convert a variable (result) to a JSON format. The serialize variable is assigned to the
			// json.Marshal() function, which takes in the result variable and returns the JSON encoded data. If there is an
			// error, it is returned to the calling function.
			serialize, err := json.Marshal(result)
			if err != nil {
				return block, err
			}

			// This code is attempting to deserialize a JSON object into a block object. The if statement is used to check for
			// errors when attempting to deserialize the JSON object into the block object. If there is an error, the block and
			// the error are returned.
			if err := json.Unmarshal(serialize, &block); err != nil {
				return block, err
			}

		} else {
			return block, errors.New("block not found!...")
		}

	case types.PlatformTron:

		// The purpose of this code is to check for an error in the response map "p.response". If the error is present, the
		// code will return a block and an error with the err.(string) value.
		if err, ok := p.response["Error"]; ok || err != nil {
			return block, errors.New(err.(string))
		}

		// This code is used to check if the key "blockID" exists in the map "p.response" and, if it does, set the "Hash" field
		// of the "block" struct to the value associated with the key. The "ok" variable is used to check if the key was found;
		// if it is not found, the code block will not be executed.
		if hash, ok := p.response["blockID"]; ok {
			block.Hash = hash.(string)
		}

		// The purpose of the above code is to check if the key "block_header" exists in the map p.response. If it exists, the
		// ok variable will be set to true, if not, it will be set to false.
		if _, ok := p.response["block_header"]; ok {

			// This code is checking for the existence of the "block_header" key in the p.response map and then, if it exists, it
			// is checking for the existence of the "raw_data" key in the "block_header" value. If both keys exist, it is
			// assigning the values of the "txTrieRoot" and "parentHash" keys in the "raw_data" value to the variables
			// "TransactionsRoot" and "ParentHash" in the "block" object.
			if raws, ok := p.response["block_header"].(map[string]interface{})["raw_data"].(map[string]interface{}); ok {
				block.TransactionsRoot = raws["txTrieRoot"].(string)
				block.ParentHash = raws["parentHash"].(string)
			}

		}

		// This code checks to see if the key "transactions" exists and is of type []interface{} in the map p.response. If both
		// conditions are true, the result is stored in the variable transactions and ok is set to true.
		if transactions, ok := p.response["transactions"].([]interface{}); ok {

			for i := 0; i < len(transactions); i++ {

				var (
					column Transaction
				)

				// This code is checking to see if the key "contract" exists within the map "raw_data" which is nested within the map
				// "transaction".  The value associated with the "contract" key is expected to be a slice of interface{} and if it
				// is, the variable "ok" will be set to true.  This allows the code to proceed to take further action with the value
				// associated with the "contract" key if it exists.
				if transaction, ok := transactions[i].(map[string]interface{})["raw_data"].(map[string]interface{})["contract"].([]interface{}); ok {

					for i := 0; i < len(transaction); i++ {

						// This statement is used to check if the value of the "value" key within the "parameter" key within the
						// "transaction" key is a map. If it is a map, the value and ok variables are created and assigned values. The
						// value variable is assigned the value of the "value" key and the ok variable is assigned the boolean "true".
						if value, ok := transaction[i].(map[string]interface{})["parameter"].(map[string]interface{})["value"].(map[string]interface{}); ok {

							// This is a conditional statement that checks if the value "amount" is present in the map "value". If it is
							// present, the statement assigns the float value of "amount" to the variable "column.Value". It does this using
							// the fmt.Sprintf function to convert the float value to a string.
							if amount, ok := value["amount"]; ok {
								column.Value = fmt.Sprintf("%v", amount.(float64))
							}

							// This code is checking whether the "owner_address" key exists in the "value" map. If the key exists, it then
							// sets the From field of the "column" struct to the value associated with the "owner_address" key. The ok
							// variable is a boolean that is set to true if the key exists, and false if it doesn't.
							if from, ok := value["owner_address"]; ok {
								column.From = from.(string)
							}

							// This code is checking for the presence of either a "to_address" or a "contract_address" key in the "value" map.
							// If either of the keys is present, the code is assigning the corresponding value to the "To" field of the
							// "column" struct. The "ok" variable is used to indicate whether the lookup of the key was successful. The '.()'
							// syntax is used to convert the value of the map to the type needed by the destination field.
							if to, ok := value["to_address"]; ok {
								column.To = to.(string)
							} else if to, ok = value["contract_address"]; ok {
								column.To = to.(string)
							}

							// This code is checking for the presence of the key "data" in the value map, and if it is present, it is decoding
							// it from a hex string into a byte array. This is then assigned to the "Data" field of the "column" object.
							if data, ok := value["data"]; ok {

								data, err := hex.DecodeString(data.(string))
								if err != nil {
									return block, err
								}

								column.Data = data
							}
						}

						// This code is checking if the key "type" exists in the map transaction[i], and if it does, it is assigning the
						// value associated with it to the variable types. The ok variable is a boolean that is set to true if the key is
						// found, and false otherwise.
						if types, ok := transaction[i].(map[string]interface{})["type"]; ok {

							// This switch statement is used to set the Type of a column based on the string value of the given type. If the
							// type provided is "TransferContract", the column type is set to "TypeInternal". If the type provided is
							// "TriggerSmartContract", the column type is set to "TypeContract".
							switch types.(string) {
							case "TransferContract":
								column.Type = TypeInternal
							case "TriggerSmartContract":
								column.Type = TypeContract
							}

						}
					}
				}

				// This code is checking to see if the transactions at index i is a map of strings to interfaces. If it is a map, it
				// sets the column's Hash value to the txID field of the map.
				if hash, ok := transactions[i].(map[string]interface{}); ok {
					column.Hash = hash["txID"].(string)
				}

				// This code checks the type of column and if the type is TypeInternal or TypeContract then it appends the column
				// to a list of block transactions.
				if column.Type == TypeInternal || column.Type == TypeContract {
					block.Transactions = append(block.Transactions, &column)
				}
			}
		}
	}

	return block, nil
}

// Status - The purpose of this code is to check the status of a transaction on either the Ethereum or Tron blockchain. It
// switches between different platforms for querying transaction receipts, sets the query to the appropriate RPC call,
// creates a JSON request with the transaction id, converts the request object into a JSON string, checks for an error
// when committing a transaction, and checks the result of the transaction to return a success or failure result
// depending on the status.
func (p *Params) Status(tx string) (success bool) {

	// The purpose of this code is to switch between different platforms for querying transaction receipts. For Ethereum, it
	// sets the query to the appropriate RPC call, and for Tron, it creates a JSON request with the transaction id and sets
	// the query to the appropriate RPC call.
	switch p.platform {
	case types.PlatformEthereum:
		p.query = []string{"-X", "POST", "-H", "Content-Type:application/json", "-H", "Accept:application/json", "-d", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%s"],"id":1}`, tx), p.rpc}
	case types.PlatformTron:

		// This is an example of a struct that is used to define a request object in the JSON format. The struct contains one
		// field, "Value" and has the type string, with the tag "json:"value". The Value field is set to the value of the
		// variable tx. The purpose of this struct is to provide a data structure that can be serialized into JSON format and
		// sent in an HTTP request.
		request := struct {
			Value string `json:"value"`
		}{
			Value: tx,
		}

		// The purpose of this code is to convert the request object into a JSON string. The Marshal function from the json
		// package is used to do this conversion. If it is successful, the success variable is returned. If an error occurs,
		// then the code returns without doing anything.
		marshal, err := json.Marshal(request)
		if err != nil {
			return success
		}

		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/gettransactioninfobyid", p.rpc), "-d", string(marshal)}
	}

	// The purpose of the above code is to check for an error when committing a transaction. If an error is found, it is
	// returned and the success value is returned.
	if err := p.commit(); err != nil {
		return success
	}

	// This is a form of conditional statement known as a "comma, ok" idiom. It is used to check if the key "result" exists
	// in a map called p.response. If it does, the value of the key will be stored in the variable result, and the ok
	// variable will be set to true. This allows the code to take different actions depending on if the key exists or not.
	if result, ok := p.response["result"]; ok {

		// This code is used to check the status of a transaction on either the Ethereum or Tron blockchain. It checks the
		// result of the transaction and returns a success or failure result depending on the status. The code checks the
		// status of the transaction by checking the "status" of the maps variable if it is an Ethereum transaction, or
		// checking the result as a string if it is a Tron transaction.
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

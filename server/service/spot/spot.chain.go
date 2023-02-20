package spot

import (
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/blockchain"
	"github.com/cryptogateway/backend-envoys/assets/common/address"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/assets/common/query"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strconv"
	"strings"
)

// depositEthereum - This code is part of a Service object in the code which handles Ethereum deposits. The purpose of this code is to
// update the Ethereum block height and scan the new block for any transactions that involve deposits. It will then
// publish the deposit transaction to an exchange. It also includes error handling, such as attempting to debug any
// errors that occur throughout the process.
func (e *Service) depositEthereum(chain *pbspot.Chain) {

	// The purpose of this code is to use to defer keyword to recover from a panic. It does this by catching the panic with
	// the recover() function, and then using the e.Context.Debug() function to log the recovered panic. If the panic is
	// successfully recovered, the code will return.
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	// This code is used to establish a connection between a client and a blockchain platform. The first line is creating a
	// new client connection to the blockchain platform, and the second line is checking for any errors that may have
	// occurred during the connection. If an error is found, the code will exit and not continue.
	client, err := blockchain.Dial(chain.GetRpc(), chain.GetPlatform())
	if e.Context.Debug(err) {
		return
	}

	// This code is checking if an error occurred when the client tried to retrieve a block by number from the blockchain.
	// If an error did occur, the code returns without doing anything else.
	blockBy, err := client.BlockByNumber(chain.GetBlock())
	if err != nil { // No debug....
		return
	}

	// This code is looping through the transactions of a block, where blockBy is the block that the transactions belong to.
	// The underscore is a special character that is used when you don't care about the index of the loop. It is commonly
	// used when you only need the value of the array.
	for _, tx := range blockBy.Transactions {

		var (
			item pbspot.Transaction
		)

		// The switch statement is used to execute one statement from multiple conditions. In this example, the value of the
		// variable tx.Type is compared against a number of possible values and the corresponding statement is executed. This
		// can be useful when there are multiple possible outcomes and each one requires a different action.
		switch tx.Type {
		case blockchain.TypeInternal:

			// This code is setting the quantity to the value of the tx.Value field, which is a hexadecimal string prefixed with
			// 0x. The SetString method of the big.Int type is used to convert the hexadecimal string to a big.Int type, which is
			// then assigned to the quantity variable.
			quantity := new(big.Int)
			quantity.SetString(strings.TrimPrefix(tx.Value, "0x"), 16)

			// This code is checking if the value of the quantity is greater than zero. If the value of quantity is greater than
			// zero, the code assigns the quantity to a decimal.New object and sets the Floating to 18.
			if value := decimal.New(quantity).Floating(18); value > 0 {

				// The purpose of this statement is to convert the "To" address of the transaction (tx.To) into a hexadecimal string
				// and assign it to the "To" field of the item object.
				item.To = address.New(tx.To).Hex()

				// This code is executing a query to determine if the user ID associated with the address and platform exists. If the
				// user ID is greater than 0, the code sets the symbol, chain ID, platform, financial type, transaction type, value,
				// hash and block of the item.
				if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2", item.GetTo(), chain.GetPlatform()).Scan(&item.UserId); item.GetUserId() > 0 {

					// This code is setting properties of an item object. Specifically, it is setting the symbol of a parent chain, the
					// id of a chain, the platform, the financial type, the transaction type, the value, the hash, and the block.
					item.Symbol = chain.GetParentSymbol()
					item.ChainId = chain.GetId()
					item.Platform = chain.GetPlatform()
					item.FinType = pbspot.FinType_CRYPTO
					item.TxType = pbspot.TxType_DEPOSIT
					item.Value = value
					item.Hash = tx.Hash
					item.Block = chain.GetBlock()
				}
			}

			break
		case blockchain.TypeContract:

			var (
				contract pbspot.Contract
			)

			// This code is attempting to retrieve logs associated with a transaction using the client's LogByTx method. It is
			// checking for any errors that occur during the retrieval process. If an error occurs, it will return without doing
			// anything else.
			logs, err := client.LogByTx(tx.Hash)
			if err != nil {
				return
			}

			// This is an if statement that is checking if the Data field of the logs variable is not nil. If it is not nil, then
			// it will execute the code in the if statement.
			if logs.Data != nil {

				// This code is checking a database table for a particular address and storing the corresponding symbol, protocol,
				// and decimals values in the contract object. If the query does not return a result, the "continue" statement will
				// prevent any further processing, and move on to the next iteration of the loop.
				if err := e.Context.Db.QueryRow("select symbol, protocol, decimals from contracts where lower(address) = $1", address.New(tx.To).Hex()).Scan(&contract.Symbol, &contract.Protocol, &contract.Decimals); err != nil { // No debug....
					continue
				}

				// The purpose of the above code is to check if the length of the "Topics" field in the "logs" object is equal to 3.
				// If it is, the code inside the if statement will be executed.
				if len(logs.Topics) == 3 {

					// The purpose of the code is to generate a cryptographic hash of the string "Transfer(address,address,uint256)".
					// This hash can be used to verify the authenticity of the string or data.
					transfer := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

					// Checking the method itself.
					switch logs.Topics[0] {
					case transfer.Hex():

						// This code is part of a loop that is attempting to access metadata from a blockchain. The purpose of this
						// specific code snippet is to parse a JSON object stored in the blockchain's main metadata using an ABI
						// (Application Binary Interface) and store it in the tokenAbi variable, so the data can be accessed later. If any
						// errors are encountered in the process, the code will skip the current iteration and continue with the loop.
						tokenAbi, err := abi.JSON(strings.NewReader(blockchain.MainMetaData.ABI))
						if e.Context.Debug(err) {
							continue
						}

						// The purpose of this code is to unpack a "Transfer" event from the "logs.Data" using the "tokenAbi" ABI. If an
						// error occurs, the code will continue on to the next line.
						instance, err := tokenAbi.Unpack("Transfer", logs.Data)
						if e.Context.Debug(err) {
							continue
						}

						// The code is checking that the element at the index 0 of the instance array is of type *big.Int. If it is, it
						// assigns the element to the number variable and sets ok to true, otherwise ok is set to false.
						if number, ok := instance[0].(*big.Int); ok {

							// The if statement is used to check if the value of the decimal.New(number).Floating(contract.GetDecimals()) is
							// greater than 0. If it is, then the code inside the block will be executed.
							if value := decimal.New(number).Floating(contract.GetDecimals()); value > 0 {

								// item.To is a variable used to store the address of the recipient of a transaction. The purpose of the code is
								// to convert the address stored in logs.Topics[2] (which is a string) to a hexadecimal value and store it in the
								// item.To variable.
								item.To = address.New(logs.Topics[2].(string)).Hex()

								// This code is querying a database to locate a user ID associated with a wallet address, platform, and protocol.
								// If a user ID is found and is greater than 0, then the item associated with that user is set to various values,
								// such as symbol, protocol, chain ID, platform, financial type, transaction type, value, hash, and block.
								if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2 and protocol = $3", item.GetTo(), chain.GetPlatform(), contract.GetProtocol()).Scan(&item.UserId); item.GetUserId() > 0 {

									// This code is setting properties of an item object. Specifically, it is setting the symbol of a parent chain, the
									// id of a chain, the platform, the financial type, the transaction type, the value, the hash, and the block.
									item.Symbol = contract.GetSymbol()
									item.Protocol = contract.GetProtocol()
									item.ChainId = chain.GetId()
									item.Platform = chain.GetPlatform()
									item.FinType = pbspot.FinType_CRYPTO
									item.TxType = pbspot.TxType_DEPOSIT
									item.Value = value
									item.Hash = tx.Hash
									item.Block = chain.GetBlock()
								}
							}
						}
					}
				}
			}
			break
		}

		// The purpose of this code is to check if the value of the item is greater than 0. If the value is greater than 0,
		// then the code inside the if statement will be executed.
		if item.GetValue() > 0 {

			// This code is setting up a transaction and checking for errors. If an error is encountered, the code will return and
			// stop further execution. This is a way to make sure that the transaction is handled correctly, and that any
			// potential errors are addressed.
			transaction, err := e.setTransaction(&item)
			if e.Context.Debug(err) {
				return
			}

			// The purpose of this code is to publish a transaction to an exchange, with a routing key of "deposit/open" and
			// "deposit/status". If there is an error, the code will print out the error and return.
			if err := e.Context.Publish(transaction, "exchange", "deposit/open", "deposit/status"); e.Context.Debug(err) {
				return
			}
		}
	}

	// This code is updating a database with the new block information. The if statement is used to check for any errors
	// that may occur during the database update, and the e.Context.Debug(err) will log any errors that occur. If an error
	// is encountered, the return statement will be executed, causing the code to exit without updating the database.
	if _, err := e.Context.Db.Exec("update chains set block = $1 where id = $2;", chain.GetBlock()+1, chain.GetId()); e.Context.Debug(err) {
		return
	}

	// This statement assigns the block associated with the provided chain ID to the 'block' element of the 'e' object. This
	// statement is used to store the block in the e object, so that it can be accessed later.
	e.block[chain.GetId()] = chain.GetBlock()

	// The purpose of e.done(chain.GetId()) is to execute the callback function associated with the e.done() method once the
	// chain.GetId() method has completed. This allows the code to wait for the chain.GetId() method to fully complete
	// before executing any further code.
	e.done(chain.GetId())
}

// depositTron - The purpose of this code is to deposit cryptocurrency on a blockchain. It checks the block number and goes through the
// list of transactions on the blockchain to find deposits. It then checks the type of transaction and parses the data to
// check if the deposit is valid. If it is valid, it sets up the transaction and publishes it. Finally, it updates the
// block number so the next deposit can be checked.
func (e *Service) depositTron(chain *pbspot.Chain) {

	// The purpose of this code is to handle a panic (or run-time error) that may occur during execution. The defer keyword
	// is used to ensure that the function is run even if the code panics. The recover() function returns the value that was
	// passed to the panic() function, which is then passed to the Context.Debug() function for logging. If the
	// Context.Debug() function returns true, the function will return without further execution. This allows the code to
	// continue running without crashing.
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	// This code is establishing a connection to a blockchain platform using the RPC protocol. The first line creates a
	// client connection to the blockchain using the RPC URL and platform credentials. The second line checks for any errors
	// that may have occurred during the connection process. If there is an error, the function will terminate.
	client, err := blockchain.Dial(chain.GetRpc(), chain.GetPlatform())
	if e.Context.Debug(err) {
		return
	}

	// This code is using the function BlockByNumber() from the client library to get a block from the blockchain. The
	// function returns a BlockBy object and an error. If an error is returned, the code will not continue and instead
	// return. This ensures that errors are not ignored and the program does not crash.
	blockBy, err := client.BlockByNumber(chain.GetBlock())
	if err != nil { // No debug....
		return
	}

	// The purpose of the above code is to loop through all the transactions in the blockBy object and perform operations on
	// each transaction. The underscore character is a blank identifier which is used when the loop variable will not be used.
	for _, tx := range blockBy.Transactions {

		var (
			item pbspot.Transaction
		)

		// The switch statement is used to execute a block of code based on the value of a certain expression. In this case,
		// the value of the expression is the tx.Type variable. Depending on the value of tx.Type, different blocks of code can be executed.
		switch tx.Type {
		case blockchain.TypeInternal: // TRX parse transfer coin.

			// This code is attempting to convert the "tx.Value" string into a float64 type, which is saved in the variable
			// "value". If the conversion results in an error, the e.Context.Debug(err) function is used to log the error and return.
			value, err := strconv.ParseFloat(tx.Value, 64)
			if e.Context.Debug(err) {
				return
			}

			// This is an if statement, which is a type of conditional statement. It checks to see if the value is greater than
			// zero. If it is, the code within the statement will be executed. If not, the code will be skipped.
			if value > 0 {

				// The purpose of this code is to convert the To address from a transaction (tx) to a Base58 encoded address. This is
				// a standard format for sending and receiving cryptocurrency, and is used to ensure that the address is valid and
				// can be used for the intended purpose.
				item.To = address.New(tx.To).Base58()

				// This code is querying the wallets table to find the user_id associated with a particular address, platform, and
				// item. If the user_id is successfully found, it then sets the symbol, chain id, platform, financial type,
				// transaction type, value, hash, and block associated with the item.
				if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2", item.GetTo(), chain.GetPlatform()).Scan(&item.UserId); item.GetUserId() > 0 {

					// This code is setting properties of an item object. Specifically, it is setting the symbol of a parent chain, the
					// id of a chain, the platform, the financial type, the transaction type, the value, the hash, and the block.
					item.Symbol = chain.GetParentSymbol()
					item.ChainId = chain.GetId()
					item.Platform = chain.GetPlatform()
					item.FinType = pbspot.FinType_CRYPTO
					item.TxType = pbspot.TxType_DEPOSIT
					item.Value = decimal.New(value).Div(1000000).Float() // value / 1000000
					item.Hash = tx.Hash
					item.Block = chain.GetBlock()
				}
			}

			break
		case blockchain.TypeContract: // Smart contract trigger transfer token.

			var (
				contract pbspot.Contract
			)

			// This code is retrieving logs associated with the given transaction, indicated by the tx.Hash parameter. It is using
			// the client's LogByTx method to achieve this and is checking for an error. If an error is encountered, the function is returned.
			logs, err := client.LogByTx(tx.Hash)
			if err != nil {
				return
			}

			// This statement is checking if the variable logs.Data is not equal to nil (null). If it is not equal to nil, then
			// some code will be executed. This is a common way of checking if a variable is initialized or not.
			if logs.Data != nil {

				// This piece of code is attempting to query the database to find the symbol, protocol, and decimals of a given
				// contract address. If it is unable to find the information, the code will skip the rest of the loop and continue on
				// with the next iteration.
				if err := e.Context.Db.QueryRow("select symbol, protocol, decimals from contracts where address = $1", address.New(tx.To).Base58()).Scan(&contract.Symbol, &contract.Protocol, &contract.Decimals); err != nil { // No debug....
					continue
				}

				// This is an if statement which checks the length of the "logs.Topics" array. If the length of the array is equal to
				// 3, then the code within the if statement will be executed.
				if len(logs.Topics) == 3 {

					// The purpose of the code snippet is to calculate the Keccak-256 hash of the string
					// "Transfer(address,address,uint256)". Keccak-256 is a cryptographic hash algorithm used to generate a unique,
					// fixed-size 256-bit (32-byte) hash. The hash is used as a unique identifier for the string, which can be used to
					// verify its integrity.
					transfers := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

					// This code is checking to see if the first element in the "logs.Topics" array is equal to the result of the
					// "strings.TrimPrefix" function applied to the "transfers.Hex()" function.
					// The "strings.TrimPrefix" function removes the "0x" prefix from the string returned by the "transfers.Hex()" function.
					if logs.Topics[0] == strings.TrimPrefix(transfers.Hex(), "0x") {

						// This code is attempting to read the ABI of a blockchain. The tokenAbi variable stores the ABI from the
						// blockchain's MainMetaData. The ABI is read in from a string and parsed as a JSON object. If an error occurs
						// while parsing the ABI, the error is logged and the code continues.
						tokenAbi, err := abi.JSON(strings.NewReader(blockchain.MainMetaData.ABI))
						if e.Context.Debug(err) {
							continue
						}

						// This code snippet is attempting to unpack a log from a smart contract using an ABI (Application Binary
						// Interface). The instance variable will contain the output of the unpacking while err will contain any errors
						// that occur during the unpacking process. The if statement then checks for any errors that have occurred, and if
						// there are any it will skip the rest of the code and continue.
						instance, err := tokenAbi.Unpack("Transfer", logs.Data)
						if e.Context.Debug(err) {
							continue
						}

						// This is an example of type assertion. It is used to check if the type of instance[0] is a *big.Int, and if it
						// is, it checks if the Int64() of the *big.Int is greater than 0. If both of these conditions are true, then the
						// if statement is executed.
						if number, ok := instance[0].(*big.Int); ok || number.Int64() > 0 {

							// The purpose of this code is to convert a number from a decimal to a floating point value. The value is then
							// compared to 0 to determine if it is greater than 0 or not. If it is greater than 0, a certain action is taken.
							if value := decimal.New(number).Floating(contract.GetDecimals()); value > 0 {

								// This code is used to convert a given log's topics element at index 2 from its original form (a string) into a
								// Base58 encoded version. This is often used when dealing with cryptographic addresses, as Base58 is a format
								// commonly used to represent them.
								item.To = address.New(logs.Topics[2].(string)).Base58()

								// This code is querying a database to find the user_id associated with a particular address, platform, and
								// protocol in order to update the item with symbol, protocol, chain id, platform, financial type, transaction
								// type, value, hash and block. The if statement is used to check if the user_id is greater than 0, indicating
								// that the query was successful and the item can be updated.
								if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2 and protocol = $3", item.GetTo(), chain.GetPlatform(), contract.GetProtocol()).Scan(&item.UserId); item.GetUserId() > 0 {

									// This code is setting properties of an item object. Specifically, it is setting the symbol of a parent chain, the
									// id of a chain, the platform, the financial type, the transaction type, the value, the hash, and the block.
									item.Symbol = contract.GetSymbol()
									item.Protocol = contract.GetProtocol()
									item.ChainId = chain.GetId()
									item.Platform = chain.GetPlatform()
									item.FinType = pbspot.FinType_CRYPTO
									item.TxType = pbspot.TxType_DEPOSIT
									item.Value = value
									item.Hash = tx.Hash
									item.Block = chain.GetBlock()
								}
							}
						}
					}
				}
			}
			break
		}

		// The purpose of this code is to check if the value of the item is greater than 0. If it is, then the code inside the
		// if statement will be executed.
		if item.GetValue() > 0 {

			// This code is setting up a transaction for an item, and then checking for any errors that may occur during that
			// transaction. If an error is detected, the code will return and terminate the transaction.
			transaction, err := e.setTransaction(&item)
			if e.Context.Debug(err) {
				return
			}

			// This code is intended to publish a transaction to an exchange using the "deposit/open" and "deposit/status" routing
			// keys. The if statement is checking for an error during the publishing process and returning if one is found. The
			// e.Context.Debug call is used to log the error for further investigation.
			if err := e.Context.Publish(transaction, "exchange", "deposit/open", "deposit/status"); e.Context.Debug(err) {
				return
			}
		}
	}

	// This code is used to update the block number associated with a particular chain ID in a database. The first line of
	// code, `if _, err := e.Context.Db.Exec("update chains set block = $1 where id = $2;", chain.GetBlock()+1,
	// chain.GetId()); e.Context.Debug(err) {`, attempts to execute an SQL statement that updates the block number for the
	// associated ID. If an error occurs, the error is logged and the code returns.
	if _, err := e.Context.Db.Exec("update chains set block = $1 where id = $2;", chain.GetBlock()+1, chain.GetId()); e.Context.Debug(err) {
		return
	}

	// This line of code is used to store a block in a blockchain. The block[chain.GetId()] is used to refer to a specific
	// block in the chain, and the chain.GetBlock() is used to retrieve this block from the chain. This line of code is used
	// to ensure that the retrieved block is stored in the blockchain, so that it can be used later.
	e.block[chain.GetId()] = chain.GetBlock()

	// e.done(chain.GetId()) is a function used to retrieve the ID of a completed chain. It is used to get the ID of a chain
	// after it has been processed and completed. This is useful when tracking the progress of a chain or when performing
	// operations on a specific chain.
	e.done(chain.GetId())
}

// transferEthereum - This function is used in a blockchain application to transfer Ethereum. It performs a variety of actions such as
// dialing the correct RPC, creating a keypair and a private key, estimating the gas for the transaction, setting a
// reserve account for the funds being transferred, and setting the reserve account to unlock. Finally, it publishes the
// transaction to the exchange and sends out an email notification.
func (e *Service) transferEthereum(userId, txId int64, symbol string, to string, value, price float64, protocol pbspot.Protocol, chain *pbspot.Chain, subscribe bool) {

	//This code is used to handle the panic situations in a program. The defer statement ensures that the function
	//following it will be executed either when the function returns normally or when the function panics. In this code,
	//the function will recover any panic situation, log the panic (if it is enabled in the context) and then return from the function.
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	// The code snippet creates several variables that are used later in the program. The variables are of various types,
	// such as keypair.CrossChain, query.Migrate, float64, blockchain.Transfer, and big.Int. These variables are used to
	// store data that will be needed throughout the program, such as fees, convert, transfer, and wei.
	var (
		cross   keypair.CrossChain
		migrate = query.Migrate{
			Context: e.Context,
		}
		fees, convert float64
		transfer      *blockchain.Transfer
		wei           *big.Int
	)

	// This code is establishing a connection between a client and a blockchain. The blockchain.Dial() function is used to
	// create a new connection and returns a client instance and an error. The chain.GetRpc() and chain.GetPlatform()
	// functions are used to get the URL and platform (e.g. Ethereum) to which the client should connect. The if statement
	// is then used to check for any errors and if there is an error, then the code will return.
	client, err := blockchain.Dial(chain.GetRpc(), chain.GetPlatform())
	if e.Context.Debug(err) {
		return
	}

	// The purpose of this code is to retrieve entropy for a given userId and then check for errors using the
	// Context.Debug() function. If an error is found, the function returns.
	entropy, err := e.getEntropy(userId)
	if e.Context.Debug(err) {
		return
	}

	// This code is creating a new owner object using the cross package. The fmt package is used to format the
	// e.Context.Secrets[1] into a specific string format. The entropy and chain.GetPlatform() parameters are also passed to
	// the cross.New function. The code is checking for any errors with the new owner object and if there is an error, the code is returning.
	owner, private, err := cross.New(fmt.Sprintf("%v-&*39~763@)", e.Context.Secrets[1]), entropy, chain.GetPlatform())
	if e.Context.Debug(err) {
		return
	}

	// This code is used to convert a hexadecimal-encoded private key into an ECDSA (Elliptic Curve Digital Signature
	// Algorithm) private key. The strings.TrimPrefix function is used to remove the "0x" prefix from the private key string
	// before it is converted. If there is an error during the conversion, to err variable is set to indicate this and the
	// code execution is stopped by returning from the function.
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(private, "0x"))
	if e.Context.Debug(err) {
		return
	}

	// client.Private(privateKey) is used to create a new account associated with the provided private key. This is used to
	// access the account using the private key to sign transactions and messages from the client.
	client.Private(privateKey)

	// This line of code sets the network of the client to the network of the chain. This allows the client to communicate
	// with the chain and make requests to it.
	client.Network(chain.GetNetwork())

	// This is a conditional statement that checks if the protocol is equal to the constant Protocol_MAINNET which is
	// defined in the pbspot package. If the protocol is equal to Protocol_MAINNET, then the code inside the if statement will be executed.
	if protocol == pbspot.Protocol_MAINNET {

		// The purpose of this code is to create a Transfer variable which will store information about a transaction on a
		// blockchain. The variable will store the address of the recipient (To), the amount of gas required for the
		// transaction (Gas), and the value of the transaction in the smallest denomination of the cryptocurrency (Value).
		transfer = &blockchain.Transfer{
			To:    to,
			Gas:   21000,
			Value: decimal.New(value).Integer(18),
		}

		// EstimateGas is a function used to estimate the gas required to execute a given transaction. The code snippet is
		// attempting to estimate the gas required for the transaction "transfer" and checks for errors. If an error is found,
		// it will be logged in the context and the function will return.
		estimate, err := client.EstimateGas(transfer)
		if e.Context.Debug(err) {
			return
		}

		// The purpose of this code is to create a new decimal value with 18 decimal places of precision, based on the
		// `estimate` value that was provided. It is done using the `decimal.New()` and `Floating()` functions from the decimal library.
		fees = decimal.New(estimate).Floating(18)

		// This code is used to calculate the number of wei (a unit of account used in the Ethereum blockchain) associated with
		// a transaction. If the user subscribes, the amount of wei is calculated by subtracting the fees from the value. If
		// the user does not subscribe, the amount of wei is simply the value.
		if subscribe {
			wei = decimal.New(decimal.New(value).Sub(fees).Float()).Integer(18)
		} else {
			wei = decimal.New(value).Integer(18)
		}

		// The purpose of the code is to create a new value of 1.5 wei and assign it to the variable 'transfer.Value'. Wei is a unit of ether, a cryptocurrency.
		transfer.Value = wei
	} else {

		// This code is checking for errors from the getContract() function, which is presumably retrieving a contract from a
		// chain. If the function returns an error, the code will print the error and then exit the function.
		contract, err := e.getContract(symbol, chain.GetId())
		if e.Context.Debug(err) {
			return
		}

		// This code is used to send a given amount of tokens to a given address. The data variable is used to store the
		// transaction data, which is calculated by converting the given value to an integer, taking into account the decimal
		// places of the contract. To err variable is used to store any errors that occur in the process. If an error occurs, the function returns.
		data, err := client.Data(to, decimal.New(value).Integer(contract.GetDecimals()).Bytes())
		if err != nil {
			return
		}

		// The purpose of the above code is to create a blockchain transfer with the given contract address, gas value, and
		// data. The transfer will be used to execute a transaction on the blockchain.
		transfer = &blockchain.Transfer{
			Contract: contract.GetAddress(),
			Gas:      65000,
			Data:     data,
		}

		// EstimateGas is used to estimate the gas required for a given transaction. In this example, the estimate is stored in
		// the variable estimate and then an error check is done with the Context.Debug function. If there is an error, the function returns.
		estimate, err := client.EstimateGas(transfer)
		if e.Context.Debug(err) {
			return
		}

		// The purpose of this line of code is to convert a decimal value to a floating-point value with 18 decimal places. The
		// decimal.New() function takes an estimate as an argument and the Floating() method is used to convert the value to a
		// floating-point type with 18 decimal places.
		fees = decimal.New(estimate).Floating(18)

		// The purpose of this code is to calculate the total cost of a product by multiplying the fees and the price together,
		// and then converting the result to a floating-point number.
		convert = decimal.New(fees).Mul(price).Float()

		// This code is used to transfer a given value from one address to another, using the client Data() method. The value
		// is first converted from a decimal to a float and then from a float to an integer, before being sent to the contract
		// in the form of bytes. If there is an error, the code returns.
		data, err = client.Data(to, decimal.New(decimal.New(value).Sub(convert).Float()).Integer(contract.GetDecimals()).Bytes())
		if err != nil {
			return
		}

		// The purpose of the statement "transfer.Data = data" is to assign the value of the variable "data" to the "Data"
		// property of the "transfer" object.
		transfer.Data = data
	}

	// This code is used to transfer funds using a client. The hash variable is used to store the hash value returned by the
	// transfer. The if statement is used to check for any errors that occur during the transfer. If an error is found, it
	// is logged with the Context.Debug() method, and the function is returned.
	hash, err := client.Transfer(transfer)
	if e.Context.Debug(err) {
		return
	}

	// The purpose of this code is to check for an error when executing a 'client.Transaction()' command, and if an error is
	// encountered, the 'e.Context.Debug(err)' command is used to log the error, and then the function will return without
	// executing any further code.
	if err = client.Transaction(); e.Context.Debug(err) {
		return
	}

	// The purpose of this statement is to check if the variable "subscribe" is true. If it is, the code that follows the
	// statement will be executed.
	if subscribe {

		var (
			charges float64
		)

		// This is an if statement used to determine which protocol should be used. In this case, it is checking if the
		// protocol is set to the mainnet protocol. If it is, then the code within the statement will be executed. If it is
		// not, then the code will not be executed and the program will continue with the next statement.
		if protocol == pbspot.Protocol_MAINNET {

			// This code is a part of a function that is attempting to set a reserve for a user on a given platform and chain. The
			// purpose of the if statement is to check if an error occurs when the reserve is being set. If an error occurs, the
			// function will return and stop executing. The e.Context.Debug(err) is used to log the error, so that it can be investigated later.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), value, chain.GetPlatform(), pbspot.Protocol_MAINNET, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// The purpose of this equation is to calculate the total charges associated with a purchase. The charges are equal to the sum of the fees and taxes associated with the purchase.
			charges = fees

		} else {

			// This code is checking to see if a reserve is set for a given userId, owner, chain symbol, fees, platform, and
			// balance. If the reserve is set, it will continue the code, but if there is an error, it will debug the error and then return.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), fees, chain.GetPlatform(), pbspot.Protocol_MAINNET, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Update the reserve account, the amount that was deposited for the withdrawal of the token is converted and debited in a partial amount, excluding commission, for example:
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); This amount is 985.3 tst and will be overwritten without commission.
			if err := e.setReserve(userId, owner, symbol, decimal.New(value).Sub(convert).Float(), chain.GetPlatform(), protocol, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Token commission - update the collection account, the commission that was deducted from the amount of the token is credited to the exchange, the commission is calculated according to the formula and commission of the parent account,
			// for example: since we make commissions for the transfer exclusively from the parent account, we need to minus the commission of 0.006 eth from the amount of 1000 tst using conversion at the price of the token,
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); this amount is 985.3 tst and will be credited for the transfer,
			// and the amount of 14.7 tst will be credited to the stock exchange, since the reverse was charged at the expense of the stock exchange to pay for gas.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges + $2 where symbol = $1;", symbol, convert); e.Context.Debug(err) {
				return
			}

			// This line of code multiplies the fees by the price and converts the result to a float. In other words, this line of
			// code is used to calculate the charges associated with a particular purchase.
			charges = decimal.New(fees).Mul(price).Float()
		}

		// This code is executing an SQL statement to update the transactions table. It is setting the fees, hash, and status
		// of a transaction with a specific ID. The e.Context.Debug(err) line is used to check for any errors that may have
		// occurred during the update and, if any errors are found, the function will return.
		if _, err := e.Context.Db.Exec("update transactions set fees = $4, hash = $3, status = $2 where id = $1;", txId, pbspot.Status_FILLED, hash, fees); e.Context.Debug(err) {
			return
		}

		// This piece of code is used to publish a transaction message on a message broker. The message contains the
		// transaction ID, fees, and hash. The message is sent to the exchange topic with the label "withdraw/status". The code
		// also checks for an error and returns if there is one.
		if err := e.Context.Publish(&pbspot.Transaction{
			Id:   txId,
			Fees: charges,
			Hash: hash,
		}, "exchange", "withdraw/status"); e.Context.Debug(err) {
			return
		}

		// This code is likely a function call to the migrate package in a programming language. The purpose of this code is to
		// email the user identified by the userId parameter with the subject "withdrawal", the value of the
		// withdrawal, and the symbol associated with the withdrawal.
		go migrate.SendMail(userId, "withdrawal", value, symbol)

	} else {

		// This is an if statement that is checking the value of the variable "protocol" against a specific value,
		// pbspot.Protocol_MAINNET. If the value of the protocol variable is equal to the value of pbspot.Protocol_MAINNET,
		// then the code within the statement will be executed. This is used to ensure that the code is only executed if the
		// protocol is set to the correct value.
		if protocol == pbspot.Protocol_MAINNET {

			// This statement is used to add the reserve amount for a user to their account. The parameters passed to the
			// setReserve function are the userId, owner, chain symbol, amount, platform, protocol, and balance type. The
			// statement then checks for any errors that occurred while setting the reserve, and if there was an error, the statement will return.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), decimal.New(value).Add(fees).Float(), chain.GetPlatform(), pbspot.Protocol_MAINNET, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// The fee for the exchange net is double the value for the transfer gas, and the amount for the token withdrawal fee.
			// When replenishing the transfer wallet to pay the commission, a double commission is withdrawn.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges - $2, fees_costs = fees_costs + $2 where symbol = $1;", chain.GetParentSymbol(), decimal.New(value).Add(fees).Float()); e.Context.Debug(err) {
				return
			}

		}
	}

	// This code is making sure that a reserve is unlocked in order to allow a user with a given ID, symbol, platform, and
	// protocol to access it. The "if err" statement is checking for any errors that might occur when attempting to set the
	// reserve unlock. If an error is encountered, the code will return, otherwise it will continue with execution.
	if err := e.setReserveUnlock(userId, symbol, chain.GetPlatform(), protocol); e.Context.Debug(err) {
		return
	}
}

// transferTron - This function is used to transfer Tron (a cryptocurrency) from one user to another. It takes the user ID, transaction
// ID, symbol, address to send to, value, price, protocol, chain and whether to subscribe or not as parameters. The
// function then attempts to dial the chain's RPC and get the owner, private key, and client. It then checks the protocol
// and sets the transfer accordingly, calculating the fees and converting the value if needed. It then attempts to
// transfer the Tron and if successful, updates the reserve and transaction status accordingly. If there is an error, the
// function will panic and debug the error. Finally, it unlocks the reserve account.
func (e *Service) transferTron(userId, txId int64, symbol, to string, value, price float64, protocol pbspot.Protocol, chain *pbspot.Chain, subscribe bool) {

	// This code is used to handle an unexpected error. The defer statement prevents the program from panicking, and
	// instead, recovers from the error and prints out a debug message. This allows the program to continue functioning
	// normally without crashing.
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	// The code snippet declares six variables - cross, migrate, fees, convert, transfer, and wei - each with a different
	// type. The purpose of this is to declare variables with different types that will be used later in the code. The types
	// of the variables indicate that they will be used for a keypair, query, float calculations, blockchain transfer, and
	// big integer calculations, respectively.
	var (
		cross   keypair.CrossChain
		migrate = query.Migrate{
			Context: e.Context,
		}
		fees, convert float64
		transfer      *blockchain.Transfer
		wei           *big.Int
	)

	// This code is used to establish a connection between the blockchain and a client. The first line of code attempts to
	// dial the blockchain and the client using the GetRpc() and GetPlatform() methods. The second line checks for any
	// errors that may have occurred during the dialing process and prints them out if the debug flag is set to true. If
	// there are errors, the code execution is stopped and no further action is taken.
	client, err := blockchain.Dial(chain.GetRpc(), chain.GetPlatform())
	if e.Context.Debug(err) {
		return
	}

	// This code is likely part of a larger program that is attempting to get entropy (a measure of randomness) from a user
	// ID. The first line attempts to retrieve the entropy from a user ID and the second line checks if that action caused
	// an error. If an error occurred, the code will exit the function.
	entropy, err := e.getEntropy(userId)
	if e.Context.Debug(err) {
		return
	}

	// This code is creating a new "cross" object that is used to securely store data. The owner and private variables are
	// used to reference the cross object. The fmt.Sprintf() function is used to format a string using the
	// e.Context.Secrets[1] value. The entropy and chain.GetPlatform() values are used to provide additional security
	// parameters. The e.Context.Debug() function is used to log any errors that occur during the creation of the cross object.
	owner, private, err := cross.New(fmt.Sprintf("%v-&*39~763@)", e.Context.Secrets[1]), entropy, chain.GetPlatform())
	if e.Context.Debug(err) {
		return
	}

	// This code is used to convert a hexadecimal string to an Elliptic Curve Digital Signature Algorithm (ECDSA) private
	// key. The crypto.HexToECDSA() function takes in a hexadecimal string as an argument, and the strings.TrimPrefix()
	// function is used to remove the "0x" prefix from the string. The if statement checks for errors, and if an error
	// occurs, the code returns without completing the rest of the code.
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(private, "0x"))
	if e.Context.Debug(err) {
		return
	}

	// The client.Private() function is used to set the private key associated with the client. This private key is then
	// used by the client to authenticate itself to the server, allowing access to protected resources on the server.
	client.Private(privateKey)

	// This line of code is checking if the protocol used is the mainnet protocol. If it is, then the code will continue to
	// execute. If it is not, then it will jump to the next line of code.
	if protocol == pbspot.Protocol_MAINNET {

		// The purpose of this code is to create a transfer record for a blockchain transaction. The transfer record contains
		// information about the recipient address, the amount of gas used for the transaction, and the value of the
		// transaction. The recipient address is converted to hexadecimal to make it easier to read, and the value is converted
		// to an integer with 6 decimal places.
		transfer = &blockchain.Transfer{
			To:    address.New(to).Hex(true),
			Gas:   10000000,
			Value: decimal.New(value).Integer(6),
		}

		// EstimateGas is used to determine the amount of gas needed to execute a transaction on the Ethereum network. The
		// estimate, err := client.EstimateGas(transfer) line of code is used to call the EstimateGas function on the client
		// and assign the result to the estimate variable. The if e.Context.Debug(err) { return } line of code is used to check
		// for any errors that may have occurred during the EstimateGas function call. If an error is present, the return keyword will terminate the function.
		estimate, err := client.EstimateGas(transfer)
		if e.Context.Debug(err) {
			return
		}

		// This statement is used to set the fees variable to the value of estimate, rounded to 6 decimal places. The
		// decimal.New() method is used to convert the estimate value from a float to a decimal, and Floating() is used to
		// round the decimal to 6 decimal places.
		fees = decimal.New(estimate).Floating(6)

		// This code is used to calculate the value of 'wei' depending on the boolean value of the 'subscribe' variable. If
		// 'subscribe' is true, then 'wei' is set to the difference between 'value' and 'fees'. Otherwise, 'wei' is set to the
		// value of 'value'.
		if subscribe {
			wei = decimal.New(decimal.New(value).Sub(fees).Float()).Integer(6)
		} else {
			wei = decimal.New(value).Integer(6)
		}

		// The purpose of this line of code is to assign a value of wei to the transfer variable. Wei is a unit of measurement
		// used in the Ethereum blockchain to represent a tiny fraction of Ether (the cryptocurrency).
		transfer.Value = wei
	} else {

		// This code is used to retrieve the contract associated to the given symbol and chain ID. It checks for any errors
		// that could have occurred during the retrieval process. If an error is found, the code will stop and return.
		contract, err := e.getContract(symbol, chain.GetId())
		if e.Context.Debug(err) {
			return
		}

		// This code is used to set the data for a client address and value. Specifically, it creates a new hexadecimal address
		// for the recipient (to), converts the value to an integer using the contract's decimal value, and then stores the
		// data in bytes. If an error occurs, the code returns without completing the task.
		data, err := client.Data(address.New(to).Hex(true), decimal.New(value).Integer(contract.GetDecimals()).Bytes())
		if err != nil {
			return
		}

		// The purpose of the code is to create a transfer object that contains the contract address, the amount of gas needed
		// to complete the transaction, and the data necessary to complete the transfer. This object is then used to make a blockchain transaction.
		transfer = &blockchain.Transfer{
			Contract: address.New(contract.GetAddress()).Hex(true),
			Gas:      10000000,
			Data:     data,
		}

		// EstimateGas is a function from the client that takes a transfer as an argument and returns an estimate of the amount
		// of gas that would be required for the transfer. The if statement is checking for any errors that may occur during
		// the estimation process, and if there is an error, the function returns.
		estimate, err := client.EstimateGas(transfer)
		if e.Context.Debug(err) {
			return
		}

		// The purpose of this line of code is to convert an estimate (which may be an integer or a float) to a decimal type
		// with 6 decimal places. This can be used for example to price items in a shopping cart or to calculate taxes.
		fees = decimal.New(estimate).Floating(6)

		// This code is used to convert a number from one data type to another. In this case, it is used to multiply two
		// decimal values (fees and price) and convert the result to a float data type.
		convert = decimal.New(fees).Mul(price).Float()

		// This code is used to send a given amount of tokens to a given address. The data variable is used to store the
		// transaction data, which is calculated by converting the given value to an integer, taking into account the decimal
		// places of the contract. To err variable is used to store any errors that occur in the process. If an error occurs, the function returns.
		data, err = client.Data(address.New(to).Hex(true), decimal.New(decimal.New(value).Sub(convert).Float()).Integer(contract.GetDecimals()).Bytes())
		if err != nil {
			return
		}

		// The purpose of this line of code is to assign the value of the variable 'data' to the variable 'transfer.Data'.
		transfer.Data = data
	}

	// This code is using the client.Transfer() method to initiate a transfer from one account to another, and checking for
	// any errors. To err variable is used to capture any errors that may occur during the transfer. If
	// e.Context.Debug(err) returns true, then the code will return without doing anything else.
	hash, err := client.Transfer(transfer)
	if e.Context.Debug(err) {
		return
	}

	// The purpose of this code is to check for any errors that may occur when executing the client.Transaction() function.
	// If an error is returned, the e.Context.Debug(err) statement will log the error and then the code will return.
	if err = client.Transaction(); e.Context.Debug(err) {
		return
	}

	// The purpose of this statement is to check if the variable "subscribe" is true. If it is, the code that follows the
	// statement will be executed.
	if subscribe {

		var (
			charges float64
		)

		// The purpose of this code is to check whether the protocol is set to the mainnet protocol. If it is, the code will
		// continue to execute, otherwise it will not.
		if protocol == pbspot.Protocol_MAINNET {

			// This code is checking for an error when setting a reserve for a user, and if an error is present, it prints the
			// error to the debug log and then returns. This code is likely part of a larger function that sets a reserve for a
			// user with a particular value and settings, so this check is ensuring that the reserve is properly set and that any errors are logged in the debug log.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), value, chain.GetPlatform(), pbspot.Protocol_MAINNET, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// The purpose of this equation is to calculate the total amount of charges owed. The equation adds together fees and taxes to determine the total amount of charges.
			charges = fees
		} else {

			// This code checks if there is an error when calling the setReserve() function, which is used to reserve funds for a
			// specific user on a blockchain. If there is an error, the code will return without executing any further code. The
			// purpose of this code is to ensure that the setReserve() function executes correctly before any other code is executed.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), fees, chain.GetPlatform(), pbspot.Protocol_MAINNET, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Update the reserve account, the amount that was deposited for the withdrawal of the token is converted and debited in a partial amount, excluding commission, for example:
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); This amount is 985.3 tst and will be overwritten without commission.
			if err := e.setReserve(userId, owner, symbol, decimal.New(value).Sub(convert).Float(), chain.GetPlatform(), protocol, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Token commission - update the collection account, the commission that was deducted from the amount of the token is credited to the exchange, the commission is calculated according to the formula and commission of the parent account,
			// for example: since we make commissions for the transfer exclusively from the parent account, we need to minus the commission of 0.006 eth from the amount of 1000 tst using conversion at the price of the token,
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); this amount is 985.3 tst and will be credited for the transfer,
			// and the amount of 14.7 tst will be credited to the stock exchange, since the reverse was charged at the expense of the stock exchange to pay for gas.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges + $2 where symbol = $1;", symbol, convert); e.Context.Debug(err) {
				return
			}

			// The purpose of this code is to calculate the charge amount based on the fees and the price. The decimal.New(fees)
			// function creates a new decimal from the fees value, and then the Mul(price) function multiplies that value with the
			// price. Finally, the Float() function converts the result to a float value.
			charges = decimal.New(fees).Mul(price).Float()
		}

		// This code is performing an update query on the "transactions" table in a database. The purpose of this query is to
		// update the existing values for the specified record in the table with the given parameters. Specifically, the "fees"
		// column will be set to the value of the "fees" variable, the "hash" column will be set to the value of the "hash"
		// variable, the "status" column will be set to the value of the "pbspot.Status_FILLED" variable, and the "id" column
		// will be set to the value of the "txId" variable. The e.Context.Debug(err) line is checking if there is an error and returning if there is.
		if _, err := e.Context.Db.Exec("update transactions set fees = $4, hash = $3, status = $2 where id = $1;", txId, pbspot.Status_FILLED, hash, fees); e.Context.Debug(err) {
			return
		}

		// This code is using the Context Publish method to publish a transaction to an exchange. The transaction includes an
		// ID, fees, and hash. The purpose of this code is to send the transaction to the exchange and provide a status update
		// on the withdrawal. If there is an error, it will be logged and the process will return.
		if err := e.Context.Publish(&pbspot.Transaction{
			Id:   txId,
			Fees: charges,
			Hash: hash,
		}, "exchange", "withdraw/status"); e.Context.Debug(err) {
			return
		}

		// The purpose of this function is to email a user when they make a withdrawal from their account. The
		// parameters provided are the userId, the type of transaction (withdrawal), the amount of the transaction (value), and the currency symbol (symbol).
		go migrate.SendMail(userId, "withdrawal", value, symbol)

	} else {

		// This is a condition that checks if the protocol being used is the Mainnet protocol. If the protocol being used is
		// the Mainnet protocol, then the condition evaluates to true and the code following it will execute. If the protocol
		// being used is not the Mainnet protocol, then the condition evaluates to false and the code following it will not execute.
		if protocol == pbspot.Protocol_MAINNET {

			// This code is intended to set a reserve for a user on a blockchain platform. The user ID, owner, symbol, value,
			// platform, protocol and balance are all supplied to the setReserve function. If any errors occur during the process,
			// the program will return. Otherwise, the reserve will be set successfully.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), decimal.New(value).Add(fees).Float(), chain.GetPlatform(), pbspot.Protocol_MAINNET, pbspot.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// The fee for the exchange net is double the value for the transfer gas, and the amount for the token withdrawal fee.
			// When replenishing the transfer wallet to pay the commission, a double commission is withdrawn.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges - $2, fees_costs = fees_costs + $2 where symbol = $1;", chain.GetParentSymbol(), decimal.New(value).Add(fees).Float()); e.Context.Debug(err) {
				return
			}

		}
	}

	// This code is making sure that a reserve is unlocked in order to allow a user with a given ID, symbol, platform, and
	// protocol to access it. The "if err" statement is checking for any errors that might occur when attempting to set the
	// reserve unlock. If an error is encountered, the code will return, otherwise it will continue with execution.
	if err := e.setReserveUnlock(userId, symbol, chain.GetPlatform(), protocol); e.Context.Debug(err) {
		return
	}
}

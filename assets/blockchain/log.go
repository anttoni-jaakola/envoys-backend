package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/pkg/errors"
	"strings"
)

// LogByTx - The purpose of the code is to create an object of type Params and use it to retrieve the log associated with a given
// transaction ID. The code will first attempt to identify the platform of the transaction, then execute different API
// calls depending on the platform. If the platform is not recognized, an error will be returned. The code will then
// attempt to commit the process and if an error is encountered, it will return the log and the error. Finally, the log()
// function of the Params object will be called and the result will be returned.
func (p *Params) LogByTx(id string) (log *Log, err error) {

	// The purpose of the statement is to create a new object of type Log and assign it to the variable log.
	log = new(Log)

	// The purpose of this code is to provide a switch statement to identify the platform, then execute different API calls
	// depending on the platform. If the platform is Ethereum, it will execute the eth_getTransactionReceipt API call and if
	// the platform is Tron, it will execute the `gettransactioninfobyid` API call. If the platform is neither Ethereum nor
	// Tron, it will return an error.
	switch p.platform {
	case types.PlatformEthereum:
		p.query = []string{"-X", "POST", "-H", "Content-Type:application/json", "-H", "Accept:application/json", "-d", fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%s"],"id":1}`, id), p.rpc}
	case types.PlatformTron:
		p.query = []string{"-X", "POST", fmt.Sprintf("%v/wallet/gettransactioninfobyid", p.rpc), "-d", fmt.Sprintf(`{"value": "%v"}`, id)}
	default:
		return log, errors.New("method not found!...")
	}

	// The purpose of the code is to check for an error after attempting to commit a process. If an error is encountered,
	// the function will return the log and the error.
	if err := p.commit(); err != nil {
		return log, err
	}

	// The purpose of the following statement is to call the log() function of the object p and return the result of the
	// function call.
	return p.log()
}

// log - The purpose of this code is to parse log data from a response and create a new Log object to store the data. It uses a
// switch statement to determine which platform the response is from, and then uses if statements to check for the
// presence of the keys "result" and "log" in the response. If the keys are present, the code decodes a hexadecimal
// string into a byte array and assigns the byte array to the Data field of the Log struct. It also assigns any topics
// found in the response to the Topics field of the Log struct. Finally, the log is returned along with a nil value.
func (p *Params) log() (log *Log, err error) {

	// The purpose of the statement is to create a new object of type Log and assign it to the variable log.
	log = new(Log)

	// This switch statement is used to test the value of the variable p.platform. Depending on the value, the code within
	// the corresponding case statement will be executed. This allows you to execute different code depending on the value
	// of p.platform.
	switch p.platform {
	case types.PlatformEthereum:

		// This code is checking if the response from p has a key "result" and if it is a map[string]interface{}.
		// If both conditions are true, it will assign the value at "result" to the variable result and assign true to the variable ok.
		// This if statement is essentially checking if the response has a key "result" and if it is a valid type.
		if result, ok := p.response["result"].(map[string]interface{}); ok {

			// This code checks if the key "logs" exists in the "result" map. If it exists, the variable "vlog" will be assigned
			// the value of the key "logs" and the variable "ok" will be assigned the boolean value of true.
			if vlog, ok := result["logs"]; ok {

				// The purpose of this statement is to check if the variable vlog is assignable to a slice of interfaces. If it is,
				// the variable logs is assigned the value of vlog. The statement also returns a boolean value in the variable ok
				// which is true if the variable vlog is assignable to a slice of interfaces, and false otherwise.
				if logs, ok := vlog.([]interface{}); ok {

					// This loop is used to iterate through the elements of an array or slice called logs. The loop starts at the first index, 0, and continues until the index is no longer less than the length of the array/slice, which is stored in the len() function. The purpose of the loop is to access each element of the array/slice.
					for i := 0; i < len(logs); i++ {

						// This if statement is checking for the presence of a key in a map of type string to type interface{}. The map is called logs, and the key being searched for is "data".
						// The statement assigns the value associated with the key to the variable data, and assigns a boolean value to ok. If the key is present in the map, ok will be true, and the value associated with the key will be assigned to the data variable.
						// This statement is used to check the presence of the key in the map, and access the associated value if it is present.
						if data, ok := logs[i].(map[string]interface{})["data"]; ok {

							// This code is used to decode a hexadecimal string into a byte array. The strings.TrimPrefix method is used to
							// remove the "0x" prefix from the string, and then the hex.DecodeString method is used to decode the string into
							// a byte array. If an error occurs in the decoding process, the code returns a log and a nil value.
							data, err := hex.DecodeString(strings.TrimPrefix(data.(string), "0x"))
							if err != nil {
								return log, nil
							}

							// The purpose of the line of code log.Data = data is to assign the value of the variable 'data' to the field
							// 'Data' of the object 'log'. This allows the data to be stored within the log object.
							log.Data = data
						}

						// This code is used to check if the logs[i] map contains the key "topics". If it does, ok will be set to true and
						// the topics field of the log struct will be set to the value of the topics key in logs[i].
						if topics, ok := logs[i].(map[string]interface{})["topics"]; ok {
							log.Topics = topics.([]interface{})
						}
					}

				}
			}
		}

	case types.PlatformTron:

		// This code is checking whether an error has occurred, and if so, creating a new error with the error message. It is
		// used to handle errors that have occurred during an operation and report them to the user.
		if err, ok := p.response["Error"]; ok || err != nil {
			return log, errors.New(err.(string))
		}

		// This code is using an if statement to check if the key "log" is present in the map p.response. If it is found, the
		// ok variable will be set to true and the value of the key will be assigned to the vlog variable.
		if vlog, ok := p.response["log"]; ok {

			//This is a type assertion, which checks whether the value of vlog is of type []interface{}. If the assertion is true, the value of vlog is assigned to the variable logs.
			if logs, ok := vlog.([]interface{}); ok {

				// The for loop is used to loop through an array of logs. It is used to iterate over each element of the array to
				// perform a certain action. In this case, it is likely used to access each log and print it out or perform some other action with it.
				for i := 0; i < len(logs); i++ {

					// The if statement is used to check for the presence of the "data" key in a map[string]interface{} stored in the
					// logs[i] element. If the "data" key is present, the data associated with that key is assigned to the data
					// variable. The ok variable is a boolean indicating whether the "data" key was found.
					if data, ok := logs[i].(map[string]interface{})["data"]; ok {

						// This code is decoding a string of hexadecimal characters into a byte array. The data variable is an interface{}
						// type which is a generic data type that can hold any value. The hex.DecodeString() function decodes the string of
						// hexadecimal characters into a byte array and returns an error if something goes wrong. If there is an error, the
						// code returns the log and a nil value. Otherwise, the code will continue.
						data, err := hex.DecodeString(data.(string))
						if err != nil {
							return log, nil
						}

						// The purpose of this statement is to assign the value of the variable "data" to the "log.Data" property. This
						// statement is used to store the value of the "data" variable in the "log.Data" property. This statement is
						// typically used in programming to store data in a particular property or variable.
						log.Data = data
					}

					// This code is used to check if the key "topics" exists in the map logs[i] and, if it does exist, assign the value
					// of the key to the "Topics" field of the "log" struct. The "ok" keyword is used to check if the type assertion was successful.
					if topics, ok := logs[i].(map[string]interface{})["topics"]; ok {
						log.Topics = topics.([]interface{})
					}
				}
			}
		}
	}

	return log, nil
}

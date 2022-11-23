package service

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/blockchain"
	"github.com/cryptogateway/backend-envoys/assets/common/address"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// depositEthereum - ethereum transfer to deposit.
func (e *ExchangeService) depositEthereum(chain *proto.Chain) {
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()
	defer e.done(chain.GetId())

	client := blockchain.Dial(chain.Rpc, blockchain.RpcEthereum)
	blockBy, err := client.BlockByNumber(chain.GetBlock())
	if err != nil {
		return
	}

	for _, tx := range blockBy.Transactions {
		var (
			item proto.Transaction
		)

		switch tx.Type {
		case blockchain.TypeInternal:

			quantity := new(big.Int)
			quantity.SetString(strings.TrimPrefix(tx.Value, "0x"), 16)

			if value := decimal.ValueFloat(quantity, 18); value > 0 {

				item.To = address.New(tx.To).Hex()

				if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2", item.GetTo(), chain.GetPlatform()).Scan(&item.UserId); item.GetUserId() > 0 {

					item.Symbol = chain.GetParentSymbol()
					item.ChainId = chain.GetId()
					item.Platform = chain.GetPlatform()
					item.FinType = proto.FinType_CRYPTO
					item.TxType = proto.TxType_DEPOSIT
					item.Value = value
					item.Hash = tx.Hash
					item.Block = chain.GetBlock()
				}
			}

			break
		case blockchain.TypeContract:

			var (
				contract proto.Contract
			)

			logs, err := client.LogByTx(tx.Hash)
			if err != nil {
				return
			}

			if logs.Data != nil {

				if err := e.Context.Db.QueryRow("select symbol, protocol, decimals from contracts where lower(address) = $1", address.New(tx.To).Hex()).Scan(&contract.Symbol, &contract.Protocol, &contract.Decimals); err != nil { // No debug....
					continue
				}

				if len(logs.Topics) == 3 {

					// Uploading logs by method - (Transfer) or (Approval).
					transfer := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

					// Checking the method itself.
					switch logs.Topics[0] {
					case transfer.Hex():

						// Unpacking token abi data.
						tokenAbi, err := abi.JSON(strings.NewReader(blockchain.MainMetaData.ABI))
						if e.Context.Debug(err) {
							continue
						}

						// Unpacking all transaction data method - (Transfer): Data.
						instance, err := tokenAbi.Unpack("Transfer", logs.Data)
						if e.Context.Debug(err) {
							continue
						}

						// The number of tokens that were sent to the final recipient.
						if number, ok := instance[0].(*big.Int); ok {

							if value := decimal.ValueFloat(number, contract.GetDecimals()); value > 0 {

								item.To = address.New(logs.Topics[2].(string)).Hex()

								if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2 and protocol = $3", item.GetTo(), chain.GetPlatform(), contract.GetProtocol()).Scan(&item.UserId); item.GetUserId() > 0 {

									item.Symbol = contract.GetSymbol()
									item.Protocol = contract.GetProtocol()
									item.ChainId = chain.GetId()
									item.Platform = chain.GetPlatform()
									item.FinType = proto.FinType_CRYPTO
									item.TxType = proto.TxType_DEPOSIT
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

		if item.GetValue() > 0 {

			transaction, err := e.setTransaction(&item)
			if e.Context.Debug(err) {
				return
			}

			if err := e.replayPusher(proto.Pusher_DepositPublic, transaction); e.Context.Debug(err) {
				return
			}

		}
	}

	// Update values block.
	if _, err := e.Context.Db.Exec("update chains set block = $1 where id = $2;", chain.GetBlock()+1, chain.GetId()); e.Context.Debug(err) {
		return
	}

	e.block[chain.GetId()] = chain.GetBlock()
}

// depositTron - tron transfer to deposit.
func (e *ExchangeService) depositTron(chain *proto.Chain) {
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()
	defer e.done(chain.GetId())

	client := blockchain.Dial(chain.Rpc, blockchain.RpcTron)
	blockBy, err := client.BlockByNumber(chain.GetBlock())
	if err != nil {
		return
	}

	for _, tx := range blockBy.Transactions {

		var (
			item proto.Transaction
		)

		switch tx.Type {
		case blockchain.TypeInternal: // TRX parse transfer coin.

			value, err := strconv.ParseFloat(tx.Value, 64)
			if e.Context.Debug(err) {
				return
			}

			if value > 0 {

				item.To = address.New(tx.To).Base58()

				if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2", item.GetTo(), chain.GetPlatform()).Scan(&item.UserId); item.GetUserId() > 0 {

					item.Symbol = chain.GetParentSymbol()
					item.ChainId = chain.GetId()
					item.Platform = chain.GetPlatform()
					item.FinType = proto.FinType_CRYPTO
					item.TxType = proto.TxType_DEPOSIT
					item.Value = decimal.FromFloat(value).Div(decimal.FromFloat(1000000)).Float64() // value / 1000000
					item.Hash = tx.Hash
					item.Block = chain.GetBlock()
				}
			}

			break
		case blockchain.TypeContract: // Smart contract trigger transfer token.

			var (
				contract proto.Contract
			)

			logs, err := client.LogByTx(tx.Hash)
			if err != nil {
				return
			}

			if logs.Data != nil {

				if err := e.Context.Db.QueryRow("select symbol, protocol, decimals from contracts where address = $1", address.New(tx.To).Base58()).Scan(&contract.Symbol, &contract.Protocol, &contract.Decimals); err != nil { // No debug....
					continue
				}

				if len(logs.Topics) == 3 {

					// Uploading logs by method - (Transfer) or (Approval).
					transfers := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

					if logs.Topics[0] == strings.TrimPrefix(transfers.Hex(), "0x") {

						// Unpacking token abi data.
						tokenAbi, err := abi.JSON(strings.NewReader(blockchain.MainMetaData.ABI))
						if e.Context.Debug(err) {
							continue
						}

						// Unpacking all transaction data method - (Transfer): Data.
						instance, err := tokenAbi.Unpack("Transfer", logs.Data)
						if e.Context.Debug(err) {
							continue
						}

						// The number of tokens that were sent to the final recipient.
						if number, ok := instance[0].(*big.Int); ok || number.Int64() > 0 {

							if value := decimal.ValueFloat(number, contract.GetDecimals()); value > 0 {

								item.To = address.New(logs.Topics[2].(string)).Base58()

								if _ = e.Context.Db.QueryRow("select user_id from wallets where address = $1 and platform = $2 and protocol = $3", item.GetTo(), chain.GetPlatform(), contract.GetProtocol()).Scan(&item.UserId); item.GetUserId() > 0 {

									item.Symbol = contract.GetSymbol()
									item.Protocol = contract.GetProtocol()
									item.ChainId = chain.GetId()
									item.Platform = chain.GetPlatform()
									item.FinType = proto.FinType_CRYPTO
									item.TxType = proto.TxType_DEPOSIT
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

		if item.GetValue() > 0 {

			transaction, err := e.setTransaction(&item)
			if e.Context.Debug(err) {
				return
			}

			if err := e.replayPusher(proto.Pusher_DepositPublic, transaction); e.Context.Debug(err) {
				return
			}

		}

	}

	// Update values block.
	if _, err := e.Context.Db.Exec("update chains set block = $1 where id = $2;", chain.GetBlock()+1, chain.GetId()); e.Context.Debug(err) {
		return
	}

	e.block[chain.GetId()] = chain.GetBlock()
}

// depositBitcoin - bitcoin transfer to deposit.
func (e *ExchangeService) depositBitcoin(userId, txId int64, symbol string, to common.Address, value, price float64, platform proto.Platform, protocol proto.Protocol, chain *proto.Chain, subscribe bool) {
}

// transferEthereum - ethereum transfer to withdraw.
func (e *ExchangeService) transferEthereum(userId, txId int64, symbol string, to common.Address, value, price float64, platform proto.Platform, protocol proto.Protocol, chain *proto.Chain, subscribe bool) {
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	var (
		cross   keypair.CrossChain
		migrate = Query{
			Context: e.Context,
		}
		fees, convert, commission float64
		data                      []byte
		wei                       *big.Int
		transfer                  *types.Transaction
	)

	entropy, err := e.getEntropy(userId)
	if e.Context.Debug(err) {
		return
	}

	owner, private, err := cross.New(fmt.Sprintf("%v-&*39~763@)", e.Context.Secrets[1]), entropy, platform)
	if e.Context.Debug(err) {
		return
	}

	client, err := ethclient.Dial(chain.Rpc)
	if e.Context.Debug(err) {
		return
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(private, "0x"))
	if e.Context.Debug(err) {
		return
	}

	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(owner))
	if e.Context.Debug(err) {
		return
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if e.Context.Debug(err) {
		return
	}

	if protocol == proto.Protocol_MAINNET {

		// Find out the network commission for the transaction, this is necessary so that in cases of successful withdrawal,
		// at the end, correct the transaction data, since the standard network commission on exchanges is twice as high.
		impost := new(big.Float)
		impost.SetString(strconv.FormatInt(gasPrice.Int64()*21000, 10))
		fees, _ = new(big.Float).Quo(impost, big.NewFloat(math.Pow10(18))).Float64()

		if subscribe {
			wei = decimal.ValueInt(decimal.FromFloat(value).Sub(decimal.FromFloat(fees)).Float64(), 18)
		} else {
			wei = decimal.ValueInt(value, 18)
		}

		transfer, err = types.SignNewTx(privateKey, types.NewEIP155Signer(big.NewInt(chain.GetNetwork())), &types.LegacyTx{
			Nonce:    nonce,
			To:       &to,
			Value:    wei,
			Gas:      uint64(21000),
			GasPrice: gasPrice,
		})
		if e.Context.Debug(err) {
			return
		}

	} else {

		contract, err := e.getContract(symbol, chain.GetId())
		if e.Context.Debug(err) {
			return
		}

		// Find out the network commission for the transaction, this is necessary so that in cases of successful withdrawal,
		// at the end, correct the transaction data, since the standard network commission on exchanges is twice as high.
		impost := new(big.Float)
		impost.SetString(strconv.FormatInt(gasPrice.Int64()*60000, 10))
		commission, _ = new(big.Float).Quo(impost, big.NewFloat(math.Pow10(int(contract.GetDecimals())))).Float64()

		convert = decimal.FromFloat(commission).Mul(decimal.FromFloat(price)).Float64()
		wei = decimal.ValueInt(decimal.FromFloat(value).Sub(decimal.FromFloat(convert)).Float64(), contract.GetDecimals())

		paddedAmount, paddedAddress, contractAddress := common.LeftPadBytes(wei.Bytes(), 32), common.LeftPadBytes(to.Bytes(), 32), common.HexToAddress(contract.GetAddress())

		data = append(data, help.SignatureKeccak256([]byte("transfer(address,uint256)"))...)
		data = append(data, paddedAddress...)
		data = append(data, paddedAmount...)

		estimated, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			From: common.HexToAddress(owner),
			To:   &contractAddress,
			Data: data,
		})
		if e.Context.Debug(err) {
			return
		}

		impost = new(big.Float)
		impost.SetString(strconv.FormatInt(gasPrice.Int64()*int64(estimated), 10))
		fees, _ = new(big.Float).Quo(impost, big.NewFloat(math.Pow10(18))).Float64() // Eth fees 18 decimal.

		transfer, err = types.SignNewTx(privateKey, types.NewEIP155Signer(big.NewInt(chain.GetNetwork())), &types.LegacyTx{
			Nonce:    nonce,
			To:       &contractAddress,
			Value:    big.NewInt(0),
			Gas:      uint64(60000),
			GasPrice: gasPrice,
			Data:     data,
		})
		if e.Context.Debug(err) {
			return
		}

	}

	if err := client.SendTransaction(context.Background(), transfer); e.Context.Debug(err) {
		return
	}

	if subscribe {

		var (
			charges float64
		)

		if protocol == proto.Protocol_MAINNET {

			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), value, platform, proto.Protocol_MAINNET, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			charges = fees

		} else {

			// Parent account - update the reserve account, since for the transfer of the token, payment for gas is withdrawn from the parent account, for example, eth.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), fees, platform, proto.Protocol_MAINNET, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Contract account - update the reserve account, the amount that was deposited for the withdrawal of the token is converted and debited in a partial amount, excluding commission, for example:
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); This amount is 985.3 tst and will be overwritten without commission.
			if err := e.setReserve(userId, owner, symbol, decimal.FromFloat(value).Sub(decimal.FromFloat(convert)).Float64(), platform, protocol, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Token commission - update the collection account, the commission that was deducted from the amount of the token is credited to the exchange, the commission is calculated according to the formula and commission of the parent account,
			// for example: since we make commissions for the transfer exclusively from the parent account, we need to minus the commission of 0.006 eth from the amount of 1000 tst using conversion at the price of the token,
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); this amount is 985.3 tst and will be credited for the transfer,
			// and the amount of 14.7 tst will be credited to the stock exchange, since the reverse was charged at the expense of the stock exchange to pay for gas.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges + $2 where symbol = $1;", symbol, convert); e.Context.Debug(err) {
				return
			}

			fees = commission
			charges = decimal.FromFloat(fees).Mul(decimal.FromFloat(price)).Float64()
		}

		if _, err := e.Context.Db.Exec("update transactions set fees = $4, hash = $3, status = $2 where id = $1;", txId, proto.Status_FILLED, transfer.Hash().String(), fees); e.Context.Debug(err) {
			return
		}

		if err := e.replayPusher(proto.Pusher_WithdrawPublic, &proto.Transaction{
			Id:   txId,
			Fees: charges,
			Hash: transfer.Hash().String(),
		}); e.Context.Debug(err) {
			return
		}

		go migrate.SamplePosts(userId, "withdrawal", value, symbol)

	} else {

		if protocol == proto.Protocol_MAINNET {

			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), decimal.FromFloat(value).Add(decimal.FromFloat(fees)).Float64(), platform, proto.Protocol_MAINNET, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// The fee for the exchange net is double the value for the transfer gas, and the amount for the token withdrawal fee.
			// When replenishing the transfer wallet to pay the commission, a double commission is withdrawn.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges - $2, fees_costs = fees_costs + $2 where symbol = $1;", chain.GetParentSymbol(), decimal.FromFloat(value).Add(decimal.FromFloat(fees)).Float64()); e.Context.Debug(err) {
				return
			}

		}
	}

	if err := e.setReserveUnlock(userId, symbol, platform, protocol); e.Context.Debug(err) {
		return
	}
}

// transferTron - tron transfer to withdraw.
func (e *ExchangeService) transferTron(userId, txId int64, symbol, to string, value, price float64, platform proto.Platform, protocol proto.Protocol, chain *proto.Chain, subscribe bool) {
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	var (
		cross   keypair.CrossChain
		migrate = Query{
			Context: e.Context,
		}
		fees, convert float64
		transfer      *blockchain.Transfer
		wei           *big.Int
	)

	client := blockchain.Dial(chain.Rpc, blockchain.RpcTron)

	entropy, err := e.getEntropy(userId)
	if e.Context.Debug(err) {
		return
	}

	owner, private, err := cross.New(fmt.Sprintf("%v-&*39~763@)", e.Context.Secrets[1]), entropy, platform)
	if e.Context.Debug(err) {
		return
	}
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(private, "0x"))
	if err != nil {
		return
	}
	client.Private(privateKey)

	if protocol == proto.Protocol_MAINNET {

		transfer = &blockchain.Transfer{
			To:    address.New(to).Hex(true),
			Gas:   10000000,
			Value: decimal.ValueInt(value, 6),
		}

		estimate, err := client.EstimateGas(transfer)
		if e.Context.Debug(err) {
			return
		}
		fees = decimal.FromInt(estimate).Div(decimal.FromInt(1000000)).Float64()

		if subscribe {
			wei = decimal.ValueInt(decimal.FromFloat(value).Sub(decimal.FromFloat(fees)).Float64(), 6)
		} else {
			wei = decimal.ValueInt(value, 6)
		}

		transfer.Value = wei

	} else {

		contract, err := e.getContract(symbol, chain.GetId())
		if e.Context.Debug(err) {
			return
		}

		data, err := client.Data(address.New(to).Hex(true), decimal.ValueInt(value, contract.GetDecimals()).Bytes())
		if err != nil {
			return
		}

		transfer = &blockchain.Transfer{
			Contract: address.New(contract.GetAddress()).Hex(true),
			Gas:      10000000,
			Data:     data,
		}

		estimate, err := client.EstimateGas(transfer)
		if err != nil {
			panic(err)
		}
		fees = decimal.FromInt(estimate).Div(decimal.FromInt(1000000)).Float64()

		convert = decimal.FromFloat(fees).Mul(decimal.FromFloat(price)).Float64()
		wei = decimal.ValueInt(decimal.FromFloat(value).Sub(decimal.FromFloat(convert)).Float64(), contract.GetDecimals())
	}

	hash, err := client.Transfer(transfer)
	if e.Context.Debug(err) {
		return
	}

	if err = client.Transaction(); e.Context.Debug(err) {
		return
	}

	if subscribe {

		var (
			charges float64
		)

		if protocol == proto.Protocol_MAINNET {

			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), value, platform, proto.Protocol_MAINNET, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			charges = fees
		} else {

			// Parent account - update the reserve account, since for the transfer of the token, payment for gas is withdrawn from the parent account, for example, eth.
			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), fees, platform, proto.Protocol_MAINNET, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Contract account - update the reserve account, the amount that was deposited for the withdrawal of the token is converted and debited in a partial amount, excluding commission, for example:
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); This amount is 985.3 tst and will be overwritten without commission.
			if err := e.setReserve(userId, owner, symbol, decimal.FromFloat(value).Sub(decimal.FromFloat(convert)).Float64(), platform, protocol, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// Token commission - update the collection account, the commission that was deducted from the amount of the token is credited to the exchange, the commission is calculated according to the formula and commission of the parent account,
			// for example: since we make commissions for the transfer exclusively from the parent account, we need to minus the commission of 0.006 eth from the amount of 1000 tst using conversion at the price of the token,
			// (fee: 0.006 eth) * (price: 2450 tst) = 14.7 tst; (value: 1000 - fees: 14.7 tst = 985.3 tst); this amount is 985.3 tst and will be credited for the transfer,
			// and the amount of 14.7 tst will be credited to the stock exchange, since the reverse was charged at the expense of the stock exchange to pay for gas.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges + $2 where symbol = $1;", symbol, convert); e.Context.Debug(err) {
				return
			}

			charges = decimal.FromFloat(fees).Mul(decimal.FromFloat(price)).Float64()
		}

		if _, err := e.Context.Db.Exec("update transactions set fees = $4, hash = $3, status = $2 where id = $1;", txId, proto.Status_FILLED, hash, fees); e.Context.Debug(err) {
			return
		}

		if err := e.replayPusher(proto.Pusher_WithdrawPublic, &proto.Transaction{
			Id:   txId,
			Fees: charges,
			Hash: hash,
		}); e.Context.Debug(err) {
			return
		}

		go migrate.SamplePosts(userId, "withdrawal", value, symbol)

	} else {

		if protocol == proto.Protocol_MAINNET {

			if err := e.setReserve(userId, owner, chain.GetParentSymbol(), decimal.FromFloat(value).Add(decimal.FromFloat(fees)).Float64(), platform, proto.Protocol_MAINNET, proto.Balance_MINUS); e.Context.Debug(err) {
				return
			}

			// The fee for the exchange net is double the value for the transfer gas, and the amount for the token withdrawal fee.
			// When replenishing the transfer wallet to pay the commission, a double commission is withdrawn.
			if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges - $2, fees_costs = fees_costs + $2 where symbol = $1;", chain.GetParentSymbol(), decimal.FromFloat(value).Add(decimal.FromFloat(fees)).Float64()); e.Context.Debug(err) {
				return
			}

		}
	}

	if err := e.setReserveUnlock(userId, symbol, platform, protocol); e.Context.Debug(err) {
		return
	}
}

// transferBitcoin - bitcoin transfer to withdraw.
func (e *ExchangeService) transferBitcoin(chain *proto.Chain) {}

package spot

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/blockchain"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/marketplace"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/query"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/status"
	"time"
)

// replayPriceScale - pair price scale.
func (e *Service) replayPriceScale() {

	for {

		func() {

			rows, err := e.Context.Db.Query(`select id, price, base_unit, quote_unit from spot_pairs where status = $1 order by id`, true)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					pair  pbspot.Pair
					price float64
				)

				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit); e.Context.Debug(err) {
					continue
				}

				migrate, err := e.GetGraph(context.Background(), &pbspot.GetRequestGraph{BaseUnit: pair.GetBaseUnit(), QuoteUnit: pair.GetQuoteUnit(), Limit: 2})
				if e.Context.Debug(err) {
					continue
				}

				if price = marketplace.Price().Unit(pair.GetBaseUnit(), pair.GetQuoteUnit()); price > 0 {

					if len(migrate.Fields) > 0 {
						price = (price + pair.GetPrice() + migrate.Fields[0].GetPrice()) / 3
					} else {
						price = (price + pair.GetPrice()) / 2
					}

					// Fix scale price.
					if (price - pair.GetPrice()) > 100 {
						price -= (price - pair.GetPrice()) - (price-pair.GetPrice())/8
					}

				}

				if price == 0 {
					if len(migrate.Fields) > 0 {
						price = (migrate.Fields[0].GetPrice() + pair.GetPrice()) / 2
					} else {
						price = pair.GetPrice()
					}
				}

				if _, err := e.Context.Db.Exec("update spot_pairs set price = $3 where base_unit = $1 and quote_unit = $2;", pair.GetBaseUnit(), pair.GetQuoteUnit(), price); e.Context.Debug(err) {
					continue
				}
			}
		}()
	}
}

// replayMarket - trade both/bot market.
func (e *Service) replayMarket() {

	// Loading at a specific time interval.
	ticker := time.NewTicker(time.Minute * e.Context.Intervals.Market)
	for range ticker.C {

		func() {

			rows, err := e.Context.Db.Query(`select id, price, base_unit, quote_unit from spot_pairs where status = $1 order by id`, true)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					pair pbspot.Pair
				)

				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit); e.Context.Debug(err) {
					continue
				}

				if _, err := e.Context.Db.Exec(`insert into spot_trades (assigning, base_unit, quote_unit, price, quantity, market) values ($1, $2, $3, $4, $5, $6)`, pbspot.Assigning_MARKET_PRICE, pair.GetBaseUnit(), pair.GetQuoteUnit(), pair.GetPrice(), 0, true); e.Context.Debug(err) {
					continue
				}

				for _, interval := range help.Depth() {

					migrate, err := e.GetGraph(context.Background(), &pbspot.GetRequestGraph{BaseUnit: pair.GetBaseUnit(), QuoteUnit: pair.GetQuoteUnit(), Limit: 2, Resolution: interval})
					if e.Context.Debug(err) {
						continue
					}

					if err := e.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/graph:%v", interval)); e.Context.Debug(err) {
						continue
					}
				}
			}
		}()
	}
}

// replayChainStatus - chain ping rpc status.
func (e *Service) replayChainStatus() {

	// Loading at a specific time interval.
	ticker := time.NewTicker(time.Minute * e.Context.Intervals.Chain)
	for range ticker.C {

		func() {

			rows, err := e.Context.Db.Query(`select id, rpc, status from spot_chains`)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					item pbspot.Chain
				)

				if err = rows.Scan(&item.Id, &item.Rpc, &item.Status); e.Context.Debug(err) {
					continue
				}

				if ok := help.Ping(item.Rpc); !ok {
					item.Status = false
				}

				if _, err := e.Context.Db.Exec("update spot_chains set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
					continue
				}
			}

		}()
	}
}

// replayTradeInit - init search for a deal by active orders.
func (e *Service) replayTradeInit(order *pbspot.Order, side pbspot.Side) {

	if err := e.Context.Publish(order, "exchange", "order/create"); e.Context.Debug(err) {
		return
	}

	rows, err := e.Context.Db.Query(`select id, assigning, base_unit, quote_unit, value, quantity, price, user_id, status from spot_orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and user_id != $4 and status = $5 and type = $6 order by id`, side, order.GetBaseUnit(), order.GetQuoteUnit(), order.GetUserId(), pbspot.Status_PENDING, pbspot.OrderType_SPOT)
	if e.Context.Debug(err) {
		return
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item pbspot.Order
		)

		if err = rows.Scan(&item.Id, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.Value, &item.Quantity, &item.Price, &item.UserId, &item.Status); err != nil {
			return
		}

		// Updating the amount of the current order.
		row, err := e.Context.Db.Query("select value from spot_orders where id = $1 and status = $2", order.GetId(), pbspot.Status_PENDING)
		if err != nil {
			return
		}

		if row.Next() {
			if err = row.Scan(&order.Value); err != nil {
				return
			}
		} else {
			order.Value = 0
		}
		row.Close()

		switch side {

		case pbspot.Side_BID: // Buy at BID price.

			if order.GetPrice() >= item.GetPrice() {
				e.Context.Logger.Infof("[BID]: (item [%v]) >= (order [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				e.replayTradeProcess(order, &item)

			} else {
				e.Context.Logger.Infof("[BID]: no matches found: (item [%v]) >= (order [%v])", order.GetPrice(), item.GetPrice())
			}

			break

		case pbspot.Side_ASK: // Sell at ASK price.

			if order.GetPrice() <= item.GetPrice() {
				e.Context.Logger.Infof("[ASK]: (order [%v]) <= (item [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				e.replayTradeProcess(order, &item)

			} else {
				e.Context.Logger.Infof("[ASK]: no matches found: (order [%v]) <= (item [%v])", order.GetPrice(), item.GetPrice())
			}

			break
		default:
			if err := e.Context.Debug(status.Error(11589, "invalid assigning trade position")); err {
				return
			}
		}

	}

	if err = rows.Err(); err != nil {
		return
	}
}

// replayTradeProcess - process trade. [params 0 - order] : [params 1 - item]
func (e *Service) replayTradeProcess(params ...*pbspot.Order) {

	var (
		value   float64
		migrate = query.Migrate{
			Context: e.Context,
		}
	)

	// If the amount of the current order is greater than or equal to the amount of the found order.
	if params[0].GetValue() >= params[1].GetValue() {

		// If the amount of the current order is greater than zero.
		if params[1].GetValue() > 0 {

			if err := e.Context.Db.QueryRow("update spot_orders set value = value - $2 where id = $1 and status = $3 returning value;", params[0].GetId(), params[1].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update spot_orders set status = $2 where id = $1;", params[0].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[0].GetUserId(), "order_filled", params[0].GetId(), e.getQuantity(params[0].GetAssigning(), params[0].GetQuantity(), params[0].GetPrice(), false), params[0].GetBaseUnit(), params[0].GetQuoteUnit(), params[0].GetAssigning())
			}

			if err := e.Context.Db.QueryRow("update spot_orders set value = value - $2 where id = $1 and status = $3 returning value;", params[1].GetId(), params[1].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update spot_orders set status = $2 where id = $1;", params[1].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[1].GetUserId(), "order_filled", params[1].GetId(), e.getQuantity(params[1].GetAssigning(), params[1].GetQuantity(), params[1].GetPrice(), false), params[1].GetBaseUnit(), params[1].GetQuoteUnit(), params[1].GetAssigning())
			}

			switch params[1].GetAssigning() {
			case pbspot.Assigning_BUY:

				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[1].GetValue()).Mul(decimal.FromFloat(params[1].GetPrice())).Float64(), false)
				if err := e.setBalance(params[0].GetQuoteUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: false, Equal: true}

				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[1].GetValue(), true)
				if err := e.setBalance(params[0].GetBaseUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: true, Equal: true}

				break
			case pbspot.Assigning_SELL:

				quantity, fees := e.getSum(params[0].GetBaseUnit(), params[1].GetValue(), false)
				if err := e.setBalance(params[0].GetBaseUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: true, Equal: true}

				quantity, fees = e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[1].GetValue()).Mul(decimal.FromFloat(params[0].GetPrice())).Float64(), true)
				if err := e.setBalance(params[0].GetQuoteUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: false, Equal: true}

				break
			}

			e.Context.Logger.Infof("[(A)%v]: (Assigning: %v), (Price: [%v]-[%v]), (Value: [%v]-[%v]), (UserID: [%v]-[%v])", params[0].GetAssigning(), params[1].GetAssigning(), params[0].GetPrice(), params[1].GetPrice(), params[0].GetValue(), params[1].GetValue(), params[0].GetUserId(), params[1].GetUserId())
		}

		// If the amount of the current order is less than the amount of the found order.
	} else if params[0].GetValue() < params[1].GetValue() {

		// If the amount of the found order is greater than zero.
		if params[0].GetValue() > 0 {

			if err := e.Context.Db.QueryRow("update spot_orders set value = value - $2 where id = $1 and status = $3 returning value;", params[0].GetId(), params[0].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update spot_orders set status = $2 where id = $1;", params[0].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[0].GetUserId(), "order_filled", params[0].GetId(), e.getQuantity(params[0].GetAssigning(), params[0].GetQuantity(), params[0].GetPrice(), false), params[0].GetBaseUnit(), params[0].GetQuoteUnit(), params[0].GetAssigning())
			}

			if err := e.Context.Db.QueryRow("update spot_orders set value = value - $2 where id = $1 and status = $3 returning value;", params[1].GetId(), params[0].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update spot_orders set status = $2 where id = $1;", params[1].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[1].GetUserId(), "order_filled", params[1].GetId(), e.getQuantity(params[1].GetAssigning(), params[1].GetQuantity(), params[1].GetPrice(), false), params[1].GetBaseUnit(), params[1].GetQuoteUnit(), params[1].GetAssigning())
			}

			switch params[1].GetAssigning() {
			case pbspot.Assigning_BUY:

				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[0].GetValue()).Mul(decimal.FromFloat(params[1].GetPrice())).Float64(), false)
				if err := e.setBalance(params[0].GetQuoteUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: false, Equal: false}

				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[0].GetValue(), true)
				if err := e.setBalance(params[0].GetBaseUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: true, Equal: false}

				break
			case pbspot.Assigning_SELL:

				quantity, fees := e.getSum(params[0].GetBaseUnit(), params[0].GetValue(), false)
				if err := e.setBalance(params[0].GetBaseUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: true, Equal: false}

				quantity, fees = e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[0].GetValue()).Mul(decimal.FromFloat(params[0].GetPrice())).Float64(), true)
				if err := e.setBalance(params[0].GetQuoteUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: false, Equal: false}

				break
			}

			e.Context.Logger.Infof("[(B)%v]: (Assigning: %v), (Price: [%v]-[%v]), (Value: [%v]-[%v]), (UserID: [%v]-[%v])", params[0].GetAssigning(), params[1].GetAssigning(), params[0].GetPrice(), params[1].GetPrice(), params[0].GetValue(), params[1].GetValue(), params[0].GetUserId(), params[1].GetUserId())
		}

	}

	if err := e.setTrade(params[0], params[1]); err != nil {
		return
	}
}

// replayDeposit - replenishment of assets, work with blockchain structures.
func (e *Service) replayDeposit() {
	e.run, e.wait, e.block = make(map[int64]bool), make(map[int64]bool), make(map[int64]int64)

	for {

		func() {

			var (
				chain pbspot.Chain
			)

			rows, err := e.Context.Db.Query("select id, rpc, rpc_key, platform, block, network, confirmation, parent_symbol from spot_chains where status = $1", true)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				if err := rows.Scan(&chain.Id, &chain.Rpc, &chain.RpcKey, &chain.Platform, &chain.Block, &chain.Network, &chain.Confirmation, &chain.ParentSymbol); e.Context.Debug(err) {
					continue
				}

				if chain.GetBlock() == 0 {
					chain.Block = 1
				}

				if block, ok := e.block[chain.GetId()]; !ok && block == chain.GetBlock() {
					continue
				}

				if e.run[chain.GetId()] {
					if _, ok := e.wait[chain.GetId()]; !ok {
						continue
					}

					e.wait[chain.GetId()] = false
				} else {
					e.run[chain.GetId()] = true
				}

				switch chain.GetPlatform() {
				case pbspot.Platform_ETHEREUM:
					go e.depositEthereum(&chain)
					break
				case pbspot.Platform_TRON:
					go e.depositTron(&chain)
					break
				}

				time.Sleep(500 * time.Millisecond)
			}

			// Confirmation deposits assets.
			e.replayConfirmation()

		}()

	}

}

// replayDeposit - withdrawal of assets, work with blockchain structures.
func (e *Service) replayWithdraw() {
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	for {

		func() {

			rows, err := e.Context.Db.Query(`select id, symbol, "to", chain_id, fees, value, price, platform, protocol, claim from spot_transactions where status = $1 and tx_type = $2 and fin_type = $3`, pbspot.Status_PENDING, pbspot.TxType_WITHDRAWS, pbspot.FinType_CRYPTO)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					item, reserve, insure pbspot.Transaction
				)

				if err := rows.Scan(&item.Id, &item.Symbol, &item.To, &item.ChainId, &item.Fees, &item.Value, &item.Price, &item.Platform, &item.Protocol, &item.Claim); e.Context.Debug(err) {
					continue
				}

				chain, err := e.getChain(item.GetChainId(), true)
				if e.Context.Debug(err) {
					continue
				}

				if item.GetProtocol() == pbspot.Protocol_MAINNET {

					// Find the reserve asset from which funds will be transferred,
					// by its platform, as well as by protocol, symbol, and number of funds.
					if _ = e.Context.Db.QueryRow("select value, user_id from spot_reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), false).Scan(&reserve.Value, &reserve.UserId); reserve.GetValue() > 0 {

						if _, err := e.Context.Db.Exec("update spot_transactions set status = $2 where id = $1;", item.GetId(), pbspot.Status_PROCESSING); e.Context.Debug(err) {
							continue
						}

						if err := e.setReserveLock(reserve.GetUserId(), item.GetSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
							continue
						}

						switch item.GetPlatform() {
						case pbspot.Platform_ETHEREUM:
							e.transferEthereum(reserve.GetUserId(), item.GetId(), item.GetSymbol(), common.HexToAddress(item.GetTo()), item.GetValue(), 0, item.GetPlatform(), item.GetProtocol(), chain, true)
							break
						case pbspot.Platform_TRON:
							e.transferTron(reserve.GetUserId(), item.GetId(), item.GetSymbol(), item.GetTo(), item.GetValue(), 0, item.GetPlatform(), item.GetProtocol(), chain, true)
							break
						}

					}

				} else {

					// Find the reserve asset from which funds will be transferred,
					// by its platform, as well as by protocol, symbol, and number of funds.
					if _ = e.Context.Db.QueryRow("select a.value, a.user_id from spot_reserves a inner join spot_reserves b on case when a.protocol > 0 then b.user_id = a.user_id and b.symbol = $6 and b.platform = a.platform and b.protocol = 0 and b.value >= $5 and b.lock = $7 end where a.symbol = $1 and a.value >= $2 and a.platform = $3 and a.protocol = $4 and a.lock = $7", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), item.GetFees(), chain.GetParentSymbol(), false).Scan(&reserve.Value, &reserve.UserId); reserve.GetValue() > 0 {

						if _, err := e.Context.Db.Exec("update spot_transactions set status = $2 where id = $1;", item.GetId(), pbspot.Status_PROCESSING); e.Context.Debug(err) {
							continue
						}

						if err := e.setReserveLock(reserve.GetUserId(), item.GetSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
							continue
						}

						switch item.GetPlatform() {
						case pbspot.Platform_ETHEREUM:
							e.transferEthereum(reserve.GetUserId(), item.GetId(), item.GetSymbol(), common.HexToAddress(item.GetTo()), item.GetValue(), item.GetPrice(), item.GetPlatform(), item.GetProtocol(), chain, true)
							break
						case pbspot.Platform_TRON:
							e.transferTron(reserve.GetUserId(), item.GetId(), item.GetSymbol(), item.GetTo(), item.GetValue(), item.GetPrice(), item.GetPlatform(), item.GetProtocol(), chain, true)
							break
						}

					} else {

						if !item.GetClaim() {

							currency, err := e.getCurrency(chain.GetParentSymbol(), false)
							if err != nil {
								return
							}

							if currency.GetFeesCharges() >= item.GetFees()*2 {

								// Find the reserve asset from which funds will be transferred,
								// by its platform, as well as by protocol, symbol, and number of funds.
								if _ = e.Context.Db.QueryRow("select value, address from spot_reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), false).Scan(&reserve.Value, &reserve.To); reserve.GetValue() > 0 {

									if _ = e.Context.Db.QueryRow("select value, user_id from spot_reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", chain.GetParentSymbol(), item.GetFees()*2, item.GetPlatform(), pbspot.Protocol_MAINNET, false).Scan(&insure.Value, &insure.UserId); insure.GetValue() > 0 {

										if err := e.setReserveLock(insure.GetUserId(), chain.GetParentSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
											continue
										}

										switch item.GetPlatform() {
										case pbspot.Platform_ETHEREUM:
											e.transferEthereum(insure.GetUserId(), item.GetId(), chain.GetParentSymbol(), common.HexToAddress(reserve.GetTo()), item.GetFees(), 0, item.GetPlatform(), pbspot.Protocol_MAINNET, chain, false)
											break
										case pbspot.Platform_TRON:
											e.transferTron(insure.GetUserId(), item.GetId(), chain.GetParentSymbol(), reserve.GetTo(), item.GetFees(), 0, item.GetPlatform(), pbspot.Protocol_MAINNET, chain, false)
											break
										}

										if _, err := e.Context.Db.Exec("update spot_transactions set claim = $2 where id = $1;", item.GetId(), true); e.Context.Debug(err) {
											continue
										}
									}
								}

							} else {
								e.Context.Logger.Info("[REVERSE]: there are no fees to pay the reverse fee")
							}
						}
					}
				}
			}
		}()

		time.Sleep(10 * time.Second)
	}

}

// replayConfirmation - confirmation of transactions by networks.
func (e *Service) replayConfirmation() {

	rows, err := e.Context.Db.Query(`select id, hash, symbol, "to", fees, chain_id, user_id, value, confirmation, block, platform, protocol, create_at from spot_transactions where status = $1 and tx_type = $2`, pbspot.Status_PENDING, pbspot.TxType_DEPOSIT)
	if e.Context.Debug(err) {
		return
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item pbspot.Transaction
		)

		if err := rows.Scan(&item.Id, &item.Hash, &item.Symbol, &item.To, &item.Fees, &item.ChainId, &item.UserId, &item.Value, &item.Confirmation, &item.Block, &item.Platform, &item.Protocol, &item.CreateAt); e.Context.Debug(err) {
			continue
		}

		chain, err := e.getChain(item.GetChainId(), true)
		if e.Context.Debug(err) {
			return
		}

		if blockchain.Dial(chain.GetRpc(), chain.GetPlatform()).Status(item.Hash) {

			if (chain.GetBlock()-item.GetBlock()) >= chain.GetConfirmation() && item.GetConfirmation() >= chain.GetConfirmation() {

				currency, err := e.getCurrency(item.GetSymbol(), false)
				if e.Context.Debug(err) {
					return
				}

				if item.GetValue() >= currency.GetMinDeposit() {

					// Crediting a new deposit to the local wallet address.
					if _, err := e.Context.Db.Exec("update spot_assets set balance = balance + $3 where symbol = $2 and user_id = $1;", item.GetUserId(), item.GetSymbol(), item.GetValue()); e.Context.Debug(err) {
						continue
					}

					item.Hook = true
					item.Status = pbspot.Status_FILLED

					if err := e.Context.Publish(&item, "exchange", "deposit/open", "deposit/status"); e.Context.Debug(err) {
						return
					}

				} else {
					item.Status = pbspot.Status_RESERVE
				}

				if err := e.setReserve(item.GetUserId(), item.GetTo(), item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), pbspot.Balance_PLUS); e.Context.Debug(err) {
					continue
				}

				// Update deposits pending status to success status.
				if _, err := e.Context.Db.Exec("update spot_transactions set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
					continue
				}

			} else {
				if _, err := e.Context.Db.Exec("update spot_transactions set confirmation = $2 where id = $1;", item.GetId(), chain.GetBlock()-item.GetBlock()); e.Context.Debug(err) {
					continue
				}
			}

		} else {

			item.Hook = true
			item.Status = pbspot.Status_FAILED

			if _, err := e.Context.Db.Exec("update spot_transactions set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
				continue
			}

			if err := e.Context.Publish(&item, "exchange", "deposit/open", "deposit/status"); e.Context.Debug(err) {
				return
			}
		}
	}
}

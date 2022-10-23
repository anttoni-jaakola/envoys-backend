package service

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/marketplace"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/status"
	"time"
)

// replayPriceScale - pair price scale.
func (e *ExchangeService) replayPriceScale() {

	for {

		func() {

			rows, err := e.Context.Db.Query(`select id, price, base_unit, quote_unit from pairs where status = $1 order by id`, true)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					pair  proto.Pair
					price float64
				)

				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit); e.Context.Debug(err) {
					continue
				}

				migrate, err := e.GetGraph(context.Background(), &proto.GetExchangeRequestGraph{BaseUnit: pair.GetBaseUnit(), QuoteUnit: pair.GetQuoteUnit(), Limit: 2})
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

				if _, err := e.Context.Db.Exec("update pairs set price = $3 where base_unit = $1 and quote_unit = $2;", pair.GetBaseUnit(), pair.GetQuoteUnit(), price); e.Context.Debug(err) {
					continue
				}
			}
		}()
	}
}

// replayMarket - trade both/bot market.
func (e *ExchangeService) replayMarket() {

	// Loading at a specific time interval.
	ticker := time.NewTicker(time.Minute * e.Context.IntervalMarket)
	for range ticker.C {

		func() {

			rows, err := e.Context.Db.Query(`select id, price, base_unit, quote_unit from pairs where status = $1 order by id`, true)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					pair proto.Pair
				)

				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit); e.Context.Debug(err) {
					continue
				}

				if _, err := e.Context.Db.Exec(`insert into trades (assigning, base_unit, quote_unit, price, quantity, market) values ($1, $2, $3, $4, $5, $6)`, proto.Assigning_MARKET_PRICE, pair.GetBaseUnit(), pair.GetQuoteUnit(), pair.GetPrice(), 0, true); e.Context.Debug(err) {
					continue
				}

				if err := e.replayPusher(proto.Pusher_TradePublic, pair.GetBaseUnit(), pair.GetQuoteUnit()); e.Context.Debug(err) {
					continue
				}
			}
		}()
	}
}

// replayChainStatus - chain ping rpc status.
func (e *ExchangeService) replayChainStatus() {

	// Loading at a specific time interval.
	ticker := time.NewTicker(time.Minute * e.Context.IntervalChainStatus)
	for range ticker.C {

		func() {

			rows, err := e.Context.Db.Query(`select id, rpc, status from chains`)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					item proto.Chain
				)

				if err = rows.Scan(&item.Id, &item.Rpc, &item.Status); e.Context.Debug(err) {
					continue
				}

				if ok := help.Ping(item.Rpc); !ok {
					item.Status = false
				}

				if _, err := e.Context.Db.Exec("update chains set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
					continue
				}
			}

		}()
	}
}

// replayTradeInit - init search for a deal by active orders.
func (e *ExchangeService) replayTradeInit(order *proto.Order, spread proto.Spread) {

	if err := e.replayPusher(proto.Pusher_OrderCreatePublic, order); e.Context.Debug(err) {
		return
	}

	rows, err := e.Context.Db.Query(`select id, assigning, base_unit, quote_unit, value, quantity, price, user_id, status from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and user_id != $4 and status = $5 and type = $6 order by id`, spread, order.GetBaseUnit(), order.GetQuoteUnit(), order.GetUserId(), proto.Status_PENDING, proto.OrderType_SPOT)
	if e.Context.Debug(err) {
		return
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item proto.Order
		)

		if err = rows.Scan(&item.Id, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.Value, &item.Quantity, &item.Price, &item.UserId, &item.Status); err != nil {
			return
		}

		// Updating the amount of the current order.
		row, err := e.Context.Db.Query("select value from orders where id = $1 and status = $2", order.GetId(), proto.Status_PENDING)
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

		switch spread {

		case proto.Spread_BID: // Buy at BID price.

			if order.GetPrice() >= item.GetPrice() {
				e.Context.Logger.Infof("[BID]: (item [%v]) >= (order [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				e.replayTradeProcess(order, &item)

			} else {
				e.Context.Logger.Infof("[BID]: no matches found: (item [%v]) >= (order [%v])", order.GetPrice(), item.GetPrice())
			}

			break

		case proto.Spread_ASK: // Sell at ASK price.

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
func (e *ExchangeService) replayTradeProcess(params ...*proto.Order) {

	var (
		value   float64
		migrate = Query{
			Context: e.Context,
		}
	)

	// If the amount of the current order is greater than or equal to the amount of the found order.
	if params[0].GetValue() >= params[1].GetValue() {

		// Комиссия (taker) обычно немного выше, чем комиссия (maker), чтобы стимулировать маркет-maker,
		// (taker) удаляют ликвидность, «забирая» доступные ордера, которые немедленно исполняются.
		// Это выставленный ордер, который выполняется моментально ответным ордером в стакане. То есть Taker расходует ликвидность, забирая ордер из списка. Если ордер выставлен по рынку, либо вы используете опцию Быстрого обмена, такая сделка всегда является Taker-заявкой.

		// If the amount of the current order is greater than zero.
		if params[1].GetValue() > 0 {

			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[0].GetId(), params[1].GetValue(), proto.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[0].GetId(), proto.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[0].GetUserId(), "order_filled", params[0].GetId(), e.getQuantity(params[0].GetAssigning(), params[0].GetQuantity(), params[0].GetPrice(), false), params[0].GetBaseUnit(), params[0].GetQuoteUnit(), params[0].GetAssigning())
			}

			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[1].GetId(), params[1].GetValue(), proto.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[1].GetId(), proto.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[1].GetUserId(), "order_filled", params[1].GetId(), e.getQuantity(params[1].GetAssigning(), params[1].GetQuantity(), params[1].GetPrice(), false), params[1].GetBaseUnit(), params[1].GetQuoteUnit(), params[1].GetAssigning())
			}

			switch params[1].GetAssigning() {

			// Если я params[0] покупаю по цене ордера params[1].
			case proto.Assigning_BUY:

				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[1].GetValue()).Mul(decimal.FromFloat(params[1].GetPrice())).Float64(), false)
				if err := e.setBalance(
					params[0].GetQuoteUnit(),
					params[0].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[0].GetId(),
					params[0].GetUserId(),
					params[0].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[1].GetPrice(),
					params[1].GetValue(),
					fees,
					false,
					false,
				); err != nil {
					return
				}

				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[1].GetValue(), true)
				if err := e.setBalance(
					params[0].GetBaseUnit(),
					params[1].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[1].GetId(),
					params[1].GetUserId(),
					params[1].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[1].GetPrice(),
					params[1].GetValue(),
					fees,
					true,
					true,
				); err != nil {
					return
				}

				break

			// Если я params[0] продаю по выставленной цене своего ордера.
			case proto.Assigning_SELL:

				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[1].GetValue()).Mul(decimal.FromFloat(params[0].GetPrice())).Float64(), true)
				if err := e.setBalance(
					params[0].GetQuoteUnit(),
					params[1].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[1].GetId(),
					params[1].GetUserId(),
					params[1].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[0].GetPrice(),
					params[1].GetValue(),
					fees,
					true,
					false,
				); err != nil {
					return
				}

				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[1].GetValue(), false)
				if err := e.setBalance(
					params[0].GetBaseUnit(),
					params[0].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[0].GetId(),
					params[0].GetUserId(),
					params[0].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[0].GetPrice(),
					params[1].GetValue(),
					fees,
					false,
					true,
				); err != nil {
					return
				}

				break
			}

			e.Context.Logger.Infof("[(A)%v]: (Assigning: %v), (Price: [%v]-[%v]), (Value: [%v]-[%v]), (UserID: [%v]-[%v])", params[0].GetAssigning(), params[1].GetAssigning(), params[0].GetPrice(), params[1].GetPrice(), params[0].GetValue(), params[1].GetValue(), params[0].GetUserId(), params[1].GetUserId())
		}

		// If the amount of the current order is less than the amount of the found order.
	} else if params[0].GetValue() < params[1].GetValue() {

		// (maker) «создают рынок» для других трейдеров и приносят ликвидность на биржу.
		// Это выставленный ордер, который выкупается не сразу, а остается в списке ордеров, чтобы быть исполненным в будущем.

		// If the amount of the found order is greater than zero.
		if params[0].GetValue() > 0 {

			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[0].GetId(), params[0].GetValue(), proto.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[0].GetId(), proto.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[0].GetUserId(), "order_filled", params[0].GetId(), e.getQuantity(params[0].GetAssigning(), params[0].GetQuantity(), params[0].GetPrice(), false), params[0].GetBaseUnit(), params[0].GetQuoteUnit(), params[0].GetAssigning())
			}

			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[1].GetId(), params[0].GetValue(), proto.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// Update order status with pending in to filled.
			if value == 0 {
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[1].GetId(), proto.Status_FILLED); err != nil {
					return
				}

				go migrate.SamplePosts(params[1].GetUserId(), "order_filled", params[1].GetId(), e.getQuantity(params[1].GetAssigning(), params[1].GetQuantity(), params[1].GetPrice(), false), params[1].GetBaseUnit(), params[1].GetQuoteUnit(), params[1].GetAssigning())
			}

			switch params[1].GetAssigning() {

			// Если я params[0] покупаю по цене ордера params[1].
			case proto.Assigning_BUY:

				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[0].GetValue()).Mul(decimal.FromFloat(params[1].GetPrice())).Float64(), false)
				if err := e.setBalance(
					params[0].GetQuoteUnit(),
					params[0].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[0].GetId(),
					params[0].GetUserId(),
					params[0].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[1].GetPrice(),
					params[0].GetValue(),
					fees,
					false,
					false,
				); err != nil {
					return
				}

				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[0].GetValue(), true)
				if err := e.setBalance(
					params[0].GetBaseUnit(),
					params[1].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[1].GetId(),
					params[1].GetUserId(),
					params[1].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[1].GetPrice(),
					params[0].GetValue(),
					fees,
					true,
					true,
				); err != nil {
					return
				}

				break

			// Если я params[0] продаю по выставленной цене своего ордера.
			case proto.Assigning_SELL:

				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.FromFloat(params[0].GetValue()).Mul(decimal.FromFloat(params[0].GetPrice())).Float64(), true)
				if err := e.setBalance(
					params[0].GetQuoteUnit(),
					params[1].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[1].GetId(),
					params[1].GetUserId(),
					params[1].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[0].GetPrice(),
					params[0].GetValue(),
					fees,
					true,
					false,
				); err != nil {
					return
				}

				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[0].GetValue(), false)
				if err := e.setBalance(
					params[0].GetBaseUnit(),
					params[0].GetUserId(),
					quantity,
					proto.Balance_PLUS,
				); err != nil {
					return
				}

				if err := e.setInternalTransfer(
					params[0].GetId(),
					params[0].GetUserId(),
					params[0].GetAssigning(),
					params[0].GetBaseUnit(),
					params[0].GetQuoteUnit(),
					params[0].GetPrice(),
					params[0].GetValue(),
					fees,
					false,
					true,
				); err != nil {
					return
				}

				break
			}

			e.Context.Logger.Infof("[(B)%v]: (Assigning: %v), (Price: [%v]-[%v]), (Value: [%v]-[%v]), (UserID: [%v]-[%v])", params[0].GetAssigning(), params[1].GetAssigning(), params[0].GetPrice(), params[1].GetPrice(), params[0].GetValue(), params[1].GetValue(), params[0].GetUserId(), params[1].GetUserId())
		}

	}

	if err := e.setTrade(params[0], params[1]); err != nil {
		return
	}

	if err := e.replayPusher(proto.Pusher_TradePublic, params[0].GetBaseUnit(), params[0].GetQuoteUnit()); err != nil {
		return
	}

	if err := e.replayPusher(proto.Pusher_TradeStatusPublic, params[0], params[1]); err != nil {
		return
	}
}

// replayDeposit - replenishment of assets, work with blockchain structures.
func (e *ExchangeService) replayDeposit() {
	e.run, e.wait, e.block = make(map[int64]bool), make(map[int64]bool), make(map[int64]int64)

	for {

		func() {

			var (
				chain proto.Chain
			)

			rows, err := e.Context.Db.Query("select id, rpc, rpc_key, platform, block, network, confirmation, parent_symbol from chains where status = $1", true)
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
				case proto.Platform_ETHEREUM:
					go e.depositEthereum(&chain)
					break
				case proto.Platform_TRON:
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
func (e *ExchangeService) replayWithdraw() {
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	for {

		func() {

			rows, err := e.Context.Db.Query(`select id, symbol, "to", chain_id, fees, value, price, platform, protocol, claim from transactions where status = $1 and tx_type = $2 and fin_type = $3`, proto.Status_PENDING, proto.TxType_WITHDRAWS, proto.FinType_CRYPTO)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			for rows.Next() {

				var (
					item, reserve, insure proto.Transaction
				)

				if err := rows.Scan(&item.Id, &item.Symbol, &item.To, &item.ChainId, &item.Fees, &item.Value, &item.Price, &item.Platform, &item.Protocol, &item.Claim); e.Context.Debug(err) {
					continue
				}

				chain, err := e.getChain(item.GetChainId(), true)
				if e.Context.Debug(err) {
					continue
				}

				if item.GetProtocol() == proto.Protocol_MAINNET {

					// Find the reserve asset from which funds will be transferred,
					// by its platform, as well as by protocol, symbol, and number of funds.
					if _ = e.Context.Db.QueryRow("select value, user_id from reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), false).Scan(&reserve.Value, &reserve.UserId); reserve.GetValue() > 0 {

						if _, err := e.Context.Db.Exec("update transactions set status = $2 where id = $1;", item.GetId(), proto.Status_PROCESSING); e.Context.Debug(err) {
							continue
						}

						if err := e.setReserveLock(reserve.GetUserId(), item.GetSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
							continue
						}

						switch item.GetPlatform() {
						case proto.Platform_ETHEREUM:
							e.transferEthereum(reserve.GetUserId(), item.GetId(), item.GetSymbol(), common.HexToAddress(item.GetTo()), item.GetValue(), 0, item.GetPlatform(), item.GetProtocol(), chain, true)
							break
						case proto.Platform_TRON:
							e.transferTron(reserve.GetUserId(), item.GetId(), item.GetSymbol(), item.GetTo(), item.GetValue(), 0, item.GetPlatform(), item.GetProtocol(), chain, true)
							break
						}

					}

				} else {

					// Find the reserve asset from which funds will be transferred,
					// by its platform, as well as by protocol, symbol, and number of funds.
					if _ = e.Context.Db.QueryRow("select a.value, a.user_id from reserves a inner join reserves b on case when a.protocol > 0 then b.user_id = a.user_id and b.symbol = $6 and b.platform = a.platform and b.protocol = 0 and b.value >= $5 and b.lock = $7 end where a.symbol = $1 and a.value >= $2 and a.platform = $3 and a.protocol = $4 and a.lock = $7", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), item.GetFees(), chain.GetParentSymbol(), false).Scan(&reserve.Value, &reserve.UserId); reserve.GetValue() > 0 {

						if _, err := e.Context.Db.Exec("update transactions set status = $2 where id = $1;", item.GetId(), proto.Status_PROCESSING); e.Context.Debug(err) {
							continue
						}

						if err := e.setReserveLock(reserve.GetUserId(), item.GetSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
							continue
						}

						switch item.GetPlatform() {
						case proto.Platform_ETHEREUM:
							e.transferEthereum(reserve.GetUserId(), item.GetId(), item.GetSymbol(), common.HexToAddress(item.GetTo()), item.GetValue(), item.GetPrice(), item.GetPlatform(), item.GetProtocol(), chain, true)
							break
						case proto.Platform_TRON:
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
								if _ = e.Context.Db.QueryRow("select value, address from reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), false).Scan(&reserve.Value, &reserve.To); reserve.GetValue() > 0 {

									if _ = e.Context.Db.QueryRow("select value, user_id from reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", chain.GetParentSymbol(), item.GetFees()*2, item.GetPlatform(), proto.Protocol_MAINNET, false).Scan(&insure.Value, &insure.UserId); insure.GetValue() > 0 {

										if err := e.setReserveLock(insure.GetUserId(), chain.GetParentSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
											continue
										}

										switch item.GetPlatform() {
										case proto.Platform_ETHEREUM:
											e.transferEthereum(insure.GetUserId(), item.GetId(), chain.GetParentSymbol(), common.HexToAddress(reserve.GetTo()), item.GetFees(), 0, item.GetPlatform(), proto.Protocol_MAINNET, chain, false)
											break
										case proto.Platform_TRON:
											e.transferTron(insure.GetUserId(), item.GetId(), chain.GetParentSymbol(), reserve.GetTo(), item.GetFees(), 0, item.GetPlatform(), proto.Protocol_MAINNET, chain, false)
											break
										}

										if _, err := e.Context.Db.Exec("update transactions set claim = $2 where id = $1;", item.GetId(), true); e.Context.Debug(err) {
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
func (e *ExchangeService) replayConfirmation() {

	rows, err := e.Context.Db.Query(`select id, hash, symbol, "to", fees, chain_id, user_id, value, confirmation, block, platform, protocol, create_at from transactions where status = $1 and tx_type = $2`, proto.Status_PENDING, proto.TxType_DEPOSIT)
	if e.Context.Debug(err) {
		return
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item proto.Transaction
		)

		if err := rows.Scan(&item.Id, &item.Hash, &item.Symbol, &item.To, &item.Fees, &item.ChainId, &item.UserId, &item.Value, &item.Confirmation, &item.Block, &item.Platform, &item.Protocol, &item.CreateAt); e.Context.Debug(err) {
			continue
		}

		chain, err := e.getChain(item.GetChainId(), true)
		if e.Context.Debug(err) {
			return
		}

		if (chain.GetBlock()-item.GetBlock()) >= chain.GetConfirmation() && item.GetConfirmation() >= chain.GetConfirmation() {

			currency, err := e.getCurrency(item.GetSymbol(), false)
			if e.Context.Debug(err) {
				return
			}

			if item.GetValue() >= currency.GetMinDeposit() {

				// Crediting a new deposit to the local wallet address.
				if _, err := e.Context.Db.Exec("update assets set balance = balance + $3 where symbol = $2 and user_id = $1;", item.GetUserId(), item.GetSymbol(), item.GetValue()); e.Context.Debug(err) {
					continue
				}

				item.Hook = true
				item.Status = proto.Status_FILLED

				if err := e.replayPusher(proto.Pusher_DepositPublic, &item); e.Context.Debug(err) {
					return
				}

			} else {
				item.Status = proto.Status_RESERVE
			}

			if err := e.setReserve(item.GetUserId(), item.GetTo(), item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), proto.Balance_PLUS); e.Context.Debug(err) {
				continue
			}

			// Update deposits pending status to success status.
			if _, err := e.Context.Db.Exec("update transactions set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
				continue
			}

		} else {
			if _, err := e.Context.Db.Exec("update transactions set confirmation = $2 where id = $1;", item.GetId(), chain.GetBlock()-item.GetBlock()); e.Context.Debug(err) {
				continue
			}
		}
	}
}

// replayPusher - public notification of customers about a new event.
func (e *ExchangeService) replayPusher(event proto.Pusher, params ...interface{}) error {

	switch event {
	case proto.Pusher_TradeStatusPublic:

		for i := 0; i < 2; i++ {
			if err := e.Context.Publish(e.getOrder(params[i].(*proto.Order).GetId()), "exchange", "order/status"); err != nil {
				return err
			}
		}
		break

	case proto.Pusher_TradePublic:

		for _, interval := range help.Depth() {

			migrate, err := e.GetGraph(context.Background(), &proto.GetExchangeRequestGraph{BaseUnit: params[0].(string), QuoteUnit: params[1].(string), Limit: 2, Resolution: interval})
			if e.Context.Debug(err) {
				continue
			}

			if err := e.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/graph:%v", interval)); err != nil {
				return err
			}
		}

		break

	case proto.Pusher_DepositPublic:

		if err := e.Context.Publish(params[0].(*proto.Transaction), "exchange", "deposit/open", "deposit/status"); err != nil {
			return err
		}

		break

	case proto.Pusher_WithdrawPublic:

		if err := e.Context.Publish(params[0].(*proto.Transaction), "exchange", "withdraw/status"); err != nil {
			return err
		}

		break

	case proto.Pusher_OrderCreatePublic:

		if err := e.Context.Publish(params[0].(*proto.Order), "exchange", "order/create"); err != nil {
			return err
		}

		break
	}

	return nil
}

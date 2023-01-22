package spot

import (
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"google.golang.org/grpc/status"
	"regexp"
	"strconv"
)

// helperAddress - проверка адресов.
func (e *Service) helperAddress(address string, platform pbspot.Platform) error {

	var regexMap = map[pbspot.Platform]string{
		pbspot.Platform_BITCOIN:  "^(bc1|[1b])[a-zA-HJ-NP-Z0-9]{25,39}$",
		pbspot.Platform_TRON:     "^([T])[a-zA-HJ-NP-Z0-9]{33}$",
		pbspot.Platform_ETHEREUM: "^(0x)[a-zA-Z0-9]{40}$",
	}

	regex, ok := regexMap[platform]
	if !ok {
		return e.Context.Error(status.Errorf(10789, "cryptocurrency not available: %s ", platform))
	}

	if !regexp.MustCompile(regex).MatchString(address) {
		return e.Context.Error(status.Errorf(90589, "the %s address you provided is not correct, %v", platform, address))
	}

	return nil
}

// helperSymbol - проверка на существование записи об криптовалюте.
func (e *Service) helperSymbol(symbol string) error {

	if _, err := e.getCurrency(symbol, true); err != nil {
		return e.Context.Error(status.Errorf(11580, "this %v currency does not exist, or this currency is disabled", symbol))
	}

	return nil
}

// helperPair - проверяем статус пары.
func (e *Service) helperPair(base, quote string) error {

	var (
		id int64
	)

	if _ = e.Context.Db.QueryRow("select id from spot_pairs where base_unit = $1 and quote_unit = $2 and status = $3", base, quote, false).Scan(&id); id > 0 {
		return e.Context.Error(status.Errorf(21605, "this pair %v-%v is temporarily unavailable", base, quote))
	}

	return nil
}

// helperWithdraw - проверка на существующие балансы, и их лимиты.
func (e *Service) helperWithdraw(quantity, reserve, balance, max, min, fees float64) error {

	var (
		proportion = decimal.New(min).Add(fees).Float()
	)

	if quantity > reserve {
		return e.Context.Error(status.Errorf(47784, "the claimed amount %v is greater than the reserve %v itself", quantity, reserve))
	}

	if quantity > balance {
		return e.Context.Error(status.Errorf(48584, "the claimed amount %v is more than what you have on your balance %v", quantity, balance))
	}

	if quantity < proportion {
		return e.Context.Error(status.Errorf(48880, "the withdrawal amount %v must not be less than the minimum amount: %v", quantity, proportion))
	}

	if quantity > max {
		return e.Context.Error(status.Errorf(70083, "the amount %v declared for withdrawal should not be more than allowed %v", quantity, max))
	}

	return nil
}

// helperInternalAsset - проверяем адрес для вывода средств, если адрес внутренний то не позволяем вывод.
func (e *Service) helperInternalAsset(address string) error {
	var (
		exist bool
	)

	_ = e.Context.Db.QueryRow("select exists(select id from spot_wallets where lower(address) = lower($1))::bool", address).Scan(&exist)

	if exist {
		return e.Context.Error(status.Errorf(717883, "you cannot use this address %v, this address is internal, please use another address", address))
	}

	return nil
}

// helperOrder - validate new order valid form.
func (e *Service) helperOrder(order *pbspot.Order) (summary float64, err error) {

	if order.GetPrice() == 0 {
		return 0, status.Errorf(65790, "impossible price %v", order.GetPrice())
	}

	switch order.GetAssigning() {
	case pbspot.Assigning_BUY:

		quantity := decimal.New(order.GetQuantity()).Mul(order.GetPrice()).Float()
		if min, max, ok := e.getRange(order.GetQuoteUnit(), quantity); !ok {
			return 0, status.Errorf(11623, "[quote]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		balance := e.getBalance(order.GetQuoteUnit(), order.GetUserId())
		if quantity > balance || order.GetQuantity() == 0 {
			return 0, status.Error(11586, "[quote]: there is not enough funds on your asset balance to place an order")
		}

		return quantity, nil

	case pbspot.Assigning_SELL:

		quantity := order.GetQuantity()
		if min, max, ok := e.getRange(order.GetBaseUnit(), order.GetQuantity()); !ok {
			return 0, status.Errorf(11587, "[base]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		balance := e.getBalance(order.GetBaseUnit(), order.GetUserId())
		if quantity > balance || order.GetQuantity() == 0 {
			return 0, status.Error(11624, "[base]: there is not enough funds on your asset balance to place an order")
		}

		return quantity, nil
	}

	return 0, status.Error(11596, "invalid input parameter")
}

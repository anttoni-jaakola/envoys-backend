package types

import "github.com/pkg/errors"

const (
	TypeZero   = ""
	TypeSpot   = "spot"
	TypeStock  = "stock"
	TypeCross  = "cross"
	TypeFuture = "future"

	KYCLevel1 = "level_1"
	KYCLevel2 = "level_2"
	KYCLevel3 = "level_3"

	UserTypeAgent  = "agent"
	UserTypeBroker = "broker"

	AssigningBuy    = "buy"
	AssigningSell   = "sell"
	AssigningSupply = "supply"

	AssigningOpen  = "open"
	AssigningClose = "close"

	PositionLong  = "long"
	PositionShort = "short"

	StatusCancel     = "cancel"
	StatusFilled     = "filled"
	StatusPending    = "pending"
	StatusReserve    = "reserve"
	StatusProcessing = "processing"
	StatusFailed     = "failed"
	StatusLock       = "lock"
	StatusAccess     = "access"
	StatsRejected    = "rejected"
	StatusBlocked    = "blocked"

	TradingMarket = "market"
	TradingLimit  = "limit"

	GroupAction = "action"
	GroupCrypto = "crypto"
	GroupFiat   = "fiat"

	AssignmentDeposit    = "deposit"
	AssignmentWithdrawal = "withdrawal"

	AllocationExternal = "external"
	AllocationInternal = "internal"
	AllocationReward   = "reward"

	PlatformBitcoin    = "bitcoin"
	PlatformEthereum   = "ethereum"
	PlatformTron       = "tron"
	PlatformVisa       = "visa"
	PlatformMastercard = "mastercard"

	PatternText = "text"

	BalanceMinus = "minus"
	BalancePlus  = "plus"

	TagNone      = "tag_none"
	TagBitcoin   = "tag_bitcoin"
	TagEthereum  = "tag_ethereum"
	TagBinance   = "tag_binance"
	TagTron      = "tag_tron"
	TagPolygon   = "tag_polygon"
	TagCronos    = "tag_cronos"
	TagFantom    = "tag_fantom"
	TagAvalanche = "tag_avalanche"

	ProtocolMainnet = "mainnet"
	ProtocolErc20   = "erc20"
	ProtocolErc721  = "erc721"
	ProtocolErc1155 = "erc1155"
	ProtocolErc998  = "erc998"
	ProtocolErc223  = "erc223"
	ProtocolBep20   = "bep20"
	ProtocolBep721  = "bep721"
	ProtocolBep1155 = "bep1155"
	ProtocolTrc20   = "trc20"
	ProtocolTrc721  = "trc721"
	ProtocolBep998  = "bep998"
	ProtocolBep223  = "bep223"
	ProtocolPrc20   = "prc20"
	ProtocolPrc721  = "prc721"
	ProtocolPrc1155 = "prc1155"
	ProtocolPrc998  = "prc998"
	ProtocolPrc223  = "prc223"
	ProtocolCrc20   = "crc20"
	ProtocolCrc721  = "crc721"
	ProtocolCrc1155 = "crc1155"
	ProtocolCrc998  = "crc998"
	ProtocolCrc223  = "crc223"
	ProtocolFrc20   = "frc20"
	ProtocolFrc721  = "frc721"
	ProtocolFrc1155 = "frc1155"
	ProtocolFrc998  = "frc998"
	ProtocolFrc223  = "frc223"
	ProtocolArc20   = "arc20"
	ProtocolArc721  = "arc721"
	ProtocolArc1155 = "arc1155"
	ProtocolArc998  = "arc998"
	ProtocolArc223  = "arc223"
)

// Tag - This function is used to check if a given string is a valid tag. It checks if the given string is present in the
// "tags" map. If it is not present, it returns an error with the message "No such tag exists".
func Tag(request string) error {
	tags := map[string]bool{
		TagNone:      true,
		TagBitcoin:   true,
		TagEthereum:  true,
		TagBinance:   true,
		TagTron:      true,
		TagPolygon:   true,
		TagCronos:    true,
		TagFantom:    true,
		TagAvalanche: true,
	}
	if _, ok := tags[request]; !ok {
		return errors.New("No such tag exists.")
	}
	return nil
}

// Platform - The purpose of this function is to check if a given string (request) is a valid platform. It looks through a map of
// valid platforms and returns an error if the requested platform does not exist.
func Platform(request string) error {
	platforms := map[string]bool{
		PlatformBitcoin:    true,
		PlatformEthereum:   true,
		PlatformTron:       true,
		PlatformVisa:       true,
		PlatformMastercard: true,
	}
	if _, ok := platforms[request]; !ok {
		return errors.New("No such platform exists.")
	}
	return nil
}

// Protocol - The purpose of this function is to serve as a lookup table of protocols that are supported by a system. It takes in a
// string representing a protocol and checks if it exists in the list of protocols. If it does, the function returns nil,
// otherwise it returns an error.
func Protocol(request string) error {
	protocols := map[string]bool{
		ProtocolMainnet: true,
		ProtocolErc20:   true,
		ProtocolErc721:  true,
		ProtocolErc1155: true,
		ProtocolErc998:  true,
		ProtocolErc223:  true,
		ProtocolBep20:   true,
		ProtocolBep721:  true,
		ProtocolBep1155: true,
		ProtocolTrc20:   true,
		ProtocolTrc721:  true,
		ProtocolBep998:  true,
		ProtocolBep223:  true,
		ProtocolPrc20:   true,
		ProtocolPrc721:  true,
		ProtocolPrc1155: true,
		ProtocolPrc998:  true,
		ProtocolPrc223:  true,
		ProtocolCrc20:   true,
		ProtocolCrc721:  true,
		ProtocolCrc1155: true,
		ProtocolCrc998:  true,
		ProtocolCrc223:  true,
		ProtocolFrc20:   true,
		ProtocolFrc721:  true,
		ProtocolFrc1155: true,
		ProtocolFrc998:  true,
		ProtocolFrc223:  true,
		ProtocolArc20:   true,
		ProtocolArc721:  true,
		ProtocolArc1155: true,
		ProtocolArc998:  true,
		ProtocolArc223:  true,
	}
	if _, ok := protocols[request]; !ok {
		return errors.New("No such protocol exists.")
	}
	return nil
}

// Status - The purpose of this code is to check if the requested status is valid. It does this by creating a map of accepted
// statuses and using an if statement to check if the requested status is in the map. If it is not, an error is returned.
func Status(request string) error {
	statuses := map[string]bool{
		StatusCancel:     true,
		StatusFilled:     true,
		StatusPending:    true,
		StatusReserve:    true,
		StatusProcessing: true,
		StatusFailed:     true,
		StatusLock:       true,
		StatusAccess:     true,
		StatsRejected:    true,
		StatusBlocked:    true,
	}
	if _, ok := statuses[request]; !ok {
		return errors.New("Invalid status")
	}
	return nil
}

func Type(request string) error {
	types := map[string]bool{
		TypeSpot:  true,
		TypeStock: true,
		TypeCross: true,
	}
	if _, ok := types[request]; !ok {
		return errors.New("Invalid type")
	}
	return nil
}

func Group(request string) error {
	groups := map[string]bool{
		GroupAction: true,
		GroupCrypto: true,
		GroupFiat:   true,
	}
	if _, ok := groups[request]; !ok {
		return errors.New("Invalid group")
	}
	return nil
}

func Position(request string) error {
	positions := map[string]bool{
		PositionLong:  true,
		PositionShort: true,
	}
	if _, ok := positions[request]; !ok {
		return errors.New("Invalid position")
	}
	return nil
}

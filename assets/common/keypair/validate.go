package keypair

import (
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
	"regexp"
)

// The purpose of these variables is to define regular expressions that can be used to validate Bitcoin, Tron, and
// Ethereum addresses. These regular expressions will ensure that the addresses entered are valid, and that they meet the
// correct address format for each cryptocurrency.
var (
	bitcoinRegex  = "^(bc1|[1b])[a-zA-HJ-NP-Z0-9]{25,39}$"
	tronRegex     = "^([T])[a-zA-HJ-NP-Z0-9]{33}$"
	ethereumRegex = "^(0x)[a-zA-Z0-9]{40}$"
)

// ValidateCryptoAddress - This function is used to validate a cryptocurrency address depending on the platform (Bitcoin, Ethereum, or Tron). It
// checks to see if the address given matches the regular expression of the platform provided. If there is no match, it
// returns an error.
func ValidateCryptoAddress(address string, platform string) error {
	var regex string

	// This code is setting up a switch statement to assign a regular expression (regex) to a given cryptocurrency platform.
	// If the platform is detected as Bitcoin, the regex variable will be assigned to the bitcoinRegex. If the platform is
	// Tron, the regex variable will be assigned to the tronRegex, and so on. If the platform is not one of the available
	// options, the switch statement will return an error.
	switch platform {
	case types.PlatformBitcoin:
		regex = bitcoinRegex
	case types.PlatformTron:
		regex = tronRegex
	case types.PlatformEthereum:
		regex = ethereumRegex
	default:
		return status.Errorf(10789, "cryptocurrency not available: %s ", platform)
	}

	// This code is checking if a given address matches a regular expression. If the address does not match the regular
	// expression, it will return an error with a message indicating that the address is not correct.
	if !regexp.MustCompile(regex).MatchString(address) {
		return status.Errorf(90589, "the %s address you provided is not correct, %v", platform, address)
	}

	return nil
}

package keypair

import (
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/sha3"
	"strings"
)

type CrossChain struct {
	extended *hdkeychain.ExtendedKey
}

// New - new generate address.
func (s *CrossChain) New(secret string, bytea []byte, platform proto.Platform) (a, p string, err error) {

	//e = bytea

	if len(bytea) == 0 {
		if bytea, err = bip39.NewEntropy(256); err != nil {
			return a, p, err
		}
	}

	seed, err := s.seed(bytea, secret)
	if err != nil {
		return a, p, err
	}

	switch platform {

	// BITCOIN Generate address.
	case proto.Platform_BITCOIN:
		private, err := s.master(seed, 44, 0, 0, 0, 0)
		if err != nil {
			return a, p, err
		}

		address, err := s.extended.Address(&chaincfg.Params{})
		if err != nil {
			return a, p, err
		}
		privateKeyBytes := crypto.FromECDSA(private.ToECDSA())

		return address.String(), hexutil.Encode(privateKeyBytes), nil

	// ETHEREUM Generate address.
	case proto.Platform_ETHEREUM:
		private, err := s.master(seed, 44, 60, 0, 0, 0)
		if err != nil {
			return a, p, err
		}
		privateKeyBytes := crypto.FromECDSA(private.ToECDSA())

		return strings.ToLower(crypto.PubkeyToAddress(private.PublicKey).String()), hexutil.Encode(privateKeyBytes), nil

	// TRON Generate address.
	case proto.Platform_TRON:
		private, err := s.master(seed, 44, 195, 0, 0, 0)
		if err != nil {
			return a, p, err
		}

		hash := sha3.NewLegacyKeccak256()
		hash.Write(append(private.PubKey().X.Bytes(), private.PubKey().Y.Bytes()...))
		hashed := hash.Sum(nil)

		bytes := append([]byte{0x41}, hashed[len(hashed)-20:]...)

		summary := sha256.Sum256(bytes)
		replay := sha256.Sum256(summary[:])

		privateKeyBytes := crypto.FromECDSA(private.ToECDSA())

		return base58.Encode(append(bytes, replay[:4]...)), hexutil.Encode(privateKeyBytes), nil
	}

	return a, p, nil
}

// master - new path to coin https://github.com/satoshilabs/slips/blob/master/slip-0044.md#registered-coin-types
func (s *CrossChain) master(seed []byte, paths ...uint32) (*btcec.PrivateKey, error) {

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	s.extended, err = masterKey.Derive(hdkeychain.HardenedKeyStart + paths[0])
	if err != nil {
		return nil, err
	}

	for i, path := range paths {

		if i == 0 {
			continue
		}

		if i < 3 {
			s.extended, err = s.extended.Derive(hdkeychain.HardenedKeyStart + path)
			if err != nil {
				return nil, err
			}
		} else {
			s.extended, err = s.extended.Derive(path)
			if err != nil {
				return nil, err
			}
		}

	}

	btcecPrivKey, err := s.extended.ECPrivKey()
	if err != nil {
		return nil, err
	}

	return (*btcec.PrivateKey)(btcecPrivKey.ToECDSA()), nil
}

// seed - new seed.
func (s *CrossChain) seed(entropy []byte, secret string) ([]byte, error) {

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic")
	}

	return bip39.NewSeedWithErrorChecking(mnemonic, secret)
}

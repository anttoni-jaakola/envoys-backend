package keypair

import (
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/sha3"
	"strings"
)

// The CrossChain struct is used to store an extended key from the HDKeychain package. This extended key is used to
// create a cross-chain transaction, which allows users to transfer digital assets between different blockchain networks.
type CrossChain struct {
	extended *hdkeychain.ExtendedKey
}

// New - This function is used to generate an address and a private key for a specific platform (Bitcoin, Ethereum, or Tron).
// It takes in a secret string, an array of bytes and a platform as parameters and returns the address, private key and
// any errors that may occur. It uses the BIP39 standard to generate the seed and then applies the seed to the chosen
// platform to generate the address and private key.
func (s *CrossChain) New(secret string, bytea []byte, platform string) (a, p string, err error) {

	// This is an if statement that checks the length of the variable bytea. If the length is equal to 0, then a certain
	// action is taken. This is used to ensure that an empty variable doesn't cause an error.
	if len(bytea) == 0 {

		// The purpose of this code is to generate a new 256-bit entropy value using the bip39 library and return it as a byte
		// array. If an error occurs, it will be returned instead of the byte array.
		if bytea, err = bip39.NewEntropy(256); err != nil {
			return a, p, err
		}
	}

	// This code is used to generate a seed from a byte array and a secret. If an error occurs while generating the seed,
	// the function returns the byte array, the secret, and the error.
	seed, err := s.seed(bytea, secret)
	if err != nil {
		return a, p, err
	}

	// The switch platform statement is a part of a programming language or syntax used to execute different commands
	// depending on the platform the code is running on. In this case, the switch platform statement is used to specify
	// different commands for different platforms, allowing the code to run in multiple environments.
	switch platform {

	case types.PlatformBitcoin:

		// This code is checking for an error when calling a function called "master" with the given parameters. If there is an
		// error, it returns an error instead of the values "a" and "p".
		private, err := s.master(seed, 44, 0, 0, 0, 0)
		if err != nil {
			return a, p, err
		}

		// This code is trying to obtain the address of the extended service (s.extended) using the chaincfg.Params{} as a
		// parameter. If an error occurs during the process, it will return the variables a, p, and err.
		address, err := s.extended.Address(&chaincfg.Params{})
		if err != nil {
			return a, p, err
		}

		// The purpose of this code is to convert the private key stored in the variable private to a byte array and store it
		// in the variable privateKeyBytes. This is necessary because private keys are often stored in byte array format for
		// security purposes.
		privateKeyBytes := crypto.FromECDSA(private.ToECDSA())

		return address.String(), hexutil.Encode(privateKeyBytes), nil

	case types.PlatformEthereum:

		// This code snippet is part of a function that is attempting to generate a master seed for an application. The
		// private, err := s.master(seed, 44, 60, 0, 0, 0) line is used to call the master function with the seed and other
		// parameters needed for the function to generate a master seed for the application. If an error occurs, the function
		// will return the error and halt any further execution.
		private, err := s.master(seed, 44, 60, 0, 0, 0)
		if err != nil {
			return a, p, err
		}

		// This statement is used to generate the private key bytes from a given private key. It takes the private key and uses
		// the ToECDSA() method to convert it to an ECDSA type, then uses the FromECDSA() method from the crypto package to
		// generate the private key bytes.
		privateKeyBytes := crypto.FromECDSA(private.ToECDSA())

		return strings.ToLower(crypto.PubkeyToAddress(private.PublicKey).String()), hexutil.Encode(privateKeyBytes), nil

	case types.PlatformTron:

		// This code is used to generate a master key from a given seed. The private variable is set to the result of the
		// s.master() function, which takes 6 parameters as input - seed, 44, 195, 0, 0, 0. If an error occurs, the function
		// returns the error and stops executing. Otherwise, the private variable is set to the result of the s.master() function.
		private, err := s.master(seed, 44, 195, 0, 0, 0)
		if err != nil {
			return a, p, err
		}

		// The purpose of this code is to create a new hash using the SHA3 algorithm with the Legacy Keccak256 variant. The
		// hash is used for cryptographic security purposes, such as message authentication, digital signatures, and data integrity.
		hash := sha3.NewLegacyKeccak256()

		// The purpose of the code is to write the X and Y coordinates of the public key associated with the given private key
		// to the hash object. The append function combines the X and Y coordinates into a single byte slice and the Write
		// function writes the bytes to the hash object.
		hash.Write(append(private.PubKey().X.Bytes(), private.PubKey().Y.Bytes()...))

		// hashed := hash.Sum(nil) is a line of code used to generate a hash value for a given input. The input is nil in this
		// case, which means that the hash function will generate a hash value from the current state of the hash. This is used
		// to generate a unique identifier for a given value, or to ensure the integrity of a given value.
		hashed := hash.Sum(nil)

		// The purpose of this code is to create an array of bytes that starts with the byte 0x41 and then includes the last 20
		// elements of the array 'hashed'. This is done by using to append() function which appends the second argument to the first argument.
		bytes := append([]byte{0x41}, hashed[len(hashed)-20:]...)

		// This line of code is used to generate a SHA-256 hash of the bytes provided as input. SHA-256 is a cryptographic
		// hashing algorithm which produces a 256-bit (32-byte) hash value. It is used to verify the integrity of data, as any
		// changes to the data will result in a different hash value. The summary variable will contain the resulting hash value.
		summary := sha256.Sum256(bytes)

		// The purpose of the code is to generate a cryptographic hash of the input data. In this case, the data being hashed
		// is the "summary" variable. The SHA256 algorithm is used to compute the hash, and the result is stored in the
		// "replay" variable.
		replay := sha256.Sum256(summary[:])

		// The purpose of this line of code is to create a byte array from a private key that is in ECDSA format. It does this
		// by using the crypto library to convert the private key to ECDSA, and then using the FromECDSA function to generate
		// the byte array.
		privateKeyBytes := crypto.FromECDSA(private.ToECDSA())

		return base58.Encode(append(bytes, replay[:4]...)), hexutil.Encode(privateKeyBytes), nil
	}

	return a, p, nil
}

// master - This function is used to generate a private key from a seed and a path specified by the paths variable. The private
// keys are derived using the hierarchical deterministic (HD) wallet protocol. The function takes the seed, which is used
// as the master key, and the paths variable that contains a list of integers representing the hierarchical path to the
// key. The function then derives the private key based on the path and returns it.
func (s *CrossChain) master(seed []byte, paths ...uint32) (*btcec.PrivateKey, error) {

	// This code is creating a master key by using the seed and the MainNetParams from the chaincfg package. It is used to
	// generate a hierarchical deterministic (HD) keychain which is used to generate a tree of public and private keys from
	// a single seed. The master key is the root of the tree, and all other keys are derived from it. If an error occurs,
	// the code returns an error.
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	// This code is used to derive a hardened key from a master key using a path. The code is trying to create a new key by
	// starting from the master key and using the path to derive the hardened key. The variable "s.extended" is assigned to
	// the hardened key and if there is an error in the derivation process, the code returns an error.
	s.extended, err = masterKey.Derive(hdkeychain.HardenedKeyStart + paths[0])
	if err != nil {
		return nil, err
	}

	// The for loop is used to iterate over the elements of the "paths" slice. The "i" variable is used for the index and
	// the "path" variable is used for the element at each index of the slice.
	for i, path := range paths {

		// The purpose of this code is to skip the rest of the current iteration of a loop and continue with the next
		// iteration. In this example, if the value of the variable i is equal to 0, the code will skip the rest of the loop
		// and continue with the next iteration.
		if i == 0 {
			continue
		}

		// This is an if statement, which is a conditional statement used in programming to control the flow of the program.
		// The statement checks if the value of the variable i is less than 3, and if it is, the code within the statement will be executed.
		if i < 3 {

			// This code is used to derive a private extended key from a given path using the hdkeychain package. The s.extended
			// is the extended key, and the path is the derivation path. The HardenedKeyStart is the starting point for the
			// derivation and is used to derive a private extended key. The code returns an error if there is an issue with the derivation.
			s.extended, err = s.extended.Derive(hdkeychain.HardenedKeyStart + path)
			if err != nil {
				return nil, err
			}

		} else {

			// This code is attempting to derive a new extended key from an existing extended key using a path. The purpose of the
			// code is to generate a new extended key from the existing one. If the Derive function returns an error, the code
			// will return nil and the error.
			s.extended, err = s.extended.Derive(path)
			if err != nil {
				return nil, err
			}
		}
	}

	// This code is attempting to retrieve the extended private key of the s object and create a btcecPrivKey object with
	// it. If this fails, an error is returned.
	btcecPrivKey, err := s.extended.ECPrivKey()
	if err != nil {
		return nil, err
	}

	return (*btcec.PrivateKey)(btcecPrivKey.ToECDSA()), nil
}

// This function is used to generate a seed from entropy bytes, a user-defined secret, and a mnemonic phrase. The purpose
// of this function is to generate a cryptographically secure seed that can be used to generate private keys, wallets,
// and accounts. It takes in entropy bytes, which is used to generate a mnemonic phrase, and a user-defined secret that
// is used to create the seed. It then checks to make sure the mnemonic phrase is valid. If the mnemonic phrase is valid,
// it uses it along with the secret to generate a seed, which is returned along with any errors.
func (s *CrossChain) seed(entropy []byte, secret string) ([]byte, error) {

	// This code is creating a mnemonic from entropy. The entropy is a random sequence of bytes used as the seed for
	// generating the mnemonic. The mnemonic is a human-readable phrase used to store and recall the seed. The purpose of
	// this code is to create a mnemonic from the entropy. If there is an error, the code will return an error.
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	// This code is used to check if a given mnemonic phrase is valid. The if statement checks to see if the mnemonic phrase
	// is not valid (indicated by the "!" operator). If the mnemonic is not valid, an error is returned.
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic")
	}

	return bip39.NewSeedWithErrorChecking(mnemonic, secret)
}

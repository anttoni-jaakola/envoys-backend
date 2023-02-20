package help

import (
	"golang.org/x/crypto/sha3"
)

// SignatureKeccak256 - The purpose of the above function is to generate a cryptographic hash of a given method using the Keccak-256
// algorithm. The result is a 32-byte array of data which is returned as the first four bytes of the generated hash. This
// function is typically used to create a signature for a given method to ensure the integrity of the data.
func SignatureKeccak256(method []byte) []byte {

	// The purpose of the hash is to generate a cryptographic hash using the SHA-3 algorithm with the legacy Keccak-256
	// hashing function. This is used to provide a secure, one-way digest of a given data set.
	hash := sha3.NewLegacyKeccak256()

	// The purpose of hash.Write(method) is to write data to a hashing object. Hashing objects are used to generate a unique
	// hash value for a given data set. The data set is typically a string of characters or a sequence of bytes. The
	// hash.Write(method) is used to write this data set to the hashing object so that it can be used to generate the unique hash value.
	hash.Write(method)

	return hash.Sum(nil)[:4]
}

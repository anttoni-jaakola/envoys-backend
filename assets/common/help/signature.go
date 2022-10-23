package help

import "golang.org/x/crypto/sha3"

// SignatureKeccak256 - method hash ethereum.
func SignatureKeccak256(method []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(method)
	return hash.Sum(nil)[:4]
}

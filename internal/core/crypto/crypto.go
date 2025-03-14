package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

type PublicKey struct {
	X *big.Int
	Y *big.Int
}

type Signature struct {
	R *big.Int
	S *big.Int
}

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*big.Int, *PublicKey) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	return privateKey.D, &PublicKey{
		X: privateKey.PublicKey.X,
		Y: privateKey.PublicKey.Y,
	}
}

// PrivateKeyToPublic converts a private key to its corresponding public key
func PrivateKeyToPublic(privKey *big.Int) *PublicKey {
	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult(privKey.Bytes())
	return &PublicKey{X: x, Y: y}
}

// Sign signs the given data with the private key
func Sign(data string, privKey *big.Int) *Signature {
	curve := elliptic.P256()
	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
		},
		D: privKey,
	}
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privKey.Bytes())

	hash := sha256.Sum256([]byte(data))
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		panic(err)
	}

	return &Signature{R: r, S: s}
}

// Verify verifies the signature of the data using the public key
func Verify(data string, sig *Signature, pub *PublicKey) bool {
	curve := elliptic.P256()
	publicKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     pub.X,
		Y:     pub.Y,
	}

	hash := sha256.Sum256([]byte(data))
	return ecdsa.Verify(publicKey, hash[:], sig.R, sig.S)
}

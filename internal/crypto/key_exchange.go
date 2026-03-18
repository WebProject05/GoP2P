package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

// GenerateKeyPair creates a private key and the public key bytes to send to the peer.
func GenerateKeyPair() (*ecdh.PrivateKey, []byte, error) {
	// We use the P-256 curve, which is highly secure and standard
	privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	
	publicKeyBytes := privateKey.PublicKey().Bytes()
	return privateKey, publicKeyBytes, nil
}

// ComputeSharedSecret takes your private key and the peer's public key 
// to generate the final AES encryption key.
func ComputeSharedSecret(privateKey *ecdh.PrivateKey, remotePubKeyBytes []byte) ([]byte, error) {
	remotePubKey, err := ecdh.P256().NewPublicKey(remotePubKeyBytes)
	if err != nil {
		return nil, errors.New("invalid public key received from peer")
	}

	// Perform the Diffie-Hellman mathematical exchange
	sharedSecret, err := privateKey.ECDH(remotePubKey)
	if err != nil {
		return nil, err
	}

	// Hash the resulting secret with SHA-256 to ensure it is exactly 
	// 32 bytes long, which is required for our AES-256-GCM encryption.
	hash := sha256.Sum256(sharedSecret)
	return hash[:], nil
}
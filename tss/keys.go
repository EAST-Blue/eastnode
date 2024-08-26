package tss

import (
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	group "github.com/bytemare/crypto"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

const peerKeyFile = "/.peer_key.json"
const frostKeyFile = "/.frost_key.json"

func GetPeerKey() (crypto.PrivKey, error) {
	var privKey crypto.PrivKey
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dirname + peerKeyFile); os.IsNotExist(err) {
		privKey, err = genPeerKey()
		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
	} else {
		privKey, err = loadPeerKey()
		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
	}

	return privKey, nil
}

func genPeerKey() (crypto.PrivKey, error) {
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generation error: %s", err)
	}
	err = savePeerKey(priv)
	if err != nil {
		return nil, fmt.Errorf("save key error: %s", err)
	}

	return priv, nil
}

func savePeerKey(privKey crypto.PrivKey) error {
	keyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to serialize key: %s", err)
	}

	id, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to get peer id: %s", err)
	}

	keyData := map[string]interface{}{
		"peer_id":  id.String(),
		"peer_key": keyBytes,
	}

	data, err := json.MarshalIndent(keyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key data: %s", err)
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := os.WriteFile(dirname+peerKeyFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %s", err)
	}

	return nil
}

func loadPeerKey() (crypto.PrivKey, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(dirname + peerKeyFile)
	if err != nil {
		return nil, err
	}

	var keyData map[string]interface{}
	if err := json.Unmarshal(data, &keyData); err != nil {
		return nil, err
	}

	if key, ok := keyData["peer_key"].(string); ok {
		keyBytes, err := b64.StdEncoding.DecodeString(key)
		if err != nil {
			return nil, err
		}

		privKey, err := crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return nil, err
		}

		return privKey, nil
	} else {
		return nil, fmt.Errorf("invalid key data")
	}

}

func SaveFrostKey(t int, n int, secretKey *group.Scalar, publicKey *group.Element, groupPublicKey *group.Element) error {
	keyData := map[string]interface{}{
		"threshold":        t,
		"participants":     n,
		"secret_key":       secretKey.Encode(),
		"public_key":       publicKey.Encode(),
		"group_public_key": groupPublicKey.Encode(),
	}

	data, err := json.MarshalIndent(keyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key data: %s", err)
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := os.WriteFile(dirname+frostKeyFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %s", err)
	}

	return nil
}

func LoadFrostKey() (int, int, []byte, []byte, []byte, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return 0, 0, nil, nil, nil, err
	}

	data, err := os.ReadFile(dirname + frostKeyFile)
	if err != nil {
		return 0, 0, nil, nil, nil, err
	}

	var keyData map[string]interface{}
	if err := json.Unmarshal(data, &keyData); err != nil {
		return 0, 0, nil, nil, nil, err
	}

	var t, n int
	var secretKey, publicKey, groupPublicKey []byte
	for key, value := range keyData {
		switch key {
		case "participants":
			n = int(value.(float64))
		case "threshold":
			t = int(value.(float64))
		case "secret_key":
			b, err := b64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return 0, 0, nil, nil, nil, nil
			}
			secretKey = b
		case "public_key":
			b, err := b64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return 0, 0, nil, nil, nil, nil
			}
			publicKey = b
		case "group_public_key":
			b, err := b64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return 0, 0, nil, nil, nil, nil
			}
			groupPublicKey = b
		}
	}

	return t, n, secretKey, publicKey, groupPublicKey, nil
}

package utils

import (
	"crypto"
	"encoding/binary"
	"encoding/hex"
	"github.com/near/borsh-go"
	"log"
	"os"
)

func Itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return b
}

func Btoi(b []byte) uint64 {
	v := binary.LittleEndian.Uint64(b)
	return v
}

func Cwd() string {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	return path
}

func SHA256(b []byte) string {
	hasher := crypto.SHA256.New()
	_, err := hasher.Write(b)

	if err != nil {
		log.Println(err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func SHA512(b []byte) string {
	hasher := crypto.SHA512.New()
	_, err := hasher.Write(b)

	if err != nil {
		log.Println(err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func DecodeHexAndBorshDeserialize(target interface{}, s string) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	err = borsh.Deserialize(target, bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BorshSerializeAndEncodeHex(target interface{}) string {
	res, err := borsh.Serialize(target)
	if err != nil {
		panic(err)
	}

	str := hex.EncodeToString(res)

	return str
}

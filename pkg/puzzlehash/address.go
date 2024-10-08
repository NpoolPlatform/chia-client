package puzzlehash

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/chia-network/go-chia-libs/pkg/types"
	bls "github.com/cloudflare/circl/ecc/bls12381"
)

func GetAddressFromPuzzleHash(ph []byte, prefix string) (string, error) {
	bits, err := bech32.ConvertBits(ph, 8, 5, true)
	if err != nil {
		return "", nil
	}
	return bech32.EncodeM(prefix, bits)
}

func GetPuzzleHashFromAddress(address string) (string, *types.Bytes32, error) {
	prefix, data, err := bech32.Decode(address)
	if err != nil {
		return "", nil, err
	}
	puzzleHash, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, err
	}

	puzzleHashBytes, err := types.BytesToBytes32(puzzleHash)
	if err != nil {
		return "", nil, err
	}

	return prefix, &puzzleHashBytes, nil
}

func NewAddressFromPkBytes(pkBytes []byte, prefix string) (string, error) {
	var p bls.G1
	err := p.SetBytes(pkBytes)
	if err != nil {
		return "", err
	}

	return NewAddressFromPK(&p, prefix)
}

func NewPuzzleHashFromPkBytes(pkBytes []byte) (string, error) {
	var p bls.G1
	err := p.SetBytes(pkBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(NewPuzzleHashFromPK(&p)), nil
}

func NewPuzzleHashBytesFromPkBytes(pkBytes []byte) ([]byte, error) {
	var p bls.G1
	err := p.SetBytes(pkBytes)
	if err != nil {
		return nil, err
	}

	return NewPuzzleHashFromPK(&p), nil
}

func NewAddressFromPK(pk *bls.G1, prefix string) (string, error) {
	bits, err := bech32.ConvertBits(NewPuzzleHashFromPK(pk), 8, 5, true)
	if err != nil {
		return "", nil
	}
	return bech32.EncodeM(prefix, bits)
}

func NewPuzzleHashFromPK(pk *bls.G1) []byte {
	return genAddress(pk.BytesCompressed())
}

func NewAddressFromPKHex(pkHex, prefix string) (string, error) {
	pkHex = strings.TrimPrefix(pkHex, "0x")
	pkBytes, err := hex.DecodeString(pkHex)
	if err != nil {
		return "", err
	}
	return NewAddressFromPkBytes(pkBytes, prefix)
}

func genAddress(pkBytes []byte) []byte {
	programBytes := NewProgramBytes(pkBytes)
	programString := hex.EncodeToString(programBytes)

	var (
		hash      = sha256.New()
		hashStack = NewAddressStack()

		hashBuf = make([]byte, 32)     // 32
		tmpBuf  = make([]byte, 1+2*32) // 1 + 32 + 32

		text    = ""
		firstFF = 0
	)

	for i := len(programString) - 2; i >= 0; i -= 2 {
		if firstFF == 1 {
			firstFF += 1
			text = programString[i : i+96+2] // 2(len) + 96(hash)
		} else {
			text = programString[i : i+2]
		}
		switch text {
		case "ff":
			p0 := hashStack.Pop()
			p1 := hashStack.Pop()
			copy(tmpBuf[:1], []byte{02})
			copy(tmpBuf[1:1+32], p0)
			copy(tmpBuf[1+32:], p1)
			hash.Write(tmpBuf)
			hashBuf = hash.Sum(nil)
			hash.Reset()
			hashStack.Append(hashBuf)
			if firstFF == 0 {
				firstFF += 1
				i -= 96
			}
		default:
			switch text {
			case "80":
				hash.Write([]byte{01})
			default:
				if len(text) != 2 {
					copy(tmpBuf[:1], []byte{01})
					decodeText, _ := hex.DecodeString(text[2:])
					copy(tmpBuf[1:49], decodeText)
					hash.Write(tmpBuf[:49])
				} else {
					copy(tmpBuf[:1], []byte{01})
					decodeText, _ := hex.DecodeString(text)
					copy(tmpBuf[1:2], decodeText)
					hash.Write(tmpBuf[:2])
				}
			}
			hashBuf = hash.Sum(nil)
			hashStack.Append(hashBuf)
			hash.Reset()
		}
	}
	return hashStack.hashes[0]
}

func NewProgramBytes(pkBytes []byte) []byte {
	ret := "ff02ffff01ff02ffff01ff02ffff03ff0bffff01ff02ffff03ffff09ff05ffff1dff0bffff1effff0bff0bffff02ff06fff" +
		"f04ff02ffff04ff17ff8080808080808080ffff01ff02ff17ff2f80ffff01ff088080ff0180ffff01ff04ffff04ff04ffff" +
		"04ff05ffff04ffff02ff06ffff04ff02ffff04ff17ff80808080ff80808080ffff02ff17ff2f808080ff0180ffff04ffff0" +
		"1ff32ff02ffff03ffff07ff0580ffff01ff0bffff0102ffff02ff06ffff04ff02ffff04ff09ff80808080ffff02ff06ffff" +
		"04ff02ffff04ff0dff8080808080ffff01ff0bffff0101ff058080ff0180ff018080ffff04ffff01b0" +
		hex.EncodeToString(pkBytes) +
		"ff018080"
	programBytes, err := hex.DecodeString(ret)
	if err != nil {
		panic(err)
	}
	return programBytes
}

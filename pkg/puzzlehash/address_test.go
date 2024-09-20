package puzzlehash

import (
	"encoding/hex"
	"testing"

	bls "github.com/cloudflare/circl/ecc/bls12381"
	"github.com/stretchr/testify/assert"
)

var (
	testPkBytes, _ = hex.DecodeString("940dd34f3e9edd8db760ea233ba45ab820e63ce2e4b19e3e185ed7f32b5950442863181a03a24185dfbefeb6f8c001e8")

	testAddress = "txch1wtlzkt6kmykf3d2e422hsk050ggxu95zguw5qn4sjgkgc2u0hygq6wwcwk"
)

func TestGenerateAddress(t *testing.T) {
	var p bls.G1
	err := p.SetBytes(testPkBytes)
	if err != nil {
		t.Fatal(err)
	}

	addr1, err := NewAddressFromPkBytes(testPkBytes, "txch")
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}
	addr2, err := NewAddressFromPK(&p, "txch")
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}
	addr3, err := NewAddressFromPKHex(hex.EncodeToString(p.BytesCompressed()), "txch")
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, addr1, testAddress)
	assert.Equal(t, addr2, testAddress)
	assert.Equal(t, addr3, testAddress)
}

func TestPH2Addr(t *testing.T) {
	prefix, puzzleHash, err := GetPuzzleHashFromAddress(testAddress)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}
	address, err := GetAddressFromPuzzleHash(puzzleHash[:], prefix)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, address, testAddress)
}

func TestPuzzleReveal(t *testing.T) {
	pkHex := "b8d50671a208e33f1fd8f85b664f5776a106a3f0c615da5068ca0fc153be606622bc4ac928ef3cf8283241ef4f44a866"
	puzzleReveal := "ff02ffff01ff02ffff01ff02ffff03ff0bffff01ff02ffff03ffff09ff05ffff1dff0bffff1effff0bff0bffff02ff06ffff04ff02ffff04ff17ff8080808080808080ffff01ff02ff17ff2f80ffff01ff088080ff0180ffff01ff04ffff04ff04ffff04ff05ffff04ffff02ff06ffff04ff02ffff04ff17ff80808080ff80808080ffff02ff17ff2f808080ff0180ffff04ffff01ff32ff02ffff03ffff07ff0580ffff01ff0bffff0102ffff02ff06ffff04ff02ffff04ff09ff80808080ffff02ff06ffff04ff02ffff04ff0dff8080808080ffff01ff0bffff0101ff058080ff0180ff018080ffff04ffff01b0b8d50671a208e33f1fd8f85b664f5776a106a3f0c615da5068ca0fc153be606622bc4ac928ef3cf8283241ef4f44a866ff018080"
	pkBytes, err := hex.DecodeString(pkHex)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	_puzzleReveal := NewProgramBytes(pkBytes)
	assert.Equal(t, puzzleReveal, hex.EncodeToString(_puzzleReveal))
}

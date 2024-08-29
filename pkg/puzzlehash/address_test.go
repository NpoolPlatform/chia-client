package puzzlehash

import (
	"encoding/hex"
	"testing"

	bls "github.com/cloudflare/circl/ecc/bls12381"
	"github.com/stretchr/testify/assert"
)

var (
	testPkBytes, _ = hex.DecodeString("940dd34f3e9edd8db760ea233ba45ab820e63ce2e4b19e3e185ed7f32b5950442863181a03a24185dfbefeb6f8c001e8")

	testAddress = "txch12g9w53gn445fj8j4mesrvqphry887y25tuja2renw4uhza7lamhqxwsmtu"
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
	address, err := GetAddressFromPuzzleHash(puzzleHash, prefix)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, address, testAddress)

}

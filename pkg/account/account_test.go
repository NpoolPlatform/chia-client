package account_test

import (
	"encoding/hex"
	"testing"

	"github.com/NpoolPlatform/chia-client/pkg/account"
	"github.com/stretchr/testify/assert"
)

func TestAccount(t *testing.T) {
	seedStr := "2d1907c542688aa32524bb82768275c3cc558661c88057b5000e40964dd8919392249040ea65cc72fd9f94e341d0c4461fce18c8287ba1387a0b182a57975524"
	pkStr := "92007aa08652875018c475872d0a3f19f3432dba01ca2f8ad46ed519179649e959232378dc81cd59ca0cd0337aae9a8b"
	skStr := "135bd00b0c64d861b8047c039020e32b2cb3056f8d0d714f7eed2fc96934ed47"

	seedBytes, err := hex.DecodeString(seedStr)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	pkBytes, err := hex.DecodeString(pkStr)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	skBytes, err := hex.DecodeString(skStr)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	acc1, err := account.GenAccountBySeedBytes(seedBytes)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	priBytes, err := acc1.PrivateKey.MarshalBinary()
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, skBytes, priBytes)

	pubBytes, err := acc1.PublicKey().MarshalBinary()
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, pkBytes, pubBytes)
}

func TestPuzzleHash(t *testing.T) {
	puzzleHash := "txch12g9w53gn445fj8j4mesrvqphry887y25tuja2renw4uhza7lamhqxwsmtu"
	skStr := "1c6198abdad4569b09554e48abc7f78d2c2833ed8235b862171a0ecf9db62d51"

	skBytes, err := hex.DecodeString(skStr)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	acc1, err := account.GenAccountBySKBytes(skBytes)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	priBytes, err := acc1.PrivateKey.MarshalBinary()
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, skBytes, priBytes)

	phStr, err := acc1.GetPuzzleHash(false)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, puzzleHash, phStr)
}

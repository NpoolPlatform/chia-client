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
	puzzleHash := "txch1wtlzkt6kmykf3d2e422hsk050ggxu95zguw5qn4sjgkgc2u0hygq6wwcwk"
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

	phStr, err := acc1.GetAddress(false)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	assert.Equal(t, puzzleHash, phStr)
}

func TestSign(t *testing.T) {
	fromSKHex := "3fefe074898e3ac7c6c17a40ec390d7c4ade53fde6c39339a93d03012bd3b7f7"
	msg := "hello"
	signatureHex := "b6cff30cf27be57b6923535a4c84324160a13f166c553bf38935807848f5755ef53965e71a6bc878106ec880ac47e27b0c9aa5a69e74d89b7d432389e3468956f58e854a224b194b11d58e9a935c2533301463c42cf8cde4d5cca3a4f0dd9a2c"

	// fromPKHex := "b5cdc71cbceee853fdc397a209640097852496d2611c252c41477dc68ea54f2b507b9a34cc909f77a70ea06824774a3d"
	// fromAddress := "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt"
	fromAcc, err := account.GenAccountBySKHex(fromSKHex)
	if !assert.Nil(t, err) {
		t.Fatal(err)
	}

	ret := hex.EncodeToString(fromAcc.Sign([]byte(msg)))
	assert.Equal(t, signatureHex, ret)
}

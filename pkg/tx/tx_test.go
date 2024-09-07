package tx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnsignedTx(t *testing.T) {
	tx, err := GenerateUnsignedTransaction("90", "0x5baecec1bc13676097ad96b29a73d599d93871b75981ec973f5c376486d08b4d", "14")
	assert.Nil(t, err)
	fmt.Println("tx: ", tx)
}

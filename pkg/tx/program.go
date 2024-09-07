package tx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	types1 "github.com/chia-network/go-chia-libs/pkg/types"
)

type treeNode struct {
	left  *treeNode
	right *treeNode
	val   []byte
}

func conditionChangeTreeHash() {
	h1, _ := types1.BytesFromHexString("0xeb22326630ee4a8f7f7c459fbf6e60dce37983597aa59ad97b18d12ec1004e0d")
	h2, _ := types1.BytesFromHexString("0xe25b0ff7a50e4afae386cdab538c70983db7f04fa835b45855114f9d790c414a")
	h3, _ := types1.BytesFromHexString("0xe4fb5940aef16e2c6cfed28ea3934dada96f4ce2b629b17cd482eb31f6a6c559")
	s1, _ := types1.BytesFromHexString("0x03e8")
	s2, _ := types1.BytesFromHexString("0x3a35293fb4")
	tree := treeNode{
		left: &treeNode{val: []byte{1}},
		right: &treeNode{
			left: &treeNode{
				left: &treeNode{val: []byte{60}},
				right: &treeNode{
					left:  &treeNode{val: h1},
					right: &treeNode{val: []byte{}},
				},
			},
			right: &treeNode{
				left: &treeNode{
					left: &treeNode{val: []byte{51}},
					right: &treeNode{
						left: &treeNode{val: h2},
						right: &treeNode{
							left:  &treeNode{val: s1},
							right: &treeNode{val: []byte{}},
						},
					},
				},
				right: &treeNode{
					left: &treeNode{
						left: &treeNode{val: []byte{51}},
						right: &treeNode{
							left: &treeNode{val: h3},
							right: &treeNode{
								left:  &treeNode{val: s2},
								right: &treeNode{val: []byte{}},
							},
						},
					},
					right: &treeNode{
						left: &treeNode{
							left: &treeNode{val: []byte{52}},
							right: &treeNode{
								left:  &treeNode{val: []byte{100}},
								right: &treeNode{val: []byte{}},
							},
						},
						right: &treeNode{val: []byte{}},
					},
				},
			},
		},
	}

	treeH := sha256tree(&tree)
	fmt.Println(hex.EncodeToString(treeH[:]))
}

func conditionAssertTreeHash() {
	h1, _ := types1.BytesFromHexString("0xece219712852ba4ce7ef918d7cf9bad18907c13213482f1ec6bb1fa848476ee6")
	tree1 := treeNode{
		left: &treeNode{val: []byte{1}},
		right: &treeNode{
			left: &treeNode{
				left: &treeNode{val: []byte{61}},
				right: &treeNode{
					left:  &treeNode{val: h1},
					right: &treeNode{val: []byte{}},
				},
			},
			right: &treeNode{val: []byte{}},
		},
	}

	treeH1 := sha256tree(&tree1)

	fmt.Println(hex.EncodeToString(treeH1[:]))
}

func sha256tree(v *treeNode) [32]byte {
	sBytes := []byte{}
	if v.left != nil {
		left := sha256tree(v.left)
		right := sha256tree(v.right)

		sBytes = append(sBytes, byte(2))
		sBytes = append(sBytes, left[:]...)
		sBytes = append(sBytes, right[:]...)
	} else {
		sBytes = append(sBytes, byte(1))
		sBytes = append(sBytes, []byte(v.val)...)
	}

	return sha256.Sum256(sBytes)
}

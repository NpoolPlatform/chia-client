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

func ptrStr(s string) *string {
	return &s
}

func demo() {
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

// "ba0d3a164a206b6dfb8d8fdf8da7aea1b94de99e604586eb9e11f34e51b0d1fa"
// def sha256tree(v):
//     pair = v.pair
//     if pair:
//         left = sha256tree(pair[0])
//         right = sha256tree(pair[1])
//         s = b"\2" + left + right
//     else:
//         s = b"\1" + v.atom
//     return hashlib.sha256(s).digest()

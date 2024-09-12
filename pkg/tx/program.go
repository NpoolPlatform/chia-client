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

func (h *txHandler) conditionChangeTreeHash() string {
	announcementMessage, _ := types1.BytesFromHexString(*h.announcementMessage)
	toPuzzleHash, _ := types1.BytesFromHexString(*h.toPuzzleHash)
	changePuzzleHash, _ := types1.BytesFromHexString(*h.changePuzzleHash)
	transferAmount, _ := types1.BytesFromHexString(h.amount.String())
	changeAmount, _ := types1.BytesFromHexString(h.changeAmount.String())
	feeAmount, _ := types1.BytesFromHexString(h.fee.String())

	tree := treeNode{
		left: &treeNode{val: []byte{1}},
		right: &treeNode{
			left: &treeNode{
				left: &treeNode{val: []byte{60}},
				right: &treeNode{
					left:  &treeNode{val: announcementMessage},
					right: &treeNode{val: []byte{}},
				},
			},
			right: &treeNode{
				left: &treeNode{
					left: &treeNode{val: []byte{51}},
					right: &treeNode{
						left: &treeNode{val: toPuzzleHash},
						right: &treeNode{
							left:  &treeNode{val: transferAmount},
							right: &treeNode{val: []byte{}},
						},
					},
				},
				right: &treeNode{
					left: &treeNode{
						left: &treeNode{val: []byte{51}},
						right: &treeNode{
							left: &treeNode{val: changePuzzleHash},
							right: &treeNode{
								left:  &treeNode{val: changeAmount},
								right: &treeNode{val: []byte{}},
							},
						},
					},
					right: &treeNode{
						left: &treeNode{
							left: &treeNode{val: []byte{52}},
							right: &treeNode{
								left:  &treeNode{val: feeAmount},
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
	return hex.EncodeToString(treeH[:])
}

func (h *txHandler) conditionAssertTreeHash() string {
	announcementID, _ := types1.BytesFromHexString(*h.announcementID)
	tree1 := treeNode{
		left: &treeNode{val: []byte{1}},
		right: &treeNode{
			left: &treeNode{
				left: &treeNode{val: []byte{61}},
				right: &treeNode{
					left:  &treeNode{val: announcementID},
					right: &treeNode{val: []byte{}},
				},
			},
			right: &treeNode{val: []byte{}},
		},
	}

	treeH1 := sha256tree(&tree1)
	return hex.EncodeToString(treeH1[:])
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

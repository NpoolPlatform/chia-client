package main

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/chia-client/pkg/account"
	"github.com/NpoolPlatform/chia-client/pkg/client"
	"github.com/NpoolPlatform/chia-client/pkg/transaction"
)

type TT struct {
	Txid string
}

type SS struct {
	Txid string
	Tf   string
	Th   string
}

func main() {
	acc, err := account.GenAccount()
	fmt.Println(err)
	fmt.Println(acc.GetAddress(true))

	skHex, err := acc.GetSKHex()
	fmt.Println(err)
	fmt.Println(skHex)

	skHexBytes := []byte(skHex)

	acc1, err := account.GenAccountBySKHex(string(skHexBytes))
	fmt.Println(err)
	fmt.Println(acc1.GetAddress(true))

	// TxDemo()
	// TestCreateSolution()
	// TestTreeHash()
	// TestClient()
}

func TestClient() {
	cli := client.NewClient("172.16.31.202:18444")
	fmt.Println(cli.CheckCoinsIsSpent(context.Background(), []string{"0xd30e03a6bbddfbbb2f9cfc5d2390ea0390ade6f5b2efb336464ba45c4eac805f"}))
}

func TxDemo() {
	// ----------------------------Check Node Heath-----------------------------
	cli := client.NewClient("172.16.31.202:18444")
	synced, err := cli.GetSyncStatus(context.Background())
	if err != nil {
		fmt.Println(1, err)
		return
	}
	if !synced {
		fmt.Println("node have not synced")
		return
	}

	// ----------------------------Prepare UnsignedTX-----------------------------
	From := "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt"
	To := "txch1pccwlj52r39yul8hp5mm3q96462up8k3xk83muwjyjhvy2vxnqwsnt40tz"
	Amount := uint64(0x7f81)
	Fee := uint64(100)

	unsignedTx, err := transaction.GenUnsignedTx(context.Background(), cli, From, To, Amount, Fee)
	if err != nil {
		fmt.Println(2, err)
		return
	}

	// ----------------------------SignTx-----------------------------
	// fromSKHex := "3fefe074898e3ac7c6c17a40ec390d7c4ade53fde6c39339a93d03012bd3b7f7"
	// fromPKHex := "b5cdc71cbceee853fdc397a209640097852496d2611c252c41477dc68ea54f2b507b9a34cc909f77a70ea06824774a3d"
	// fromAddress := "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt"
	fromSKHex := "3fefe074898e3ac7c6c17a40ec390d7c4ade53fde6c39339a93d03012bd3b7f7"
	spendBundle, err := transaction.GenSignedSpendBundle(unsignedTx, fromSKHex)
	if err != nil {
		fmt.Println(3, err)
		return
	}

	// ----------------------------BroadcostTX-----------------------------
	fmt.Println(client.PrettyStruct(spendBundle))

	txid, err := cli.PushTX(context.Background(), spendBundle)
	if err != nil {
		fmt.Println(4, err)
		return
	}

	// ----------------------------SyncTX-----------------------------
	fmt.Println(cli.CheckTxIDInMempool(context.Background(), txid))
	fmt.Println(cli.CheckCoinsIsSpent(context.Background(), unsignedTx.SpentCoinIDs))
}

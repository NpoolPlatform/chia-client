package tx

import (
	"crypto/sha256"
	"fmt"

	wlog "github.com/NpoolPlatform/go-service-framework/pkg/wlog"

	puzzle1 "github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
	"github.com/NpoolPlatform/chia-client/pkg/types"
	"github.com/shopspring/decimal"
)

// TODO
func SelectCoins(amount decimal.Decimal) ([]types.Coin, error) {
	return []types.Coin{}, nil
}

// TODO
func GetSpendableBalance() (decimal.Decimal, error) {
	return decimal.RequireFromString("0"), nil
}

// TODO
func DecoratedTargetPuzzleHash(innerPuzzle string, newPuzzleHash string) (types.Bytes32, error) {
	puzzleHash, err := types.Bytes32FromHexString(newPuzzleHash)
	if err != nil {
		return types.Bytes32{}, err
	}
	// TODO
	return puzzleHash, nil
}

// TODO
func GetNewPuzzleHash() (types.Bytes32, error) {
	return types.Bytes32{}, nil
}

// TODO
func MakeSolution(primaries []types.Payment, fee decimal.Decimal, conditions []types.Condition) (types.SerializedProgram, error) {
	return types.SerializedProgram{}, nil
}

// TODO
func MakeSpend(coin types.Coin, puzzleReveal types.SerializedProgram, solution types.SerializedProgram) (types.CoinSpend, error) {
	return types.CoinSpend{}, nil
}

func stdHash(myBytes []types.Bytes32) string {
	newBytes := sha256.New()
	for _, b := range myBytes {
		newBytes.Write(b[:])
	}
	newBytes.Sum(nil)
	_bytes := fmt.Sprintf("%x", newBytes)

	return _bytes
}

// TODO
func CreateCoinAnnouncement(message string) types.Condition {
	return &types.MyCondition{}
}

// TODO
func AssertCoinAnnouncement(assertedID string, assertedMsg string) types.Condition {
	return &types.MyCondition{}
}

// TODO
func GetPuzzleProgramFromPuzzle(puzzle string) types.SerializedProgram {
	return types.SerializedProgram{}
}

func GenerateUnsignedTransaction(
	amount string,
	newPuzzleHash string,
	fee string,
) ([]types.CoinSpend, error) {
	_amount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	_fee, err := decimal.NewFromString(fee)
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	totalAmount := _amount.Add(_fee)

	totalBalance, err := GetSpendableBalance()
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	if totalAmount.Cmp(totalBalance) > 0 {
		return nil, wlog.Errorf("insufficient funds")
	}

	coins, err := SelectCoins(totalAmount)
	if err != nil {
		return nil, err
	}
	if len(coins) == 0 {
		return nil, wlog.Errorf("invalid coin")
	}

	spendAmount := decimal.NewFromInt(0)
	for _, coin := range coins {
		coinAmount := fmt.Sprintf("%d", coin.Amount)
		spendAmount.Add(decimal.RequireFromString(coinAmount))
	}

	change := spendAmount.Sub(totalAmount)
	if change.Cmp(decimal.NewFromInt(0)) < 0 {
		return nil, wlog.Errorf("negative change %s", change)
	}

	var primaryAnnouncement types.Condition
	originID := coins[0].ID()
	primaries := []types.Payment{} // about change and transfer amount
	spends := []types.CoinSpend{}
	for _, coin := range coins {
		if originID == coin.ID() {
			puzzle := puzzle1.NewProgramString(coin.PuzzleHash[:])
			decoratedTargetPuzzleHash, err := DecoratedTargetPuzzleHash(puzzle, newPuzzleHash)
			if err != nil {
				return nil, err
			}

			primaries = append(primaries, types.Payment{
				PuzzleHash: decoratedTargetPuzzleHash,
				Amount:     _amount.BigInt().Uint64(),
				Memos:      []types.Bytes{},
			})

			if change.Cmp(decimal.NewFromInt(0)) > 0 {
				changePuzzleHash, err := GetNewPuzzleHash()
				if err != nil {
					return nil, err
				}
				primaries = append(primaries, types.Payment{
					PuzzleHash: changePuzzleHash,
					Amount:     change.BigInt().Uint64(),
					Memos:      []types.Bytes{},
				})
			}

			messages := []types.Bytes32{}
			for _, c := range coins {
				messages = append(messages, c.ID())
			}

			for _, primary := range primaries {
				c := types.Coin{
					ParentCoinInfo: coin.ID(),
					PuzzleHash:     primary.PuzzleHash,
					Amount:         primary.Amount,
				}
				messages = append(messages, c.ID())
			}
			message := stdHash(messages)
			announcement := CreateCoinAnnouncement(message)

			solution, err := MakeSolution(primaries, decimal.RequireFromString(fee), []types.Condition{announcement})
			if err != nil {
				return nil, err
			}

			primaryAnnouncement = AssertCoinAnnouncement(coin.ID().String(), message)
			spend, err := MakeSpend(coin, GetPuzzleProgramFromPuzzle(puzzle), solution)
			if err != nil {
				return nil, err
			}
			spends = append(spends, spend)
		}
	}
	// process the non-origin coins now that we have the primary announcement hash
	for _, coin := range coins {
		if coin.ID() == originID {
			continue
		}
		puzzle := puzzle1.NewProgramString(coin.PuzzleHash[:])
		solution, err := MakeSolution([]types.Payment{}, decimal.RequireFromString(fee), []types.Condition{primaryAnnouncement})
		if err != nil {
			return nil, err
		}
		spend, err := MakeSpend(coin, GetPuzzleProgramFromPuzzle(puzzle), solution)
		if err != nil {
			return nil, err
		}
		spends = append(spends, spend)
	}
	return spends, nil
}

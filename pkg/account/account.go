package account

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/NpoolPlatform/chia-client/pkg/bls"
	"github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
)

// have not impl verify signature,
// please use 'chia keys verify --message "" --signature "" --public_key ""' to verify

const (
	IKM_BYTES_LEN = 128

	TPREFIX = "txch"
	PREFIX  = "xch"
)

type Account struct {
	ikm []byte
	*bls.PrivateKey[bls.G1]
}

func CheckAddress(address string, mainnet bool) error {
	prefix, _, err := puzzlehash.GetPuzzleHashFromAddress(address)
	if err != nil {
		return err
	}
	if mainnet && prefix == PREFIX {
		return nil
	}
	if !mainnet && prefix == TPREFIX {
		return nil
	}
	return fmt.Errorf("invalid address for network")
}

func GenAccount() (*Account, error) {
	chiaAcc := &Account{
		ikm: make([]byte, IKM_BYTES_LEN),
	}
	rand.Read(chiaAcc.ikm)

	if err := chiaAcc.genSKFromSeed(); err != nil {
		return nil, err
	}

	return chiaAcc, nil
}

func GenAccountBySeedBytes(seedBytes []byte) (*Account, error) {
	chiaAcc := &Account{
		ikm: seedBytes,
	}
	if err := chiaAcc.genSKFromSeed(); err != nil {
		return nil, err
	}

	return chiaAcc, nil
}

func GenAccountBySKBytes(skBytes []byte) (*Account, error) {
	sk, err := bls.KeyGenFromSKBytes[bls.G1](skBytes)
	if err != nil {
		return nil, err
	}

	chiaAcc := &Account{}
	chiaAcc.PrivateKey = sk
	return chiaAcc, nil
}

func GenAccountBySKHex(sk string) (*Account, error) {
	skBytes, err := hex.DecodeString(sk)
	if err != nil {
		return nil, err
	}
	return GenAccountBySKBytes(skBytes)
}

func (ca *Account) GetSKHex() (string, error) {
	skBytes, err := ca.PrivateKey.MarshalBinary()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(skBytes), nil
}

func (ca *Account) GetPKHex() (string, error) {
	skBytes, err := ca.PublicKey().MarshalBinary()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(skBytes), nil
}

func (ca *Account) GetPKBytes() ([]byte, error) {
	return ca.PublicKey().MarshalBinary()
}

func (ca *Account) GetAddress(mainnet bool) (string, error) {
	prefix := PREFIX
	if !mainnet {
		prefix = TPREFIX
	}
	pkBytes, err := ca.PublicKey().MarshalBinary()
	if err != nil {
		return "", err
	}
	return puzzlehash.NewAddressFromPkBytes(pkBytes, prefix)
}

func (ca *Account) GetPuzzleHashStr() (string, error) {
	pkBytes, err := ca.PublicKey().MarshalBinary()
	if err != nil {
		return "", err
	}

	return puzzlehash.NewPuzzleHashFromPkBytes(pkBytes)
}

func (ca *Account) GetPuzzleHashBytes() ([]byte, error) {
	pkBytes, err := ca.PublicKey().MarshalBinary()
	if err != nil {
		return nil, err
	}

	return puzzlehash.NewPuzzleHashBytesFromPkBytes(pkBytes)
}

func (ca *Account) genSKFromSeed() (err error) {
	ca.PrivateKey, err = bls.KeyGenV3[bls.G1](ca.ikm)
	return err
}

func (ca *Account) Sign(msg []byte) []byte {
	pkBytes, _ := ca.GetPKBytes()
	msg = append(pkBytes, msg...)
	return bls.Sign(ca.PrivateKey, msg)
}

func AggregateSigns(msgs [][]byte) ([]byte, error) {
	return bls.Aggregate[bls.KeyG1SigG2](*new(bls.KeyG1SigG2), msgs)
}

func AggregatePubKeys(pks [][]byte) ([]byte, error) {
	return bls.Aggregate[bls.KeyG1SigG2](*new(bls.KeyG1SigG2), pks)
}

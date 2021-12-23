package stake

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/vangjvn/devchain/sdk"
	"github.com/vangjvn/devchain/types"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDeclareCandidacy             = 0x55
	ByteTxUpdateCandidacy              = 0x56
	ByteTxWithdrawCandidacy            = 0x57
	ByteTxVerifyCandidacy              = 0x58
	ByteTxActivateCandidacy            = 0x59
	ByteTxUpdateCandidacyAccount       = 0x63
	ByteTxAcceptCandidacyAccountUpdate = 0x64
	ByteTxDeactivateCandidacy          = 0x65
	TypeTxDeclareCandidacy             = "stake/declareCandidacy"
	TypeTxUpdateCandidacy              = "stake/updateCandidacy"
	TypeTxVerifyCandidacy              = "stake/verifyCandidacy"
	TypeTxWithdrawCandidacy            = "stake/withdrawCandidacy"
	TypeTxActivateCandidacy            = "stake/activateCandidacy"
	TypeTxDeactivateCandidacy          = "stake/deactivateCandidacy"
	TypeTxUpdateCandidacyAccount       = "stake/updateCandidacyAccount"
	TypeTxAcceptCandidacyAccountUpdate = "stake/acceptCandidacyAccountUpdate"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclareCandidacy{}, TypeTxDeclareCandidacy, ByteTxDeclareCandidacy)
	sdk.TxMapper.RegisterImplementation(TxUpdateCandidacy{}, TypeTxUpdateCandidacy, ByteTxUpdateCandidacy)
	sdk.TxMapper.RegisterImplementation(TxWithdrawCandidacy{}, TypeTxWithdrawCandidacy, ByteTxWithdrawCandidacy)
	sdk.TxMapper.RegisterImplementation(TxVerifyCandidacy{}, TypeTxVerifyCandidacy, ByteTxVerifyCandidacy)
	sdk.TxMapper.RegisterImplementation(TxActivateCandidacy{}, TypeTxActivateCandidacy, ByteTxActivateCandidacy)
	sdk.TxMapper.RegisterImplementation(TxDeactivateCandidacy{}, TypeTxDeactivateCandidacy, ByteTxDeactivateCandidacy)
	sdk.TxMapper.RegisterImplementation(TxUpdateCandidacyAccount{}, TypeTxUpdateCandidacyAccount, ByteTxUpdateCandidacyAccount)
	sdk.TxMapper.RegisterImplementation(TxAcceptCandidacyAccountUpdate{}, TypeTxAcceptCandidacyAccountUpdate, ByteTxAcceptCandidacyAccountUpdate)
}

//Verify interface at compile time
var _, _, _, _, _, _, _, _ sdk.TxInner = &TxDeclareCandidacy{}, &TxUpdateCandidacy{}, &TxWithdrawCandidacy{}, TxVerifyCandidacy{}, &TxActivateCandidacy{}, &TxUpdateCandidacyAccount{}, &TxAcceptCandidacyAccountUpdate{}, &TxDeactivateCandidacy{}

type TxDeclareCandidacy struct {
	PubKey      string      `json:"pub_key"`
	Description Description `json:"description"`
}

func (tx TxDeclareCandidacy) ValidateBasic() error {
	return nil
}

func NewTxDeclareCandidacy(pubKey types.PubKey, description Description) sdk.Tx {
	return TxDeclareCandidacy{
		PubKey:      types.PubKeyString(pubKey),
		Description: description,
	}.Wrap()
}

func (tx TxDeclareCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxUpdateCandidacy struct {
	PubKey      string      `json:"pub_key"`
	Description Description `json:"description"`
}

func (tx TxUpdateCandidacy) ValidateBasic() error {
	return nil
}

func NewTxUpdateCandidacy(pubKey types.PubKey, description Description) sdk.Tx {
	return TxUpdateCandidacy{
		PubKey:      types.PubKeyString(pubKey),
		Description: description,
	}.Wrap()
}

func (tx TxUpdateCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxVerifyCandidacy struct {
	CandidateAddress common.Address `json:"candidate_address"`
	Verified         bool           `json:"verified"`
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxVerifyCandidacy) ValidateBasic() error {
	return nil
}

func NewTxVerifyCandidacy(candidateAddress common.Address, verified bool) sdk.Tx {
	return TxVerifyCandidacy{
		CandidateAddress: candidateAddress,
		Verified:         verified,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxVerifyCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxWithdrawCandidacy struct{}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxWithdrawCandidacy) ValidateBasic() error {
	return nil
}

func NewTxWithdrawCandidacy() sdk.Tx {
	return TxWithdrawCandidacy{}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxWithdrawCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxActivateCandidacy struct{}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxActivateCandidacy) ValidateBasic() error {
	return nil
}

func NewTxActivateCandidacy() sdk.Tx {
	return TxActivateCandidacy{}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxActivateCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxDeactivateCandidacy struct{}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxDeactivateCandidacy) ValidateBasic() error {
	return nil
}

func NewTxDeactivateCandidacy() sdk.Tx {
	return TxDeactivateCandidacy{}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxDeactivateCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxUpdateCandidacyAccount struct {
	NewCandidateAddress common.Address `json:"new_candidate_account"`
}

func (tx TxUpdateCandidacyAccount) ValidateBasic() error {
	return nil
}

func NewTxUpdateCandidacyAccount(newCandidateAddress common.Address) sdk.Tx {
	return TxUpdateCandidacyAccount{
		NewCandidateAddress: newCandidateAddress,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxUpdateCandidacyAccount) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxAcceptCandidacyAccountUpdate struct {
	AccountUpdateRequestId int64 `json:"account_update_request_id"`
}

func (tx TxAcceptCandidacyAccountUpdate) ValidateBasic() error {
	return nil
}

func NewTxAcceptCandidacyAccountUpdate(accountUpdateRequestId int64) sdk.Tx {
	return TxAcceptCandidacyAccountUpdate{
		accountUpdateRequestId,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxAcceptCandidacyAccountUpdate) Wrap() sdk.Tx { return sdk.Tx{tx} }

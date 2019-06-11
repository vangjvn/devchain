package stake

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	ethstat "github.com/ethereum/go-ethereum/core/state"

	"github.com/second-state/devchain/commons"
	"github.com/second-state/devchain/sdk"
	"github.com/second-state/devchain/sdk/errors"
	"github.com/second-state/devchain/sdk/state"
	"github.com/second-state/devchain/types"
	"github.com/second-state/devchain/utils"
)

// DelegatedProofOfStake - interface to enforce delegation stake
type delegatedProofOfStake interface {
	declareCandidacy(TxDeclareCandidacy, sdk.Int) error
	updateCandidacy(TxUpdateCandidacy, sdk.Int) error
	withdrawCandidacy(TxWithdrawCandidacy) error
	verifyCandidacy(TxVerifyCandidacy) error
	activateCandidacy(TxActivateCandidacy) error
	deactivateCandidacy(TxDeactivateCandidacy) error
	updateCandidateAccount(TxUpdateCandidacyAccount, sdk.Int) (int64, error)
	acceptCandidateAccountUpdateRequest(TxAcceptCandidacyAccountUpdate, sdk.Int) error
}

func SetGenesisValidator(val types.GenesisValidator, store state.SimpleDB) error {
	if val.Address == "0000000000000000000000000000000000000000" {
		return ErrBadValidatorAddr()
	}

	addr := common.HexToAddress(val.Address)

	// create and save the empty candidate
	bond := GetCandidateByAddress(addr)
	if bond != nil {
		return ErrCandidateExistsAddr()
	}

	params := utils.GetParams()
	deliverer := deliver{
		store:  store,
		sender: addr,
		params: params,
		ctx:    types.NewContext("", 0, 0, nil),
	}

	desc := Description{
		Name:     val.Name,
		Website:  val.Website,
		Location: val.Location,
		Email:    val.Email,
		Profile:  val.Profile,
	}

	tx := TxDeclareCandidacy{types.PubKeyString(val.PubKey), desc}
	return deliverer.declareGenesisCandidacy(tx, val)
}

// CheckTx checks if the tx is properly structured
func CheckTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	// get the sender
	sender, err := getTxSender(ctx)
	if err != nil {
		return res, err
	}

	params := utils.GetParams()
	checker := check{
		store:  store,
		sender: sender,
		params: params,
		ctx:    ctx,
	}

	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		gasFee := utils.CalGasFee(params.DeclareCandidacyGas, params.GasPrice)
		return res, checker.declareCandidacy(txInner, gasFee)
	case TxUpdateCandidacy:
		gasFee := utils.CalGasFee(params.UpdateCandidacyGas, params.GasPrice)
		return res, checker.updateCandidacy(txInner, gasFee)
	case TxWithdrawCandidacy:
		return res, checker.withdrawCandidacy(txInner)
	case TxVerifyCandidacy:
		return res, checker.verifyCandidacy(txInner)
	case TxActivateCandidacy:
		return res, checker.activateCandidacy(txInner)
	case TxDeactivateCandidacy:
		return res, checker.deactivateCandidacy(txInner)
	case TxUpdateCandidacyAccount:
		gasFee := utils.CalGasFee(params.UpdateCandidateAccountGas, params.GasPrice)
		_, err := checker.updateCandidateAccount(txInner, gasFee)
		return res, err
	case TxAcceptCandidacyAccountUpdate:
		gasFee := utils.CalGasFee(params.AcceptCandidateAccountUpdateRequestGas, params.GasPrice)
		return res, checker.acceptCandidateAccountUpdateRequest(txInner, gasFee)
	}

	return res, errors.ErrUnknownTxType(tx)
}

// DeliverTx executes the tx if valid
func DeliverTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx, hash []byte) (res sdk.DeliverResult, err error) {
	_, err = CheckTx(ctx, store, tx)
	if err != nil {
		return
	}

	sender, err := getTxSender(ctx)
	if err != nil {
		return
	}

	params := utils.GetParams()
	deliverer := deliver{
		store:  store,
		sender: sender,
		params: params,
		ctx:    ctx,
	}
	res.GasFee = big.NewInt(0)

	// Run the transaction
	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		gasFee := utils.CalGasFee(params.DeclareCandidacyGas, params.GasPrice)
		err := deliverer.declareCandidacy(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.DeclareCandidacyGas)
			res.GasFee = gasFee.Int
		}
		return res, err
	case TxUpdateCandidacy:
		gasFee := utils.CalGasFee(params.UpdateCandidacyGas, params.GasPrice)
		err := deliverer.updateCandidacy(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.UpdateCandidacyGas)
			res.GasFee = gasFee.Int
		}
		return res, err
	case TxWithdrawCandidacy:
		return res, deliverer.withdrawCandidacy(txInner)
	case TxVerifyCandidacy:
		return res, deliverer.verifyCandidacy(txInner)
	case TxActivateCandidacy:
		return res, deliverer.activateCandidacy(txInner)
	case TxDeactivateCandidacy:
		return res, deliverer.deactivateCandidacy(txInner)
	case TxUpdateCandidacyAccount:
		gasFee := utils.CalGasFee(params.UpdateCandidateAccountGas, params.GasPrice)
		id, err := deliverer.updateCandidateAccount(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.UpdateCandidateAccountGas)
			res.GasFee = gasFee.Int
		}
		res.Data = []byte(strconv.Itoa(int(id)))
		return res, err
	case TxAcceptCandidacyAccountUpdate:
		gasFee := utils.CalGasFee(params.AcceptCandidateAccountUpdateRequestGas, params.GasPrice)
		err := deliverer.acceptCandidateAccountUpdateRequest(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.AcceptCandidateAccountUpdateRequestGas)
			res.GasFee = gasFee.Int
		}
		return res, err
	}

	return
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx types.Context) (sender common.Address, err error) {
	senders := ctx.GetSigners()
	if len(senders) != 1 {
		return sender, ErrMissingSignature()
	}
	return senders[0], nil
}

//_______________________________________________________________________

type check struct {
	store  state.SimpleDB
	sender common.Address
	params *utils.Params
	ctx    types.Context
}

var _ delegatedProofOfStake = check{} // enforce interface at compile time

func (c check) declareCandidacy(tx TxDeclareCandidacy, gasFee sdk.Int) error {
	pk, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	// check to see if the pubkey or address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate != nil {
		return ErrAddressAlreadyDeclared()
	}

	candidate = GetCandidateByPubKey(pk)
	if candidate != nil {
		return ErrPubKeyAleadyDeclared()
	}

	return nil
}

func (c check) updateCandidacy(tx TxUpdateCandidacy, gasFee sdk.Int) error {
	if !utils.IsBlank(tx.PubKey) {
		_, err := types.GetPubKey(tx.PubKey)
		if err != nil {
			return err
		}
	}

	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if !utils.IsBlank(tx.PubKey) {
		pk, err := types.GetPubKey(tx.PubKey)
		if err != nil {
			return err
		}

		candidate = GetCandidateByPubKey(pk)
		if candidate != nil {
			return ErrPubKeyAleadyDeclared()
		}
	}

	return nil
}

func (c check) withdrawCandidacy(tx TxWithdrawCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	return nil
}

func (c check) verifyCandidacy(tx TxVerifyCandidacy) error {
	// check to see if the candidate address to be verified has been registered before
	candidate := GetCandidateByAddress(tx.CandidateAddress)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	// check to see if the request was initiated by a special account
	if c.sender != common.HexToAddress(utils.GetParams().FoundationAddress) {
		return ErrVerificationDisallowed()
	}

	return nil
}

func (c check) activateCandidacy(tx TxActivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if candidate.Active == "Y" {
		return ErrCandidateAlreadyActivated()
	}

	return nil
}

func (c check) deactivateCandidacy(tx TxDeactivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if candidate.Active == "N" {
		return ErrCandidateAlreadyDeactivated()
	}
	return nil
}

func (c check) updateCandidateAccount(tx TxUpdateCandidacyAccount, gasFee sdk.Int) (int64, error) {
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return 0, ErrBadRequest()
	}

	tmp := GetCandidateByAddress(tx.NewCandidateAddress)
	if tmp != nil {
		return 0, ErrBadRequest()
	}

	// check if the new address has been used
	exists := getCandidateAccountUpdateRequestByToAddress(tx.NewCandidateAddress)
	if len(exists) > 0 {
		return 0, ErrBadRequest()
	}

	// check if the address has been changed
	ownerAddress := common.HexToAddress(candidate.OwnerAddress)
	if utils.IsEmptyAddress(tx.NewCandidateAddress) || tx.NewCandidateAddress == ownerAddress {
		return 0, ErrBadRequest()
	}

	// check if the candidate has sufficient funds
	if err := checkBalance(c.ctx.EthappState(), c.sender, gasFee); err != nil {
		return 0, err
	}

	return 0, nil
}

func (c check) acceptCandidateAccountUpdateRequest(tx TxAcceptCandidacyAccountUpdate, gasFee sdk.Int) error {
	req := getCandidateAccountUpdateRequestById(tx.AccountUpdateRequestId)
	if req == nil {
		return ErrBadRequest()
	}

	tmp := GetCandidateByAddress(req.ToAddress)
	if tmp != nil {
		return ErrBadRequest()
	}

	if req.ToAddress != c.sender || req.State != "PENDING" {
		return ErrBadRequest()
	}

	return nil
}

//_____________________________________________________________________

type deliver struct {
	store  state.SimpleDB
	sender common.Address
	params *utils.Params
	ctx    types.Context
}

var _ delegatedProofOfStake = deliver{} // enforce interface at compile time

// These functions assume everything has been authenticated,
// now we just perform action and save
func (d deliver) declareCandidacy(tx TxDeclareCandidacy, gasFee sdk.Int) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	now := d.ctx.BlockTime()
	candidate := &Candidate{
		PubKey:       pubKey,
		OwnerAddress: d.sender.String(),
		VotingPower:  0,
		CreatedAt:    now,
		Description:  tx.Description,
		Verified:     "N",
		Active:       "Y",
		BlockHeight:  d.ctx.BlockHeight(),
		State:        "Candidate",
	}

	// check if the validator has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, gasFee); err != nil {
		return err
	}

	SaveCandidate(candidate)
	return nil
}

func (d deliver) declareGenesisCandidacy(tx TxDeclareCandidacy, val types.GenesisValidator) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	power, _ := strconv.ParseInt(val.Power, 10, 64)
	candidate := &Candidate{
		PubKey:       pubKey,
		OwnerAddress: d.sender.String(),
		VotingPower:  power,
		CreatedAt:    0,
		Description:  tx.Description,
		Verified:     "N",
		Active:       "Y",
		BlockHeight:  d.ctx.BlockHeight(),
		State:        "Validator",
	}

	SaveCandidate(candidate)
	return nil
}

func (d deliver) updateCandidacy(tx TxUpdateCandidacy, gasFee sdk.Int) error {
	// create and save the empty candidate
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	// If other information was updated, set the verified status to false
	if len(tx.Description.Name) > 0 {
		candidate.Verified = "N"
		candidate.Description.Name = tx.Description.Name
	}
	if len(tx.Description.Email) > 0 {
		candidate.Verified = "N"
		candidate.Description.Email = tx.Description.Email
	}
	if len(tx.Description.Website) > 0 {
		candidate.Verified = "N"
		candidate.Description.Website = tx.Description.Website
	}
	if len(tx.Description.Location) > 0 {
		candidate.Verified = "N"
		candidate.Description.Location = tx.Description.Location
	}
	if len(tx.Description.Profile) > 0 {
		candidate.Verified = "N"
		candidate.Description.Profile = tx.Description.Profile
	}

	// check if the delegator has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, gasFee); err != nil {
		return err
	}

	if !utils.IsBlank(tx.PubKey) {
		newPk, _ := types.GetPubKey(tx.PubKey)

		// save the previous pubkey which will be used to update validator set
		var updates PubKeyUpdates
		tuple := PubKeyUpdate{candidate.PubKey, newPk, candidate.VotingPower}
		b := d.store.Get(utils.PubKeyUpdatesKey)
		if b == nil {
			updates = PubKeyUpdates{tuple}
		} else {
			json.Unmarshal(b, &updates)
			updates = append(updates, tuple)
		}
		b, err := json.Marshal(updates)
		if err != nil {
			panic(err)
		}

		d.store.Set(utils.PubKeyUpdatesKey, b)
	}

	updateCandidate(candidate)
	return nil
}

func (d deliver) withdrawCandidacy(tx TxWithdrawCandidacy) error {
	// create and save the empty candidate
	validatorAddress := d.sender
	candidate := GetCandidateByAddress(validatorAddress)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	candidate.Active = "N"
	updateCandidate(candidate)
	return nil
}

func (d deliver) verifyCandidacy(tx TxVerifyCandidacy) error {
	// verify candidacy
	candidate := GetCandidateByAddress(tx.CandidateAddress)
	if tx.Verified {
		candidate.Verified = "Y"
	} else {
		candidate.Verified = "N"
	}
	updateCandidate(candidate)
	return nil
}

func (d deliver) activateCandidacy(tx TxActivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	candidate.Active = "Y"
	updateCandidate(candidate)
	return nil
}

func (d deliver) deactivateCandidacy(tx TxDeactivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	candidate.Active = "N"
	updateCandidate(candidate)
	return nil
}

func (d deliver) updateCandidateAccount(tx TxUpdateCandidacyAccount, gasFee sdk.Int) (int64, error) {
	// check if the delegator has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, gasFee); err != nil {
		return 0, err
	}

	// only charge gas fee here
	d.ctx.EthappState().SubBalance(d.sender, gasFee.Int)
	d.ctx.EthappState().AddBalance(utils.HoldAccount, gasFee.Int)

	candidate := GetCandidateByAddress(d.sender)
	req := &CandidateAccountUpdateRequest{
		CandidateId: candidate.Id,
		FromAddress: d.sender, ToAddress: tx.NewCandidateAddress,
		CreatedBlockHeight: d.ctx.BlockHeight(),
		State:              "PENDING",
	}
	id := saveCandidateAccountUpdateRequest(req)
	return id, nil
}

func (d deliver) acceptCandidateAccountUpdateRequest(tx TxAcceptCandidacyAccountUpdate, gasFee sdk.Int) error {
	req := getCandidateAccountUpdateRequestById(tx.AccountUpdateRequestId)
	if req == nil {
		return ErrBadRequest()
	}

	if req.ToAddress != d.sender || req.State != "PENDING" {
		return ErrBadRequest()
	}

	candidate := GetCandidateById(req.CandidateId)
	if candidate == nil {
		return ErrBadRequest()
	}

	// check if the candidate has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, gasFee); err != nil {
		return err
	}

	candidate.OwnerAddress = req.ToAddress.String()
	updateCandidate(candidate)

	// lock coins from the new account
	//commons.Transfer(req.ToAddress, utils.HoldAccount, delegation.Shares().Add(gasFee))
	d.ctx.EthappState().SubBalance(req.ToAddress, gasFee.Int)
	d.ctx.EthappState().AddBalance(utils.HoldAccount, gasFee.Int)

	// mark the request as completed
	req.State = "COMPLETED"
	req.AcceptedBlockHeight = d.ctx.BlockHeight()
	updateCandidateAccountUpdateRequest(req)

	return nil
}

func checkBalance(state *ethstat.StateDB, addr common.Address, amount sdk.Int) error {
	balance, err := commons.GetBalance(state, addr)
	if err != nil {
		return err
	}

	if balance.LT(amount) {
		return ErrInsufficientFunds()
	}

	return nil
}

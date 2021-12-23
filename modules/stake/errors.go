// nolint
package stake

import (
	"fmt"

	"github.com/vangjvn/devchain/sdk/errors"
)

var (
	errBadAmount                          = fmt.Errorf("Amount must be > 0")
	errBadValidatorAddr                   = fmt.Errorf("Candidate does not exist for that address")
	errCandidateExistsAddr                = fmt.Errorf("Candidate already exists, cannot re-declare candidate")
	errMissingSignature                   = fmt.Errorf("Missing signature")
	errInsufficientFunds                  = fmt.Errorf("Insufficient funds")
	errCandidateVerificationDisallowed    = fmt.Errorf("Verification disallowed")
	errCandidateVerifiedAlready           = fmt.Errorf("Candidate has been verified already")
	errReachMaxAmount                     = fmt.Errorf("Validator has reached its declared max amount CMTs to be staked")
	errAddressAlreadyDeclared             = fmt.Errorf("Address has been declared")
	errPubKeyAlreadyDeclared              = fmt.Errorf("PubKey has been declared")
	errCandidateAlreadyActivated          = fmt.Errorf("Candidate has been activated")
	errCandidateAlreadyDeactivated        = fmt.Errorf("Candidate has been deactivated")
	errBadRequest                         = fmt.Errorf("Bad request")

	invalidInput = errors.CodeTypeBaseInvalidInput
)

func ErrBadValidatorAddr() error {
	return errors.WithCode(errBadValidatorAddr, errors.CodeTypeBaseUnknownAddress)
}
func ErrCandidateExistsAddr() error {
	return errors.WithCode(errCandidateExistsAddr, errors.CodeTypeBaseInvalidInput)
}

func ErrMissingSignature() error {
	return errors.WithCode(errMissingSignature, errors.CodeTypeUnauthorized)
}

func ErrInsufficientFunds() error {
	return errors.WithCode(errInsufficientFunds, errors.CodeTypeBaseInvalidInput)
}

func ErrBadAmount() error {
	return errors.WithCode(errBadAmount, errors.CodeTypeBaseInvalidOutput)
}

func ErrVerificationDisallowed() error {
	return errors.WithCode(errCandidateVerificationDisallowed, errors.CodeTypeBaseInvalidOutput)
}

func ErrBadRequest() error {
	return errors.WithCode(errBadRequest, errors.CodeTypeBaseInvalidOutput)
}

func ErrAddressAlreadyDeclared() error {
	return errors.WithCode(errAddressAlreadyDeclared, errors.CodeTypeBaseInvalidOutput)
}

func ErrPubKeyAleadyDeclared() error {
	return errors.WithCode(errPubKeyAlreadyDeclared, errors.CodeTypeBaseInvalidOutput)
}

func ErrCandidateAlreadyActivated() error {
	return errors.WithCode(errCandidateAlreadyActivated, errors.CodeTypeBaseInvalidOutput)
}

func ErrCandidateAlreadyDeactivated() error {
	return errors.WithCode(errCandidateAlreadyDeactivated, errors.CodeTypeBaseInvalidOutput)
}

package commands

import (
	"fmt"
	"github.com/vangjvn/devchain/utils"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/common"

	"github.com/vangjvn/devchain/modules/stake"
	txcmd "github.com/vangjvn/devchain/sdk/client/commands/txs"
	"github.com/vangjvn/devchain/types"
)

/*
The stake/declare tx allows a potential validator to declare its candidacy. Signed by the validator.

* Validator address

The stake/slot/propose tx allows a potential validator to offer a slot of CMTs and corresponding ROI. It returns a tx ID. Signed by the validator.

* Validator address
* CMT amount
* Proposed ROI

The stake/slot/accept tx is used by a delegator to accept and stake CMTs for an ID. Signed by the user.

* Slot ID
* CMT amount
* Delegator address

The stake/slot/cancel tx is to cancel all remianing amounts from an unaccepted slot by its creator using the ID. Signed by the validator.

* Slot ID
* Validator address
*/

// nolint
const (
	FlagPubKey                 = "pubkey"
	FlagAmount                 = "amount"
	FlagMaxAmount              = "max-amount"
	FlagCompRate               = "comp-rate"
	FlagAddress                = "address"
	FlagCandidateAddress       = "candidate-address"
	FlagName                   = "name"
	FlagEmail                  = "email"
	FlagWebsite                = "website"
	FlagLocation               = "location"
	FlagProfile                = "profile"
	FlagVerified               = "verified"
	FlagCubeBatch              = "cube-batch"
	FlagSig                    = "sig"
	FlagDelegatorAddress       = "delegator-address"
	FlagNewCandidateAddress    = "new-candidate-address"
	FlagAccountUpdateRequestId = "account-update-request-id"
)

// nolint
var (
	CmdDeclareCandidacy = &cobra.Command{
		Use:   "declare-candidacy",
		Short: "Allows a potential validator to declare its candidacy",
		RunE:  cmdDeclareCandidacy,
	}
	CmdUpdateCandidacy = &cobra.Command{
		Use:   "update-candidacy",
		Short: "Allows a validator candidate to change its candidacy",
		RunE:  cmdUpdateCandidacy,
	}
	CmdWithdrawCandidacy = &cobra.Command{
		Use:   "withdraw-candidacy",
		Short: "Allows a validator/candidate to withdraw",
		RunE:  cmdWithdrawCandidacy,
	}
	CmdVerifyCandidacy = &cobra.Command{
		Use:   "verify-candidacy",
		Short: "Allows the foundation to verify a validator/candidate's information",
		RunE:  cmdVerifyCandidacy,
	}
	CmdActivateCandidacy = &cobra.Command{
		Use:   "activate-candidacy",
		Short: "Allows a validator to activate itself",
		RunE:  cmdActivateCandidacy,
	}
	CmdDeactivateCandidacy = &cobra.Command{
		Use:   "deactivate-candidacy",
		Short: "Allows a validator to deactivate itself",
		RunE:  cmdDeactivateCandidacy,
	}
	CmdUpdateCandidacyAccount = &cobra.Command{
		Use:   "update-candidacy-account",
		Short: "Allows a validator to update its account",
		RunE:  cmdUpdateCandidacyAccount,
	}
	CmdAcceptCandidacyAccountUpdate = &cobra.Command{
		Use:   "accept-candidacy-account-update",
		Short: "Accept the candidate's account update request and become a candidate",
		RunE:  cmdAcceptCandidacyAccountUpdate,
	}
)

func init() {

	// define the flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")

	fsAmount := flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount.String(FlagAmount, "", "Amount of CMTs")

	fsCandidate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCandidate.String(FlagMaxAmount, "", "Max amount of CMTs to be staked")
	fsCandidate.String(FlagName, "", "name")
	fsCandidate.String(FlagWebsite, "", "website")
	fsCandidate.String(FlagLocation, "", "location")
	fsCandidate.String(FlagEmail, "", "email")
	fsCandidate.String(FlagProfile, "", "profile")

	fsCompRate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCompRate.String(FlagCompRate, "0", "The compensation percentage of block awards to be distributed to the validator")

	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagAddress, "", "Account address")

	fsVerified := flag.NewFlagSet("", flag.ContinueOnError)
	fsVerified.String(FlagVerified, "false", "true or false")

	fsValidatorAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsValidatorAddress.String(FlagCandidateAddress, "", "validator address")

	fsNewValidatorAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsNewValidatorAddress.String(FlagNewCandidateAddress, "", "new validator address")

	fsCubeBatch := flag.NewFlagSet("", flag.ContinueOnError)
	fsCubeBatch.String(FlagCubeBatch, "", "cube batch number")

	fsSig := flag.NewFlagSet("", flag.ContinueOnError)
	fsSig.String(FlagSig, "", "cube signature")

	fsDelegatorAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegatorAddress.String(FlagDelegatorAddress, "", "delegator address")

	fsAccountUpdateRequestId := flag.NewFlagSet("", flag.ContinueOnError)
	fsAccountUpdateRequestId.Int64(FlagAccountUpdateRequestId, 0, "account update request ID")

	// add the flags
	CmdDeclareCandidacy.Flags().AddFlagSet(fsPk)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCandidate)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCompRate)

	CmdUpdateCandidacy.Flags().AddFlagSet(fsPk)
	CmdUpdateCandidacy.Flags().AddFlagSet(fsCandidate)
	CmdUpdateCandidacy.Flags().AddFlagSet(fsCompRate)

	CmdVerifyCandidacy.Flags().AddFlagSet(fsValidatorAddress)
	CmdVerifyCandidacy.Flags().AddFlagSet(fsVerified)

	CmdUpdateCandidacyAccount.Flags().AddFlagSet(fsNewValidatorAddress)
	CmdAcceptCandidacyAccountUpdate.Flags().AddFlagSet(fsAccountUpdateRequestId)
}

func cmdDeclareCandidacy(cmd *cobra.Command, args []string) error {
	pk, err := types.GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	description := stake.Description{
		Name:     viper.GetString(FlagName),
		Email:    viper.GetString(FlagEmail),
		Website:  viper.GetString(FlagWebsite),
		Location: viper.GetString(FlagLocation),
		Profile:  viper.GetString(FlagProfile),
	}

	tx := stake.NewTxDeclareCandidacy(pk, description)
	return txcmd.DoTx(tx)
}

func cmdUpdateCandidacy(cmd *cobra.Command, args []string) error {
	pk := types.PubKey{}
	if !utils.IsBlank(viper.GetString(FlagPubKey)) {
		tmp, err := types.GetPubKey(viper.GetString(FlagPubKey))
		if err != nil {
			return err
		}
		pk = tmp
	}

	newCandidateAddress := common.HexToAddress(viper.GetString(FlagNewCandidateAddress))
	if utils.IsBlank(newCandidateAddress.String()) {
		return fmt.Errorf("please enter new candidate address using --new-candidate-address")
	}

	description := stake.Description{
		Name:     viper.GetString(FlagName),
		Email:    viper.GetString(FlagEmail),
		Website:  viper.GetString(FlagWebsite),
		Location: viper.GetString(FlagLocation),
		Profile:  viper.GetString(FlagProfile),
	}

	tx := stake.NewTxUpdateCandidacy(pk, description)
	return txcmd.DoTx(tx)
}

func cmdWithdrawCandidacy(cmd *cobra.Command, args []string) error {
	tx := stake.NewTxWithdrawCandidacy()
	return txcmd.DoTx(tx)
}

func cmdVerifyCandidacy(cmd *cobra.Command, args []string) error {
	candidateAddress := common.HexToAddress(viper.GetString(FlagCandidateAddress))
	if candidateAddress.String() == "" {
		return fmt.Errorf("please enter candidate address using --validator-address")
	}

	verified := viper.GetBool(FlagVerified)
	tx := stake.NewTxVerifyCandidacy(candidateAddress, verified)
	return txcmd.DoTx(tx)
}

func cmdActivateCandidacy(cmd *cobra.Command, args []string) error {
	tx := stake.NewTxActivateCandidacy()
	return txcmd.DoTx(tx)
}

func cmdDeactivateCandidacy(cmd *cobra.Command, args []string) error {
	tx := stake.NewTxDeactivateCandidacy()
	return txcmd.DoTx(tx)
}

func cmdUpdateCandidacyAccount(cmd *cobra.Command, args []string) error {
	newCandidateAddress := common.HexToAddress(viper.GetString(FlagNewCandidateAddress))
	if newCandidateAddress.String() == "" {
		return fmt.Errorf("please enter new candidate address using --new-candidate-address")
	}

	tx := stake.NewTxUpdateCandidacyAccount(newCandidateAddress)
	return txcmd.DoTx(tx)
}

func cmdAcceptCandidacyAccountUpdate(cmd *cobra.Command, args []string) error {
	updateAccountRequestId := viper.GetInt64(FlagAccountUpdateRequestId)
	if updateAccountRequestId == 0 {
		return fmt.Errorf("account-update-request-id must be present")
	}

	tx := stake.NewTxAcceptCandidacyAccountUpdate(updateAccountRequestId)
	return txcmd.DoTx(tx)
}

package utils

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/second-state/devchain/sdk"
)

type Params struct {
	ProposalExpirePeriod                   uint64 `json:"proposal_expire_period" type:"uint"`
	DeclareCandidacyGas                    uint64 `json:"declare_candidacy_gas" type:"uint"`
	UpdateCandidacyGas                     uint64 `json:"update_candidacy_gas" type:"uint"`
	UpdateCandidateAccountGas              uint64 `json:"update_candidate_account_gas" type:"uint"`
	AcceptCandidateAccountUpdateRequestGas uint64 `json:"accept_candidate_account_update_request_gas" type:"uint"`
	TransferFundProposalGas                uint64 `json:"transfer_fund_proposal_gas" type:"uint"`
	ChangeParamsProposalGas                uint64 `json:"change_params_proposal_gas" type:"uint"`
	DeployLibEniProposalGas                uint64 `json:"deploy_libeni_proposal_gas" type:"uint"`
	RetireProgramProposalGas               uint64 `json:"retire_program_proposal_gas" type:"uint"`
	UpgradeProgramProposalGas              uint64 `json:"upgrade_program_proposal_gas" type:"uint"`
	GasPrice                               uint64 `json:"gas_price" type:"uint"`
	LowPriceTxGasLimit                     uint64 `json:"low_price_tx_gas_limit" type:"uint"`
	LowPriceTxSlotsCap                     int    `json:"low_price_tx_slots_cap" type:"int"`
	FoundationAddress                      string `json:"foundation_address" type:"string"`
}

func DefaultParams() *Params {
	return &Params{
		ProposalExpirePeriod:                   7 * 24 * 3600 / CommitSeconds,
		DeclareCandidacyGas:                    1e6, // gas setting for declareCandidacy
		UpdateCandidacyGas:                     1e6, // gas setting for updateCandidacy
		UpdateCandidateAccountGas:              1e6, // gas setting for UpdateCandidateAccountGas
		AcceptCandidateAccountUpdateRequestGas: 1e6, // gas setting for AcceptCandidateAccountUpdateRequestGas
		TransferFundProposalGas:                2e6,
		ChangeParamsProposalGas:                2e6,
		RetireProgramProposalGas:               2e6,
		UpgradeProgramProposalGas:              2e6,
		DeployLibEniProposalGas:                2e6,
		GasPrice:                               0,
		LowPriceTxGasLimit:                     9223372036854775807, // Maximum gas limit for low-price transaction
		LowPriceTxSlotsCap:                     2147483647,          // Maximum number of low-price transaction slots per block
		FoundationAddress:                      "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
	}
}

var (
	// Keys for store prefixes
	ParamKey            = []byte{0x01} // key for global parameters
	AwardInfosKey       = []byte{0x02} // key for award infos
	AbsentValidatorsKey = []byte{0x03} // key for absent validators
	PubKeyUpdatesKey    = []byte{0x04} // key for absent validators
	dirty               = false
	params              = new(Params)
)

// load/save the global params
func LoadParams(b []byte) {
	json.Unmarshal(b, params)
}

func UnloadParams() (b []byte) {
	b, _ = json.Marshal(*params)
	return
}

func GetParams() *Params {
	return params
}

func SetParams(p *Params) {
	params = p
}

func CleanParams() (before bool) {
	before = dirty
	dirty = false
	return
}

func SetParam(name, value string) bool {
	pv := reflect.ValueOf(params).Elem()
	top := pv.Type()
	for i := 0; i < pv.NumField(); i++ {
		fv := pv.Field(i)
		if top.Field(i).Tag.Get("json") == name {
			switch fv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if iv, err := strconv.ParseInt(value, 10, 64); err == nil {
					fv.SetInt(iv)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if iv, err := strconv.ParseUint(value, 10, 64); err == nil {
					fv.SetUint(iv)
				}
			case reflect.String:
				fv.SetString(value)
			case reflect.Bool:
				if iv, err := strconv.ParseBool(value); err == nil {
					fv.SetBool(iv)
				}
			case reflect.Struct:
				switch reflect.TypeOf(fv.Interface()).Name() {
				case "Rat":
					v := sdk.NewRat(0, 1)
					if err := json.Unmarshal([]byte("\""+value+"\""), &v); err == nil {
						fv.Set(reflect.ValueOf(v))
					}
				}
			}
			dirty = true
			return true
		}
	}

	return false
}

func CheckParamType(name, value string) bool {
	pv := reflect.ValueOf(params).Elem()
	top := pv.Type()
	for i := 0; i < pv.NumField(); i++ {
		if top.Field(i).Tag.Get("json") == name {
			switch top.Field(i).Tag.Get("type") {
			case "bool":
				if _, err := strconv.ParseBool(value); err == nil {
					return true
				}
			case "int":
				if _, err := strconv.ParseInt(value, 10, 64); err == nil {
					return true
				}
			case "uint":
				if _, err := strconv.ParseUint(value, 10, 64); err == nil {
					return true
				}
			case "float":
				if iv, err := strconv.ParseFloat(value, 64); err == nil {
					if iv > 0 {
						return true
					}
				}
			case "json":
				var s map[string]interface{}
				if err := json.Unmarshal([]byte(value), &s); err == nil {
					return true
				}
				var b []interface{}
				if err := json.Unmarshal([]byte(value), &b); err == nil {
					return true
				}
			case "string":
				return true
			case "rat":
				v := sdk.NewRat(0, 1)
				if err := json.Unmarshal([]byte("\""+value+"\""), &v); err == nil {
					return true
				}
			}
			return false
		}
	}

	return false
}

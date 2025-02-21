package sdk

import (
	"strings"

	"encoding/json"
	"github.com/vangjvn/devchain/sdk/errors"
)

const maxTxSize = 10240

// TxInner is the interface all concrete transactions should implement.
//
// It adds bindings for clean un/marhsaling of the various implementations
// both as json and binary, as well as some common functionality to move them.
//
// +gen wrapper:"Tx"
type TxInner interface {
	Wrap() Tx

	// ValidateBasic should be a stateless check and just verify that the
	// tx is properly formated (required strings not blank, signatures exist, etc.)
	// this can also be run on the client-side for better debugging before posting a tx
	ValidateBasic() error
}

// please review again after implementing "middleware"

// TxLayer provides a standard way to deal with "middleware" tx,
// That add context to an embedded tx.
type TxLayer interface {
	TxInner
	Next() Tx
}

func (t Tx) IsLayer() bool {
	_, ok := t.Unwrap().(TxLayer)
	return ok
}

func (t Tx) GetLayer() TxLayer {
	l, _ := t.Unwrap().(TxLayer)
	return l
}

// env lets us parse an envelope and just grab the type
type env struct {
	Kind string `json:"type"`
}

func (t Tx) GetKind() (string, error) {
	// render as json
	d, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	// parse json
	text := env{}
	err = json.Unmarshal(d, &text)
	if err != nil {
		return "", err
	}
	// grab the type we used in json
	return text.Kind, nil
}

func (t Tx) GetMod() (string, error) {
	kind, err := t.GetKind()
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(kind, "/", 2)
	if len(parts) != 2 {
		return "", errors.ErrUnknownTxType(t)
	}
	return parts[0], nil
}

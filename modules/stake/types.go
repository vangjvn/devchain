package stake

import (
	"bytes"
	"encoding/json"
	"github.com/vangjvn/devchain/sdk/state"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/vangjvn/devchain/types"
	"github.com/vangjvn/devchain/utils"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"golang.org/x/crypto/ripemd160"
)

//_________________________________________________________________________

// Candidate defines the total Amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// candidate, the candidate is credited with a DelegatorBond whose number of
// bond shares is based on the Amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
// NOTE if the Owner.Empty() == true then this is a candidate who has revoked candidacy
type Candidate struct {
	Id                    int64        `json:"id"`
	PubKey                types.PubKey `json:"pub_key"`                 // Pubkey of candidate
	OwnerAddress          string       `json:"owner_address"`           // Sender of BondTx - UnbondTx returns here
	VotingPower           int64        `json:"voting_power"`
	CreatedAt             int64        `json:"created_at"`
	Description           Description  `json:"description"`
	Verified              string       `json:"verified"`
	Active                string       `json:"active"`
	BlockHeight           int64        `json:"block_height"`
	State                 string       `json:"state"`
}

type Description struct {
	Name     string `json:"name"`
	Website  string `json:"website"`
	Location string `json:"location"`
	Email    string `json:"email"`
	Profile  string `json:"profile"`
}

// Validator returns a copy of the Candidate as a Validator.
// Should only be called when the Candidate qualifies as a validator.
func (c *Candidate) Validator() Validator {
	return Validator(*c)
}

func (c *Candidate) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(c, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

func (c *Candidate) CalcVotingPower() (res int64) {
	return 1000
}

func (c Candidate) IsActive() bool {
	return c.Active == "Y"
}

// Validator is one of the top Candidates
type Validator Candidate

// ABCIValidator - Get the validator from a bond value
func (v Validator) ABCIValidator() abci.Validator {
	pk := v.PubKey.PubKey.(ed25519.PubKeyEd25519)
	return abci.Validator{
		PubKey: abci.PubKey{
			Type: abci.PubKeyEd25519,
			Data: pk[:],
		},
		Power: v.VotingPower,
	}
}

//_________________________________________________________________________

type Candidates []*Candidate

var _ sort.Interface = Candidates{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (cs Candidates) Len() int      { return len(cs) }
func (cs Candidates) Swap(i, j int) { cs[i], cs[j] = cs[j], cs[i] }
func (cs Candidates) Less(i, j int) bool {
	vp1, vp2 := cs[i].VotingPower, cs[j].VotingPower
	pk1, pk2 := cs[i].PubKey.Address(), cs[j].PubKey.Address()

	//note that all ChainId and App must be the same for a group of candidates
	if vp1 != vp2 {
		return vp1 > vp2
	}
	return bytes.Compare(pk1, pk2) == -1
}

// Sort - Sort the array of bonded values
func (cs Candidates) Sort() {
	sort.Sort(cs)
}

// update the voting power and save
func (cs Candidates) updateVotingPower(updates PubKeyUpdates) Candidates {
	// update voting power
	for _, c := range cs {
		if len(updates) != 0 {
			newPk, exists, vp := updates.GetNewPubKey(c.PubKey)
			if exists && vp > 0 {
				c.PubKey = newPk
			}
		}

		if c.Active == "N" {
			c.VotingPower = 0
			c.State = "Candidate"
		} else {
			c.VotingPower = c.CalcVotingPower()
			c.State = "Validator"
		}
		updateCandidate(c)
	}

	cs.Sort()

	return cs
}

// Validators - get the most recent updated validator set from the
// Candidates. These bonds are already sorted by VotingPower from
// the UpdateVotingPower function which is the only function which
// is to modify the VotingPower
func (cs Candidates) Validators() Validators {
	cs.Sort()

	//test if empty
	if len(cs) == 1 {
		if cs[0].VotingPower == 0 {
			return nil
		}
	}

	validators := make(Validators, len(cs))
	for i, c := range cs {
		if c.VotingPower == 0 { //exit as soon as the first Voting power set to zero is found
			return validators[:i]
		}
		validators[i] = c.Validator()
	}

	return validators
}

//_________________________________________________________________________

// Validators - list of Validators
type Validators []Validator

var _ sort.Interface = Validators{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (vs Validators) Len() int      { return len(vs) }
func (vs Validators) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }
func (vs Validators) Less(i, j int) bool {
	pk1, pk2 := vs[i].PubKey, vs[j].PubKey
	return bytes.Compare(pk1.Address(), pk2.Address()) == -1
}

// Sort - Sort validators by pubkey
func (vs Validators) Sort() {
	sort.Sort(vs)
}

// determine all changed validators between two validator sets
func (vs Validators) validatorsChanged(vs2 Validators) (changed []abci.Validator) {

	//first sort the validator sets
	vs.Sort()
	vs2.Sort()

	max := len(vs) + len(vs2)
	changed = make([]abci.Validator, max)
	i, j, n := 0, 0, 0 //counters for vs loop, vs2 loop, changed element

	for i < len(vs) && j < len(vs2) {
		if bytes.Compare(vs[i].PubKey.Address(), vs2[j].PubKey.Address()) != 0 {
			// pk1 > pk2, a new validator was introduced between these pubkeys
			if bytes.Compare(vs[i].PubKey.Address(), vs2[j].PubKey.Address()) == 1 {
				changed[n] = vs2[j].ABCIValidator()
				n++
				j++
				continue
			} // else, the old validator has been removed
			pk := vs[i].PubKey.PubKey.(ed25519.PubKeyEd25519)
			changed[n] = abci.Ed25519Validator(pk[:], 0)
			n++
			i++
			continue
		}
		if vs[i].VotingPower != vs2[j].VotingPower {
			changed[n] = vs2[j].ABCIValidator()
			n++
		}
		j++
		i++
	}

	// add any excess validators in set 2
	for ; j < len(vs2); j, n = j+1, n+1 {
		changed[n] = vs2[j].ABCIValidator()
	}

	// remove any excess validators left in set 1
	for ; i < len(vs); i, n = i+1, n+1 {
		pk := vs[i].PubKey.PubKey.(ed25519.PubKeyEd25519)
		changed[n] = abci.Ed25519Validator(pk[:], 0)
	}

	return changed[:n]
}

func (vs Validators) Remove(i int) Validators {
	copy(vs[i:], vs[i+1:])
	return vs[:len(vs)-1]
}

// UpdateValidatorSet - Updates the voting power for the candidate set and
// returns the subset of validators which have changed for Tendermint
func UpdateValidatorSet(store state.SimpleDB) (change []abci.Validator, err error) {
	// get the validators before update
	candidates := GetCandidates()
	v1 := candidates.Validators()

	// check if there are any pubkeys need to update
	var updates PubKeyUpdates
	b := store.Get(utils.PubKeyUpdatesKey)
	if b != nil {
		json.Unmarshal(b, &updates)
	}

	v2 := candidates.updateVotingPower(updates).Validators()
	change = v1.validatorsChanged(v2)

	if len(updates) != 0 {
		for _, c := range candidates {
			newPk, exists, vp := updates.GetNewPubKey(c.PubKey)
			if exists && vp == 0 {
				c.PubKey = newPk
				updateCandidate(c)
			}
		}
		store.Remove(utils.PubKeyUpdatesKey)
	}

	return
}

// Deactivate the validators
func (vs Validators) Deactivate() {
	// update voting power
	for _, v := range vs {
		v.Active = "N"
		v.VotingPower = 0
		c := Candidate(v)
		updateCandidate(&c)
	}
}

func (vs Validators) Contains(pk types.PubKey) bool {
	for _, v := range vs {
		if v.PubKey == pk {
			return true
		}
	}
	return false
}

type CandidateAccountUpdateRequest struct {
	Id                  int64          `json:"id"`
	CandidateId         int64          `json:"candidate_id"`
	FromAddress         common.Address `json:"from_address"`
	ToAddress           common.Address `json:"to_address"`
	CreatedBlockHeight  int64          `json:"created_block_height"`
	AcceptedBlockHeight int64          `json:"accepted_block_height"`
	State               string         `json:"state"`
}

func (c *CandidateAccountUpdateRequest) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(c, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

type PubKeyUpdate struct {
	OldPubKey   types.PubKey `json:"old_pub_key"`
	NewPubKey   types.PubKey `json:"new_pub_key"`
	VotingPower int64        `json:"voting_power"`
}

type PubKeyUpdates []PubKeyUpdate

func (tuples PubKeyUpdates) GetNewPubKey(pk types.PubKey) (res types.PubKey, exists bool, votingPower int64) {
	exists = false
	for _, tuple := range tuples {
		if bytes.Compare(tuple.OldPubKey.Bytes(), pk.Bytes()) == 0 {
			return tuple.NewPubKey, true, tuple.VotingPower
		}
	}

	return types.PubKey{}, exists, 0
}

package stake

import (
	"database/sql"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vangjvn/devchain/sdk/dbm"
	"github.com/vangjvn/devchain/types"
)

var (
	deliverSqlTx *sql.Tx
)

func SetDeliverSqlTx(tx *sql.Tx) {
	deliverSqlTx = tx
}

func ResetDeliverSqlTx() {
	deliverSqlTx = nil
}

func getDb() *sql.DB {
	db, err := dbm.Sqliter.GetDB()
	if err != nil {
		panic(err)
	}
	return db
}

type SqlTxWrapper struct {
	tx        *sql.Tx
	withBlock bool
}

func getSqlTxWrapper() *SqlTxWrapper {
	var wrapper = &SqlTxWrapper{
		tx:        deliverSqlTx,
		withBlock: true,
	}
	if wrapper.tx == nil {
		db := getDb()
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		wrapper.tx = tx
		wrapper.withBlock = false
	}
	return wrapper
}

func (wrapper *SqlTxWrapper) Commit() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Commit(); err != nil {
			panic(err)
		}
	}
}

func (wrapper *SqlTxWrapper) Rollback() {
	if !wrapper.withBlock {
		if err := wrapper.tx.Rollback(); err != nil {
			panic(err)
		}
	}
}

func buildQueryClause(cond map[string]interface{}) (clause string, params []interface{}) {
	if cond == nil || len(cond) == 0 {
		return "", nil
	}

	clause = ""
	for k, v := range cond {
		s := fmt.Sprintf("%s = ?", k)

		if len(clause) == 0 {
			clause = s
		} else {
			clause = fmt.Sprintf("%s and %s", clause, s)
		}
		params = append(params, v)
	}

	if len(clause) != 0 {
		clause = fmt.Sprintf(" where %s", clause)
	}

	return
}

func GetCandidateById(id int64) *Candidate {
	cond := make(map[string]interface{})
	cond["id"] = id
	candidates := getCandidatesInternal(cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func GetCandidateByAddress(address common.Address) *Candidate {
	cond := make(map[string]interface{})
	cond["address"] = address.String()
	candidates := getCandidatesInternal(cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func GetCandidateByPubKey(pubKey types.PubKey) *Candidate {
	cond := make(map[string]interface{})
	cond["pub_key"] = types.PubKeyString(pubKey)
	candidates := getCandidatesInternal(cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func GetCandidates() (candidates Candidates) {
	cond := make(map[string]interface{})
	candidates = getCandidatesInternal(cond)
	return candidates
}

func getCandidatesInternal(cond map[string]interface{}) (candidates Candidates) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	clause, params := buildQueryClause(cond)
	rows, err := txWrapper.tx.Query("select id, pub_key, address, voting_power, name, website, location, profile, email, verified, active, block_height, state, created_at from candidates"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	candidates = composeCandidateResults(rows)
	return
}

func composeCandidateResults(rows *sql.Rows) (candidates Candidates) {
	for rows.Next() {
		var pubKey, address, name, website, location, profile, email, state, verified, active string
		var id, votingPower, blockHeight, createdAt int64
		err := rows.Scan(&id, &pubKey, &address, &votingPower, &name, &website, &location, &profile, &email, &verified, &active, &blockHeight, &state, &createdAt)
		if err != nil {
			panic(err)
		}
		pk, _ := types.GetPubKey(pubKey)
		description := Description{
			Name:     name,
			Website:  website,
			Location: location,
			Profile:  profile,
			Email:    email,
		}
		candidate := &Candidate{
			Id:           id,
			PubKey:       pk,
			OwnerAddress: address,
			VotingPower:  votingPower,
			Description:  description,
			Verified:     verified,
			CreatedAt:    createdAt,
			Active:       active,
			BlockHeight:  blockHeight,
			State:        state,
		}
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func SaveCandidate(candidate *Candidate) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into candidates(pub_key, address, voting_power, name, website, location, profile, email, verified, active, hash, block_height, state, created_at) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		types.PubKeyString(candidate.PubKey),
		candidate.OwnerAddress,
		candidate.VotingPower,
		candidate.Description.Name,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Profile,
		candidate.Description.Email,
		candidate.Verified,
		candidate.Active,
		common.Bytes2Hex(candidate.Hash()),
		candidate.BlockHeight,
		candidate.State,
		candidate.CreatedAt,
	)
	if err != nil {
		panic(err)
	}
}

func updateCandidate(candidate *Candidate) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update candidates set address = ?, voting_power = ?, name =?, website = ?, location = ?, profile = ?, email = ?, verified = ?, active = ?, hash = ?, state = ?, pub_key = ? where id = ?")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		candidate.OwnerAddress,
		candidate.VotingPower,
		candidate.Description.Name,
		candidate.Description.Website,
		candidate.Description.Location,
		candidate.Description.Profile,
		candidate.Description.Email,
		candidate.Verified,
		candidate.Active,
		common.Bytes2Hex(candidate.Hash()),
		candidate.State,
		types.PubKeyString(candidate.PubKey),
		candidate.Id,
	)
	if err != nil {
		panic(err)
	}
}

func saveCandidateAccountUpdateRequest(req *CandidateAccountUpdateRequest) int64 {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("insert into candidate_account_update_requests(candidate_id, from_address, to_address, created_block_height, accepted_block_height, state, hash) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		req.CandidateId,
		req.FromAddress.String(),
		req.ToAddress.String(),
		req.CreatedBlockHeight,
		req.AcceptedBlockHeight,
		req.State,
		common.Bytes2Hex(req.Hash()),
	)
	if err != nil {
		panic(err)
	}

	lastInsertId, _ := result.LastInsertId()
	return lastInsertId
}

func getCandidateAccountUpdateRequestById(id int64) *CandidateAccountUpdateRequest {
	cond := make(map[string]interface{})
	cond["id"] = id
	reqs := getCandidateAccountUpdateRequestInternal(cond)

	if len(reqs) == 0 {
		return nil
	} else {
		return reqs[0]
	}
}

func getCandidateAccountUpdateRequestByToAddress(toAddress common.Address) (res []*CandidateAccountUpdateRequest) {
	cond := make(map[string]interface{})
	cond["to_address"] = toAddress.String()
	res = getCandidateAccountUpdateRequestInternal(cond)
	return
}

func getCandidateAccountUpdateRequestInternal(cond map[string]interface{}) (reqs []*CandidateAccountUpdateRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	clause, params := buildQueryClause(cond)
	rows, err := txWrapper.tx.Query("select id, candidate_id, from_address, to_address, created_block_height, accepted_block_height, state from candidate_account_update_requests"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	reqs = composeCandidateAccountUpdateRequestResults(rows)
	return
}

func composeCandidateAccountUpdateRequestResults(rows *sql.Rows) (reqs []*CandidateAccountUpdateRequest) {
	for rows.Next() {
		var id, candidateId, createdBlockHeight, acceptedBlockHeight int64
		var fromAddress, toAddress, state string
		err := rows.Scan(&id, &candidateId, &fromAddress, &toAddress, &createdBlockHeight, &acceptedBlockHeight, &state)
		if err != nil {
			return nil
		}

		req := &CandidateAccountUpdateRequest{
			Id:                  id,
			CandidateId:         candidateId,
			FromAddress:         common.HexToAddress(fromAddress),
			ToAddress:           common.HexToAddress(toAddress),
			CreatedBlockHeight:  createdBlockHeight,
			AcceptedBlockHeight: acceptedBlockHeight,
			State:               state,
		}
		reqs = append(reqs, req)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
	return
}

func updateCandidateAccountUpdateRequest(req *CandidateAccountUpdateRequest) {
	txWrapper := getSqlTxWrapper()
	defer txWrapper.Commit()

	stmt, err := txWrapper.tx.Prepare("update candidate_account_update_requests set accepted_block_height = ?, state = ?, hash = ? where id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.AcceptedBlockHeight,
		req.State,
		common.Bytes2Hex(req.Hash()),
		req.Id,
	)
	if err != nil {
		panic(err)
	}
}

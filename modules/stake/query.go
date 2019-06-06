package stake

import (
	"database/sql"
	"github.com/ethereum/go-ethereum/common"
)

func QueryCandidates() (candidates Candidates) {
	db := getDb()
	cond := make(map[string]interface{})
	cond["active"] = "Y"
	return queryCandidates(db, cond)
}

func QueryCandidateByAddress(address common.Address) *Candidate {
	db := getDb()
	cond := make(map[string]interface{})
	cond["address"] = address.String()
	candidates := queryCandidates(db, cond)
	if len(candidates) == 0 {
		return nil
	} else {
		return candidates[0]
	}
}

func queryCandidates(db *sql.DB, cond map[string]interface{}) (candidates Candidates) {
	clause, params := buildQueryClause(cond)
	rows, err := db.Query("select id, pub_key, address, voting_power, name, website, location, profile, email, verified, active, block_height, state, created_at from candidates"+clause, params...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	candidates = composeCandidateResults(rows)
	return
}

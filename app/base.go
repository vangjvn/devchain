package app

import (
	"bytes"
	"database/sql"
	goerr "errors"
	"math/big"
	"strings"

	"github.com/second-state/devchain/modules/governance"
	"github.com/second-state/devchain/modules/stake"
	"github.com/second-state/devchain/sdk"
	"github.com/second-state/devchain/sdk/dbm"
	"github.com/second-state/devchain/sdk/errors"
	"github.com/second-state/devchain/sdk/state"
	"github.com/second-state/devchain/server"
	ttypes "github.com/second-state/devchain/types"
	"github.com/second-state/devchain/utils"
	"github.com/second-state/devchain/version"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"golang.org/x/crypto/ripemd160"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	EthApp       *EthermintApplication
	checkedTx    map[common.Hash]*types.Transaction
	ethereum     *eth.Ethereum
	blockTime    int64
	deliverSqlTx *sql.Tx
	proposer     abci.Validator
}

var (
	_            abci.Application = &BaseApp{}
	toBeShutdown                  = false
)

// NewBaseApp extends a StoreApp with a handler and a ticker,
// which it binds to the proper abci calls
func NewBaseApp(store *StoreApp, ethApp *EthermintApplication, ethereum *eth.Ethereum) (*BaseApp, error) {
	// init pending proposals
	pendingProposals := governance.GetPendingProposals()
	if len(pendingProposals) > 0 {
		proposalsTS := make(map[string]int64)
		proposalsBH := make(map[string]int64)
		for _, pp := range pendingProposals {
			if pp.ExpireTimestamp > 0 {
				proposalsTS[pp.Id] = pp.ExpireTimestamp
			} else {
				proposalsBH[pp.Id] = pp.ExpireBlockHeight
			}

			if pp.Type == governance.DEPLOY_LIBENI_PROPOSAL {
				dp := governance.GetProposalById(pp.Id)
				if dp.Detail["status"] != "ready" {
					governance.DownloadLibEni(dp)
				}
			}
		}
		utils.PendingProposal.BatchAddTS(proposalsTS)
		utils.PendingProposal.BatchAddBH(proposalsBH)
	}

	b := store.Append().Get(utils.ParamKey)
	if b != nil {
		utils.LoadParams(b)
	}

	app := &BaseApp{
		StoreApp:  store,
		EthApp:    ethApp,
		checkedTx: make(map[common.Hash]*types.Transaction),
		ethereum:  ethereum,
	}
	return app, nil
}

// InitChain - ABCI
func (app *StoreApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	return
}

// Info implements abci.Application. It returns the height and hash,
// as well as the abci name and version.
//
// The height is the block that holds the transactions, not the apphash itself.
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	ethInfoRes := app.EthApp.Info(req)

	lbh := ethInfoRes.LastBlockHeight

	if big.NewInt(lbh).Cmp(bigZero) == 0 {
		return ethInfoRes
	}

	rp := governance.GetRetiringProposal(version.Version)
	if rp != nil {
		if rp.ExpireBlockHeight <= lbh {
			rp = governance.GetProposalById(rp.Id)
			if rp.Detail["status"] == "success" {
				server.StopFlag <- true
			}
		} else if rp.ExpireBlockHeight == lbh+1 {
			if rp.Result == "Approved" {
				utils.RetiringProposalId = rp.Id
			}
		} else {
			// check ahead one block
			utils.PendingProposal.Add(rp.Id, 0, rp.ExpireBlockHeight-1)
		}
	}

	travisInfoRes := app.StoreApp.Info(req)

	// If the chain has just relaunched from a retired version,
	// then use the old algorithm to match the old hash
	var travisDbHash []byte
	if governance.GetLatestRetiredHeight() == lbh {
		travisDbHash = app.StoreApp.GetOldDbHash()
	} else {
		travisDbHash = app.StoreApp.GetDbHash()
	}

	travisInfoRes.LastBlockAppHash = finalAppHash(ethInfoRes.LastBlockAppHash, travisInfoRes.LastBlockAppHash, travisDbHash, travisInfoRes.LastBlockHeight, nil)
	return travisInfoRes
}

// DeliverTx - ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Error("DeliverTx: Received invalid transaction", "err", err)
		return errors.DeliverResult(err)
	}

	if utils.IsEthTx(tx) {
		if checkedTx, ok := app.checkedTx[tx.Hash()]; ok {
			tx = checkedTx
		} else {
			// force cache from of tx
			networkId := big.NewInt(int64(app.ethereum.NetVersion()))
			signer := types.NewEIP155Signer(networkId)

			if _, err := types.Sender(signer, tx); err != nil {
				app.logger.Debug("DeliverTx: Received invalid transaction", "tx", tx, "err", err)
				return errors.DeliverResult(err)
			}
		}
		resp := app.EthApp.DeliverTx(tx)
		app.logger.Debug("EthApp DeliverTx response", "resp", resp)
		return resp
	}

	app.logger.Info("DeliverTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.blockTime, app.EthApp.DeliverTxState())
	return app.deliverHandler(ctx, app.Append(), tx)
}

// CheckTx - ABCI
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Error("CheckTx: Received invalid transaction", "err", err)
		return errors.CheckResult(err)
	}

	if utils.IsEthTx(tx) {
		resp := app.EthApp.CheckTx(tx)
		app.logger.Debug("EthApp CheckTx response", "resp", resp)
		if resp.IsErr() {
			return errors.CheckResult(goerr.New(resp.String()))
		}
		app.checkedTx[tx.Hash()] = tx
		return sdk.NewCheck(0, "").ToABCI()
	}

	app.logger.Info("CheckTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.blockTime, app.EthApp.checkTxState)
	return app.checkHandler(ctx, app.Check(), tx)
}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.blockTime = req.GetHeader().Time
	app.EthApp.BeginBlock(req)

	// init deliver sql tx for statke
	db, err := dbm.Sqliter.GetDB()
	if err != nil {
		panic(err)
	}
	deliverSqlTx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	app.deliverSqlTx = deliverSqlTx
	stake.SetDeliverSqlTx(deliverSqlTx)
	governance.SetDeliverSqlTx(deliverSqlTx)
	// init end

	app.proposer = req.Header.Proposer

	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	app.EthApp.EndBlock(req)
	utils.BlockGasFee = big.NewInt(0).Add(utils.BlockGasFee, app.TotalUsedGasFee)

	// Deactivate validators that not in the list of preserved validators
	if utils.RetiringProposalId != "" {
		if proposal := governance.GetProposalById(utils.RetiringProposalId); proposal != nil {
			pks := strings.Split(proposal.Detail["preserved_validators"].(string), ",")
			vs := stake.GetCandidates().Validators()
			inaVs := make(stake.Validators, 0)
			abciVs := make([]abci.Validator, 0)
			pvSize := 0
			for _, v := range vs {
				i := 0
				for ; i < len(pks); i++ {
					if pks[i] == ttypes.PubKeyString(v.PubKey) {
						v.VotingPower = 1000
						abciVs = append(abciVs, v.ABCIValidator())
						pvSize++
						break
					}
				}
				if i == len(pks) {
					inaVs = append(inaVs, v)
					pk := v.PubKey.PubKey.(ed25519.PubKeyEd25519)
					abciVs = append(abciVs, abci.Ed25519Validator(pk[:], 0))
				}
			}
			if pvSize >= 1 {
				inaVs.Deactivate()
				app.AddValChange(abciVs)
				toBeShutdown = true
				governance.UpdateRetireProgramStatus(utils.RetiringProposalId, "success")
			} else {
				governance.UpdateRetireProgramStatus(utils.RetiringProposalId, "rejected")
			}
		} else {
			app.logger.Error("Getting invalid RetiringProposalId")
		}
	}

	if !toBeShutdown { // should not update validator set twice if the node is to be shutdown
		// calculate the validator set difference
		diff, err := stake.UpdateValidatorSet(app.Append())
		if err != nil {
			panic(err)
		}
		app.AddValChange(diff)
	}

	return app.StoreApp.EndBlock(req)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	if toBeShutdown {
		server.StopFlag <- true
	}

	app.checkedTx = make(map[common.Hash]*types.Transaction)
	ethAppCommit, err := app.EthApp.Commit()
	if err != nil {
		// Rollback transaction
		if app.deliverSqlTx != nil {
			err := app.deliverSqlTx.Rollback()
			if err != nil {
				panic(err)
			}
			stake.ResetDeliverSqlTx()
			governance.ResetDeliverSqlTx()
		}
	} else {
		if app.deliverSqlTx != nil {
			// Commit transaction
			err := app.deliverSqlTx.Commit()
			if err != nil {
				panic(err)
			}
			stake.ResetDeliverSqlTx()
			governance.ResetDeliverSqlTx()
		}
	}

	workingHeight := app.WorkingHeight()

	if dirty := utils.CleanParams(); workingHeight == 1 || dirty {
		state := app.Append()
		state.Set(utils.ParamKey, utils.UnloadParams())
	}

	// reset store app
	app.TotalUsedGasFee = big.NewInt(0)

	res = app.StoreApp.Commit()
	dbHash := app.StoreApp.GetDbHash()
	res.Data = finalAppHash(ethAppCommit.Data, res.Data, dbHash, workingHeight, nil)

	return
}

func finalAppHash(ethCommitHash []byte, travisCommitHash []byte, dbHash []byte, workingHeight int64, store *state.SimpleDB) []byte {

	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	buf.Write(ethCommitHash)
	buf.Write(travisCommitHash)
	buf.Write(dbHash)
	hasher.Write(buf.Bytes())
	hash := hasher.Sum(nil)
	return hash
}

const expect = require("chai").expect
const async = require("async")
const http = require("http")
const logger = require("./logger")
const { Settings } = require("./constants")
const Globals = require("./global_vars")

const transfer = (f, t, v, gasPrice, nonce) => {
  let payload = {
    from: f,
    to: t,
    value: v,
    gasPrice: gasPrice || 0
  }
  if (nonce) payload.nonce = nonce
  let hash = null
  try {
    hash = web3.cmt.sendTransaction(payload)
    logger.debug(`transfer ${v} wei from ${f} to ${t}, hash: ${hash}`)
    // check hash
    expect(hash).to.not.empty
  } catch (err) {
    logger.error(err.message)
  }
  return hash
}

const getBalance = (index = null) => {
  let balance = new Array(4)
  for (i = 0; i < 4; i++) {
    if (index === null || i == index) {
      balance[i] = web3.cmt.getBalance(Globals.Accounts[i], "latest")
    }
  }
  balance[4] = web3.cmt.getBalance(web3.cmt.defaultAccount, "latest")
  logger.debug(`balance in wei: --> ${balance}`)
  return index == null ? balance : balance[index]
}

const newContract = function(deployAddress, abi, bytecode, cb) {
  let tokenContract = web3.cmt.contract(abi)
  let contractInstance = tokenContract.new(
    {
      from: deployAddress,
      data: bytecode,
      gas: "4700000"
    },
    function(e, contract) {
      if (e) throw e
      if (typeof contract.address !== "undefined") {
        logger.debug(
          "Contract mined! address: " +
            contract.address +
            " transactionHash: " +
            contract.transactionHash
        )
        expect(contract.address).to.not.empty
        cb(contract.address)
      }
    }
  )
  return contractInstance
}

const tokenTransfer = function(f, t, v, gasPrice, nonce) {
  let tokenContract = web3.cmt.contract(Globals.ETH.abi)
  let tokenInstance = tokenContract.at(Globals.ETH.contractAddress)
  let option = {
    from: f,
    gasPrice: gasPrice || 0
  }
  if (nonce) option.nonce = nonce
  let hash = null
  try {
    hash = tokenInstance.transfer.sendTransaction(t, v, option)
    logger.debug("token transfer hash: ", hash)
    // check hash
    expect(hash).to.not.empty
  } catch (err) {
    logger.error(err.message)
  }
  return hash
}

const tokenKill = deployAdrress => {
  let tokenContract = web3.cmt.contract(Globals.ETH.abi)
  let tokenInstance = tokenContract.at(Globals.ETH.contractAddress)
  let hash = tokenInstance.kill({ from: deployAdrress })
  logger.debug("token kill hash: ", hash)
  return hash
}

const getTokenBalance = () => {
  let tokenContract = web3.cmt.contract(Globals.ETH.abi)
  let tokenInstance = tokenContract.at(Globals.ETH.contractAddress)

  let balance = new Array(4)
  for (i = 0; i < 4; i++) {
    balance[i] = tokenInstance.balanceOf(Globals.Accounts[i])
  }
  logger.debug(`token balance: --> ${balance}`)
  return balance
}

const vote = (proposalId, from, answer) => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  web3.cmt.governance.vote(
    {
      from: from,
      proposalId: proposalId,
      answer: answer
    },
    (err, res) => {
      if (err) {
        logger.error(err.message)
      } else {
        expectTxSuccess(res)
      }
    }
  )
}

const getProposal = proposalId => {
  expect(proposalId).to.not.be.empty
  if (proposalId === "") return

  let r = web3.cmt.governance.listProposals()
  expect(r.data.length).to.be.above(0)
  if (r.data.length > 0) {
    proposal = r.data.filter(d => d.Id == proposalId)
    expect(proposal.length).to.equal(1)
    return proposal[0]
  }
  return {}
}

const waitInterval = function(txhash, cb) {
  let startingBlock = web3.cmt.blockNumber
  let startingTime = Math.round(new Date().getTime() / 1000)

  logger.debug("Starting block:", startingBlock)
  let interval = setInterval(() => {
    let blocksGone = web3.cmt.blockNumber - startingBlock
    let timeGone = Math.round(new Date().getTime() / 1000) - startingTime

    if (blocksGone > Settings.BlockTimeout) {
      clearInterval(interval)
      cb(new Error(`Pending full after ${Settings.BlockTimeout} blocks`))
      return
    }
    if (timeGone > Settings.WaitTimeout) {
      clearInterval(interval)
      logger.error(`Pending full after ${Settings.WaitTimeout} seconds`)
      process.exit(1)
    }

    let receipt = web3.cmt.getTransactionReceipt(txhash)
    logger.debug(`Blocks Passed ${blocksGone}, ${txhash} receipt: ${receipt}`)

    if (receipt != null && receipt.blockNumber > 0) {
      clearInterval(interval)
      cb(null, receipt)
    }
  }, Settings.IntervalMs || 100)
}

const waitMultiple = function(arrTxhash, cb) {
  let waitAll = arrTxhash
    .filter(e => {
      return e
    })
    .map(txhash => {
      return waitInterval.bind(null, txhash)
    })

  async.parallel(waitAll, (err, res) => {
    if (err) {
      return cb(err, res)
    }
    cb(null, res)
  })
}

const waitBlocks = (done, blocks = 1) => {
  let startingBlock = web3.cmt.blockNumber
  logger.debug("waiting start: ", startingBlock)
  let startingTime = Math.round(new Date().getTime() / 1000)
  let interval = setInterval(() => {
    let blocksGone = web3.cmt.blockNumber - startingBlock
    let timeGone = Math.round(new Date().getTime() / 1000) - startingTime
    logger.debug(`Blocks Passed ${blocksGone}`)
    if (blocksGone == blocks) {
      logger.debug("waiting end. ")
      clearInterval(interval)
      done()
    }
    if (timeGone > Settings.WaitTimeout) {
      clearInterval(interval)
      logger.error(`Pending full after ${Settings.WaitTimeout} seconds`)
      process.exit(1)
    }
  }, Settings.IntervalMs || 100)
}

const expectTxFail = (r, check_err, deliver_err) => {
  logger.debug(r)
  expect(r)
    .to.have.property("height")
    .and.to.eq(0)

  if (check_err) {
    expect(r.check_tx.code).to.eq(check_err)
  } else if (deliver_err) {
    expect(r.deliver_tx.code).to.eq(deliver_err)
  }
}

const expectTxSuccess = r => {
  logger.debug(r)
  expect(r)
    .to.have.property("height")
    .and.to.gt(0)
}

const gasFee = txType => {
  let gasPrice = web3.toBigNumber(Globals.Params.gas_price)
  let gasLimit = 0
  switch (txType) {
    case "declareCandidacy":
      gasLimit = web3.toBigNumber(Globals.Params.declare_candidacy_gas)
      break
    case "updateCandidacy":
      gasLimit = web3.toBigNumber(Globals.Params.update_candidacy_gas)
      break
    case "proposeTransferFund":
      gasLimit = web3.toBigNumber(Globals.Params.transfer_fund_proposal_gas)
      break
    case "proposeChangeParam":
      gasLimit = web3.toBigNumber(Globals.Params.change_params_proposal_gas)
      break
    case "proposeDeployLibEni":
      gasLimit = web3.toBigNumber(Globals.Params.deploy_libeni_proposal_gas)
      break
    case "updateAccount":
      gasLimit = web3.toBigNumber(Globals.Params.update_candidate_account_gas)
      break
    case "acceptAccountUpdate":
      gasLimit = web3.toBigNumber(Globals.Params.accept_candidate_account_update_request_gas)
      break
  }
  return gasPrice.times(gasLimit)
}

const addFakeValidators = () => {
  if (Globals.TestMode == "single") {
    let result = web3.cmt.stake.validator.list()
    let valsToAdd = 4 - result.data.length

    if (valsToAdd > 0) {
      Globals.Accounts.forEach((acc, idx) => {
        if (idx >= valsToAdd) return
        let payload = {
          from: acc,
          pubKey: Globals.PubKeys[idx],
        }
        let r = web3.cmt.stake.validator.declare(payload)
        logger.debug(r)
        logger.debug(`validator ${acc} added`)
      })
    }
  }
}

const removeFakeValidators = () => {
  if (Globals.TestMode == "single") {
    let result = web3.cmt.stake.validator.list()
    result.data.forEach((val, idx) => {
      // skip the first one
      if (idx == 0) return
      // remove all others
      let acc = val.owner_address
      let r = web3.cmt.stake.validator.withdraw({ from: acc })
      logger.debug(r)
      logger.debug(`validator ${acc} removed`)
    })
  }
}

const getTMValidators = cb => {
  let url = `http://${Settings.Nodes[0].domain}:26657/validators`
  http.get(url, resp => {
    let data = ""
    // A chunk of data has been recieved.
    resp.on("data", chunk => {
      data += chunk
    })
    // The whole response has been received. Print out the result.
    resp.on("end", () => {
      cb(null, JSON.parse(data))
    })
  })
}

module.exports = {
  transfer,
  getBalance,
  newContract,
  tokenTransfer,
  tokenKill,
  getTokenBalance,
  vote,
  getProposal,
  waitInterval,
  waitMultiple,
  waitBlocks,
  expectTxFail,
  expectTxSuccess,
  gasFee,
  addFakeValidators,
  removeFakeValidators,
  getTMValidators
}

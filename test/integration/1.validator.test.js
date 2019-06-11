const chai = require("chai")
const chaiSubset = require("chai-subset")
chai.use(chaiSubset)
const expect = chai.expect

const logger = require("./logger")
const Utils = require("./global_hooks")
const Globals = require("./global_vars")
const { Settings } = require("./constants")

describe("Validator Test", function() {
  let existingValidator = {}
  let tx_result
  let newAccount, newAccount2

  before(function() {
    Utils.addFakeValidators()
  })

  after(function() {
    Utils.removeFakeValidators()
  })

  before(function() {
    // get existing validator
    tx_result = web3.cmt.stake.validator.list()
    expect(tx_result.data.length).be.above(0)

    logger.debug("current validators: ", JSON.stringify(tx_result.data))
    existingValidator = tx_result.data[0]
    expect(existingValidator).be.an("object")
  })

  describe("Declare Candidacy", function() {
    it("for an existing initial validator account — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: existingValidator.owner_address,
        pubKey: Globals.PubKeys[3]
      }
      tx_result = web3.cmt.stake.validator.declare(payload)
      Utils.expectTxFail(tx_result)
    })

    it("associate to an existing validator pubkey — fail", function() {
      if (Object.keys(existingValidator).length == 0) return
      let payload = {
        from: Globals.Accounts[3],
        pubKey: existingValidator.pub_key.value
      }
      tx_result = web3.cmt.stake.validator.declare(payload)
      Utils.expectTxFail(tx_result)
    })

    it("Succeeds", function(done) {
      let payload = {
        from: Globals.Accounts[3],
        pubKey: Globals.PubKeys[3],
      }
      tx_result = web3.cmt.stake.validator.declare(payload)
      Utils.expectTxSuccess(tx_result)
      Utils.waitBlocks(done, 2)

      // balance after
      // balance_new = Utils.getBalance(3)
      // let gasFee = Utils.gasFee("declareCandidacy")
      // expect(balance_new.minus(balance_old).toNumber()).to.equal(
      //     -gasFee.plus(amounts.self).toNumber()
      // )
    })

    it("5 validators on tendermint", function(done) {
      before(function(done) {
        Utils.waitBlocks(done, 3)
      })
      Utils.getTMValidators((err, res) => {
        expect(err).to.be.null
        expect(res).to.be.not.null
        expect(res.result.validators.length).to.eq(5)
        // backup validator on tendermint
        let result = res.result.validators.filter(v => v.pub_key.value == Globals.PubKeys[3])
        expect(result.length).to.eq(1)
        done()
      })
    })
  })

  describe("The foundation account verifies account D. ", function() {
    it("Update the verified status to Y", function() {
      let payload = {
        from: web3.cmt.defaultAccount,
        candidateAddress: Globals.Accounts[3],
        verified: true
      }
      tx_result = web3.cmt.stake.validator.verify(payload)
      Utils.expectTxSuccess(tx_result)
      // check validator's status
      tx_result = web3.cmt.stake.validator.list()
      tx_result.data.forEach(d => (d.owner_address = d.owner_address.toLowerCase()))
      expect(tx_result.data).to.containSubset([
        { owner_address: Globals.Accounts[3], verified: "Y" }
      ])
    })
  })

  describe("Query validator D. ", function() {
    it("make sure all the information are accurate.", function() {
      tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
      // check validator's information
      logger.debug(tx_result.data)
      expect(tx_result.data.owner_address.toLowerCase()).to.eq(Globals.Accounts[3])
      expect(tx_result.data.verified).to.eq("Y")
      expect(tx_result.data.pub_key.value).to.eq(Globals.PubKeys[3])
      expect(tx_result.data.state).to.eq("Validator")
    })
  })

  describe("Deactivate and activate", function() {
    describe("Deactivate D", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[3]
        }
        tx_result = web3.cmt.stake.validator.deactivate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 1)
      })
      it("active=N, state=Candidate, vp=tvp=0", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.active).to.be.eq("N")
        expect(tx_result.data.state).to.be.eq("Candidate")
        expect(tx_result.data.voting_power).to.be.eq(0)
        // expect(tx_result.data.tendermint_voting_power).to.be.eq(0)
      })
    })
    describe("Deactivate C", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[2]
        }
        tx_result = web3.cmt.stake.validator.deactivate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 1)
      })
      it("active=N, state=Candidate, vp=tvp=0", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[2], 0)
        expect(tx_result.data.active).to.be.eq("N")
        expect(tx_result.data.state).to.be.eq("Candidate")
        expect(tx_result.data.voting_power).to.be.eq(0)
        // expect(tx_result.data.tendermint_voting_power).to.be.eq(0)
      })
    })
    describe("Activate D", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[3]
        }
        tx_result = web3.cmt.stake.validator.activate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 3)
      })
      it("active=Y, state=Validator, vp>1, tvp=1000", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
        expect(tx_result.data.active).to.be.eq("Y")
        expect(tx_result.data.state).to.be.eq("Validator")
        expect(tx_result.data.voting_power).to.be.eq(1000)
        // expect(tx_result.data.tendermint_voting_power).to.eq(10)
      })
    })
    describe("Activate C", function() {
      before(function(done) {
        let payload = {
          from: Globals.Accounts[2]
        }
        tx_result = web3.cmt.stake.validator.activate(payload)
        Utils.expectTxSuccess(tx_result)
        Utils.waitBlocks(done, 3)
      })
      it("C active=Y, state=Validator, vp>1, tvp=1000", function() {
        tx_result = web3.cmt.stake.validator.query(Globals.Accounts[2], 0)
        expect(tx_result.data.active).to.be.eq("Y")
        expect(tx_result.data.state).to.be.eq("Validator")
        expect(tx_result.data.voting_power).to.be.eq(1000)
        // expect(tx_result.data.tendermint_voting_power).to.eq(10)
      })
    })
  })

  describe("Voting Power", function() {
    it("check validator D's voting power", function() {
      let val_D = web3.cmt.stake.validator.query(Globals.Accounts[3], 0).data
      expect(val_D.voting_power).to.be.eq(1000)
    })
  })

  describe("Update Candidacy", function() {
    it("The verified status will set to false", function() {
      let website = "http://aaa.com"
      let pubKey = "LY3sRPcr63CE9uIJivApXlcYXKUoidtD+64mIljrYxk="
      let payload = {
        from: Globals.Accounts[3],
        pubKey: pubKey,
        description: {
          website: website
        }
      }
      tx_result = web3.cmt.stake.validator.update(payload)
      Utils.expectTxSuccess(tx_result)
      // check validator
      tx_result = web3.cmt.stake.validator.query(Globals.Accounts[3], 0)
      expect(tx_result.data.pub_key.value).to.be.eq(pubKey)
      expect(tx_result.data.description.website).to.be.eq(website)
      expect(tx_result.data.verified).to.be.eq("N")
    })
  })

  describe("Update D's account address", function() {
    let accountUpdateRequestId, accountUpdateRequestId2
    before(function() {
      newAccount = web3.personal.newAccount(Settings.Passphrase)
      web3.personal.unlockAccount(newAccount, Settings.Passphrase)
      newAccount2 = web3.personal.newAccount(Settings.Passphrase)
      web3.personal.unlockAccount(newAccount2, Settings.Passphrase)
      // balance before
      balance_old = Utils.getBalance(3)
    })

    it("fail if update to an existing delegator's address", function() {
      let payload = { from: Globals.Accounts[3], newCandidateAccount: Globals.Accounts[1] }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxFail(tx_result)
    })

    it("update validator's account address", function() {
      let payload = { from: Globals.Accounts[3], newCandidateAccount: newAccount }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxSuccess(tx_result)
      accountUpdateRequestId = Number(
        Buffer.from(tx_result.deliver_tx.data, "base64").toString("utf-8")
      )
    })

    it("fail if update to an address that has been used in update_account", function() {
      let payload = { from: Globals.Accounts[3], newCandidateAccount: newAccount }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxFail(tx_result)
    })

    it("request to update B's address to newAccount2", function() {
      let payload = { from: Globals.Accounts[1], newCandidateAccount: newAccount2 }
      tx_result = web3.cmt.stake.validator.updateAccount(payload)
      Utils.expectTxSuccess(tx_result)
      accountUpdateRequestId2 = Number(
        Buffer.from(tx_result.deliver_tx.data, "base64").toString("utf-8")
      )
    })

    describe("new account accept the update", function() {
      it("succeed", function() {
        let payload = { from: newAccount, accountUpdateRequestId: accountUpdateRequestId }
        tx_result = web3.cmt.stake.validator.acceptAccountUpdate(payload)
        Utils.expectTxSuccess(tx_result)
      })
    })
  })

  describe("Candidate drops candidacy", function() {
    let theAccount

    before(function() {
      theAccount = newAccount ? newAccount : Globals.Accounts[3]
    })

    it("Withdraw Candidacy", function(done) {
      let payload = { from: theAccount }
      tx_result = web3.cmt.stake.validator.withdraw(payload)
      Utils.expectTxSuccess(tx_result)
      Utils.waitBlocks(done, 3)
    })

    it("Account D no longer a validator, and genesis validator restored", function() {
      // check validators, no theAccount
      tx_result = web3.cmt.stake.validator.list()
      tx_result.data.forEach(d => (d.owner_address = d.owner_address.toLowerCase()))
      expect(tx_result.data).to.not.containSubset([{ owner_address: theAccount }])
      // check validators restored
      let vals = tx_result.data.filter(d => d.state == "Validator")
      expect(vals.length).to.eq(4)
    })

    it("4 validators on tendermint", function(done) {
      Utils.getTMValidators((err, res) => {
        expect(err).to.be.null
        expect(res).to.be.not.null
        expect(res.result.validators.length).to.eq(4)
        res.result.validators.forEach(v => {
          expect(Number(v.voting_power)).to.eq(1000)
        })
        done()
      })
    })
  })
})

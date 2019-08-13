require_relative './helper.rb'

describe "API Tests" do

  context "Check /ping responds" do
    it "should return 'pong'" do
      resp = get("/ping")
      expect(resp.code).to eq("200")
      expect(resp.body['ping']).to eq "pong"
    end
  end

  context "Check that an empty tx hash returns properly" do
    it "should have no values" do
      resp = get("/tx/bogus")
      expect(resp.body['request']).to eq(""), resp.body.inspect
      expect(resp.body['status']).to eq(""), resp.body.inspect
      expect(resp.body['txhash']).to eq(""), resp.body.inspect
    end
  end

  context "Create a pool" do

    it "should show up in listing of pools" do
      resp = get("/pools")
      # should have one pool added via genesis
      # If this line is failing, are we starting with a clean blockchain? Or
      # did we run before genesis could init pools?
      expect(resp.body.length).to eq(1), "Are you working from a clean blockchain? Did you wait until 1 block was create? \n(#{resp.code}: #{resp.body.inspect})"
    end

    it "create a pool for bnb" do
      tx = makeTx(memo: "create:TCAN-014")
      resp = processTx([tx])
      expect(resp.code).to eq("200"), resp.body.inspect
    end

    it "should be able to get the pool" do
      resp = get("/pool/TCAN-014")
      expect(resp.body['ticker']).to eq("TCAN-014"), resp.body.inspect
      expect(resp.body['status']).to eq("Bootstrap"), resp.body.inspect
    end

    it "check we cannot set pool status as non-admin" do
      skip "TODO - this check should pass, but doesn't"
      tx = makeTx(memo: "ADMIN:POOLSTATUS:TCAN-014:Enabled")
      resp = processTx(tx, user="alice")
      expect(resp.code).to eq("500")
    end

    it "set pool status to active, and that we can do multiple txs" do
      tx1 = makeTx(memo: "ADMIN:POOLSTATUS:TCAN-014:Enabled")
      tx2 = makeTx(memo: "ADMIN:POOLSTATUS:BNB:Enabled")
      resp = processTx([tx1, tx2])
      expect(resp.code).to eq("200")

      resp = get("/pool/TCAN-014")
      expect(resp.code).to eq("200")
      expect(resp.body['status']).to eq("Enabled"), resp.body.inspect

      resp = get("/pool/BNB")
      expect(resp.code).to eq("200")
      expect(resp.body['status']).to eq("Enabled"), resp.body.inspect
    end


    it "should not create a duplicate pool" do
      tx = makeTx(memo: "create:TCAN-014")
      resp = processTx(tx)
      expect(resp.code).to eq("200")
      
      resp = get("/pools")
      # should have one pool added via genesis
      expect(resp.body.length).to eq(2), resp.body.inspect
    end

    it "should show up in listing of pools" do
      resp = get("/pools")
      expect(resp.body[1]['ticker']).to eq("TCAN-014"), resp.body.inspect
    end

  end
  
  context "Stake/Unstake" do

    coins = [
      {'denom': "RUNE-B1A", "amount": "23.495"},
      {'denom': "TCAN-014", "amount": "3.3485"},
    ]
    sender = "bnb" + get_rand(39).downcase

    it "should be able to stake" do

      tx = makeTx(memo: "stake:TCAN-014", coins: coins, sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      resp = get("/pool/TCAN-014/stakers")
      expect(resp.code).to eq("200"), resp.body.inspect
      expect(resp.body['stakers'].length).to eq(1), resp.body['stakers'].inspect
      expect(resp.body['stakers'][0]['units']).to eq("13.42175000"), resp.body['stakers'][0].inspect
    end

    it "should be able to unstake" do
      tx = makeTx(memo: "withdraw:TCAN-014:100", sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      resp = get("/pool/TCAN-014/stakers")
      expect(resp.code).to eq("200"), resp.body.inspect
      expect(resp.body['stakers']).to eq(nil), resp.body.inspect
    end

    it "swap" do
      # stake some coins first
      tx = makeTx(memo: "stake:TCAN-014", coins: coins, sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      # make a swap
      txid = txid()
      coins = [{'denom': "RUNE-B1A", "amount": "0.2"}]
      tx = makeTx(
        memo: "swap:TCAN-014:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:0.160053", 
        coins: coins,
        hash: txid,
      )
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      resp = get("/pool/TCAN-014")
      expect(resp.code).to eq("200")
      expect(resp.body['balance_rune']).to eq("3.54850000"), resp.body.inspect
      expect(resp.body['balance_token']).to eq("22.24541406"), resp.body.inspect
    end
  end

end

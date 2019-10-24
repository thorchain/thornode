require_relative './helper.rb'

TRUST_BNB_ADDRESS="bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6"

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
      resp = get("/tx/A9A65505553D777E5CE957A74153F21EDD8AAA4B0868F2537E97E309945425B9")
      expect(resp.body['memo']).to eq(""), resp.body.inspect
      expect(resp.body['status']).to eq(""), resp.body.inspect
      expect(resp.body['txhash']).to eq(""), resp.body.inspect
    end
  end

  context "Check we have no completed events" do
    it "should be a nil" do
      resp = get("/events/1")
      expect(resp.body).to eq([]), resp.body.inspect
    end
  end


  context "Create a pool" do

    it "should show up in listing of pools" do
      resp = get("/pools")
      # Previously we add BNB pool in genesis , but now we removed it
      expect(resp.body).to eq([]), "Are you working from a clean blockchain? Did you wait until 1 block was create? \n(#{resp.code}: #{resp.body})"
    end

    it "create a pool for bnb" do
      tx = makeTx(memo: "create:BNB")
      resp = processTx([tx])
      expect(resp.code).to eq("200"), resp.body.inspect
    end

    it "create a pool for TCAN-014" do
      tx = makeTx(memo: "create:TCAN-014")
      resp = processTx([tx])
      expect(resp.code).to eq("200"), resp.body.inspect
    end

    it "pool should be enabled" do
      resp = get("/pool/TCAN-014")
      expect(resp.code).to eq("200")
      expect(resp.body['status']).to eq("Enabled"), resp.body.inspect

      resp = get("/pool/BNB.BNB")
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
      expect(resp.body[1]['asset']['symbol']).to eq("TCAN-014"), resp.body.inspect
    end

  end

  context "Stake/Unstake" do

    coins = [
      {'asset': {'chain': 'BNB', 'symbol': 'RUNE-B1A', 'ticker': 'RUNE'}, "amount": "2349500000"},
      {'asset': {'chain': 'BNB', 'symbol': 'TCAN-014', 'ticker': 'TCAN'}, "amount": "334850000"},
    ]
    sender = "bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38"

    it "should be able to stake" do

      tx = makeTx(memo: "stake:TCAN-014", coins: coins, sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      resp = get("/pool/TCAN-014/stakers")
      expect(resp.code).to eq("200"), resp.body.inspect
      expect(resp.body['stakers'].length).to eq(1), resp.body['stakers'].inspect
      expect(resp.body['stakers'][0]['units']).to eq("1342175000"), resp.body['stakers'][0].inspect
    end

    it "should be able to unstake" do
      tx = makeTx(memo: "withdraw:TCAN-014", sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      resp = get("/pool/BNB.TCAN-014/stakers")
      expect(resp.code).to eq("200"), resp.body.inspect
      expect(resp.body['stakers']).to eq(nil), resp.body.inspect
    end

    txid = txid() # outside it state so its value is available in multiple "it" statements
    it "swap" do
      # stake some coins first
      tx = makeTx(memo: "stake:BNB.TCAN-014", coins: coins, sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      # make a swap
      coins = [
        {'asset': {'chain': 'BNB', 'symbol': 'TCAN-014', 'ticker': 'TCAN'}, "amount": "20000000"},
      ]
      tx = makeTx(
        memo: "swap:RUNE-B1A:bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq:124958592",
        coins: coins,
        hash: txid,
      )
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      resp = get("/pool/TCAN-014")
      expect(resp.code).to eq("200")
      expect(resp.body['balance_rune']).to eq("2224541407"), resp.body.inspect
      expect(resp.body['balance_token']).to eq("354850000"), resp.body.inspect

      # another swap ,it should fail due to price protection
      tx1 = makeTx(
        memo: "swap:RUNE-B1A:bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq:134958590",
        coins: coins,
        hash: txid(),
      )
      resp = processTx(tx1)
      expect(resp.code).to eq("200"), resp.body.inspect

      # pool balance should not change
      resp = get("/pool/BNB.TCAN-014")
      expect(resp.code).to eq("200")
      expect(resp.body['balance_rune']).to eq("2224541407"), resp.body.inspect
      expect(resp.body['balance_token']).to eq("354850000"), resp.body.inspect
    end

    it "Send outbound tx and mark tx'es as complete" do
      # find the block height of the previous swap transaction
      i = 1
      found = false
      until i > 100
        resp = get("/txoutarray/#{i}")
        puts resp.body
        if not resp.body['chains'].include?("BNB")
          i = i + 1
          next
        end
        arr = resp.body['chains']['BNB']
        unless arr['tx_array'].empty?
          if arr['tx_array'][0]['to'] == "bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq"
            # we have found the block height of our last swap
            found = true
            newTxId = txid()
            tx = makeTx(memo: "outbound:#{i}", hash:newTxId, sender:TRUST_BNB_ADDRESS)
            resp = processTx(tx)
            expect(resp.code).to eq("200"), resp.body.inspect

            # wait for 1s give statechain more time to process the tx
            sleep(1)
            resp = get("/tx/#{txid}")
            expect(resp.code).to eq("200")
            expect(resp.body['txhash']).to eq(newTxId), resp.body.inspect
            break
          end
        end
        i = i + 1
      end

      expect(found).to eq(true)

    end

    it "check events are completed" do
      resp = get("/events/1")
      expect(resp.body.count).to eq(3), resp.body.inspect
      expect(resp.body[2]['pool']['symbol']).to eq("TCAN-014"), resp.body[2].inspect
      expect(resp.body[2]['type']).to eq("swap"), resp.body[2].inspect
      expect(resp.body[2]['in_hash']).to eq(txid), resp.body[2].inspect
    end

    it "add tokens to a pool" do
      coins = [
        {'asset': {'chain': 'BNB', 'symbol': 'RUNE-B1A', 'ticker': 'RUNE'}, "amount": "20000000"},
        {'asset': {'chain': 'BNB', 'symbol': 'TCAN-014', 'ticker': 'TCAN'}, "amount": "20000000"},
      ]
      tx = makeTx(memo: "add:TCAN-014", coins: coins, sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect

      resp = get("/pool/TCAN-014")
      expect(resp.code).to eq("200")
      expect(resp.body['balance_rune']).to eq("2244541407"), resp.body.inspect
      expect(resp.body['balance_token']).to eq("374850000"), resp.body.inspect
    end

    it "adds gas" do
      coins = [
        {'asset': {'chain': 'BNB', 'symbol': 'RUNE-B1A', 'ticker': 'RUNE'}, "amount": "20000000"},
      ]
      tx = makeTx(memo: "GAS", coins: coins, sender: sender)
      resp = processTx(tx)
      expect(resp.code).to eq("200"), resp.body.inspect
    end

  end

  context "Block heights" do
    it "ensure we have non-zero block height" do
      resp = get("/lastblock")
      expect(resp.code).to eq("200")
      # expect(resp.body['lastobservedin'].to_i).to be > 0, resp.body.inspect
      expect(resp.body['lastsignedout'].to_i).to be > 0, resp.body.inspect
      expect(resp.body['statechain'].to_i).to be > 0, resp.body.inspect
    end
  end

end

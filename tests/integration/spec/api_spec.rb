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
      expect(resp.body['request']).to eq(""), resp.body
      expect(resp.body['status']).to eq(""), resp.body
      expect(resp.body['txhash']).to eq(""), resp.body
    end
  end

  context "Create a pool" do

    it "should show up in listing of pools" do
      resp = get("/pools")
      # should have one pool added via genesis
      expect(resp.body.length).to eq(1), resp.body
    end

    it "create a pool for bnb" do
      resp = processTx("create:TCAN-014")
      expect(resp.code).to eq("200"), "Are you working from a clean blockchain? \n(#{resp.code}: #{resp.body})"
    end

    it "should be able to get the pool" do
      resp = get("/pool/TCAN-014")
      expect(resp.body['ticker']).to eq("TCAN-014"), resp.body
    end


    it "should not create a duplicate pool" do
      resp = processTx("create:TCAN-014")
      expect(resp.code).to eq("500")
      
      resp = get("/pools")
      # should have one pool added via genesis
      expect(resp.body.length).to eq(2), resp.body
    end

    it "should show up in listing of pools" do
      resp = get("/pools")
      expect(resp.body[1]['ticker']).to eq("TCAN-014"), resp.body
    end

  end

end

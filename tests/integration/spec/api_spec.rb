require_relative './helper.rb'

describe "API Tests" do

  context "Check /ping responds" do
    it "should return 'pong'" do
      resp = get("/ping")
      expect(resp.body['ping']).to eq "pong"
    end
  end

  context "Check that an empty tx hash returns properly" do
    it "should have no values" do
      resp = get("/tx/bogus")
      expect(resp.body['request']).to eq ""
      expect(resp.body['status']).to eq ""
      expect(resp.body['txhash']).to eq ""
    end
  end

  context "Create a pool" do
    it "create a pool for bnb" do
      resp = processTx("AF64E866F7EDD74A558BF1519FB12700DDE51CD0DB5166ED37C568BE04E0C7F3")
      puts resp.body
      expect(resp.code).to eq(200), "Are you working from a clean blockchain?"
    end
  end

end

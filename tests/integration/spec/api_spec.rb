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
      resp = processTx("FE0EC391E62AF2E291A41C38DEB3DF180FB0FD0E21E8B3866EFF8912F65FD1EE")
      puts resp.body
      expect(resp.body['txhash']).to eq ""
    end
  end

end

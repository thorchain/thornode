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
      puts resp.body
      expect(resp.body['request']).to eq ""
      expect(resp.body['status']).to eq ""
      expect(resp.body['txhash']).to eq ""
    end
  end

end

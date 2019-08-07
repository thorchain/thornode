require 'net/http'
require 'json'

HOST = "localhost"
PORT = 1317

def get(path)
  resp = Net::HTTP.get_response(HOST, "/swapservice#{path}", PORT)
  resp.body = JSON.parse(resp.body)
  return resp
end

describe "API" do
  context "When testing the /ping API endpoint" do
    it "should return 'pong'" do
      resp = get("/ping")
      expect(resp.body['ping']).to eq "pong"
    end
  end

  context "Empty tx" do
    it "should have no values" do
      resp = get("/tx/bogus")
      puts resp.body
      expect(resp.body['done']).to eq false
    end
  end
end

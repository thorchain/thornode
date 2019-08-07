require 'net/http'
require 'json'

HOST = "localhost"
PORT = 1317

def get(path)
  resp = Net::HTTP.get_response(HOST, "/swapservice#{path}", PORT)
  resp.body = JSON.parse(resp.body)
  return resp
end



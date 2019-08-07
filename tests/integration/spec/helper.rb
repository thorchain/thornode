require 'net/http'
require 'pp'
require 'json'
# require 'tempfile'

HOST = "localhost"
PORT = 1317
HTTP = Net::HTTP.new(HOST, PORT)

def get(path)
  resp = Net::HTTP.get_response(HOST, "/swapservice#{path}", PORT)
  resp.body = JSON.parse(resp.body)
  return resp
end

def processTx(hash, mode = 'block')
  request = Net::HTTP::Post.new("/swapservice/binance/tx")
  request.body = {
    'tx_hash': hash,
    'base_req': {
      'chain_id': "sschain",
      'from': 'rune1ewqpdu8lf30skrdv7u8twh50twq5k2puvusfsn' # TODO: make address configurable
    }
  }.to_json
  resp = HTTP.request(request)

  # write unsigned json to disk
  File.open("/tmp/unSigned.json", "w") { |file| file.puts resp.body}
  signed = `echo "password" | sscli tx sign /tmp/unSigned.json --from jack`
  signed = JSON.parse(signed)
  signedJson = {
    'mode': mode,
    'tx': signed['value'],
  }
  # pp signedJson

  request = Net::HTTP::Post.new("/txs")
  request.body = signedJson.to_json
  resp = HTTP.request(request)
  resp.body = JSON.parse(resp.body)

  return resp
end

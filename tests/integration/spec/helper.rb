require 'net/http'
require 'pp'
require 'json'
# require 'tempfile'

HOST = ENV['APIHOST'] || "localhost"
PORT = ENV['APIPORT'] || 1317
HTTP = Net::HTTP.new(HOST, PORT)

def get(path)
  resp = Net::HTTP.get_response(HOST, "/swapservice#{path}", PORT)
  resp.body = JSON.parse(resp.body)
  return resp
end

def processTx(hash, mode = 'block')
  request = Net::HTTP::Post.new("/swapservice/binance/tx")
  address = `sscli keys show jack -a`.strip!
  request.body = {
    'tx_hash': hash,
    'base_req': {
      'chain_id': "sschain",
      'from': address
    }
  }.to_json
  resp = HTTP.request(request)

  # write unsigned json to disk
  File.open("/tmp/unSigned.json", "w") { |file| file.puts resp.body}
  signedTx = `echo "password" | sscli tx sign /tmp/unSigned.json --from jack --node "tcp://daemon:26657"`
  signedTx = JSON.parse(signedTx)
  signedJson = {
    'mode': mode,
    'tx': signedTx['value'],
  }
  # pp signedJson

  request = Net::HTTP::Post.new("/txs")
  request.body = signedJson.to_json
  resp = HTTP.request(request)
  resp.body = JSON.parse(resp.body)

  return resp
end

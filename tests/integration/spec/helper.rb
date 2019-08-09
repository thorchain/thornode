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

def processTx(memo, mode = 'block')
  request = Net::HTTP::Post.new("/swapservice/binance/tx")
  address = `sscli keys show jack -a`.strip!
  hash = 'AF64E866F7EDD74A558BF1519FB12700DDE51CD0DB5166ED37C568BE04E0C'
  request.body = {
    'blockHeight': '376',
    'count': '1',
    'base_req': {
      'chain_id': "sschain",
      'from': address
    },
    'txArray': [
      {
        'tx': hash,
        'sender': address,
        'MEMO': memo,
        'coins': [{
          'denom': 'BNB',
          'amount': '1',
        }],
      }
    ],
  }.to_json
  puts(request.body)
  resp = HTTP.request(request)

  # write unsigned json to disk
  File.open("/tmp/unSigned.json", "w") { |file| file.puts resp.body}
  signedTx = `echo "password" | sscli tx sign /tmp/unSigned.json --from jack`
  puts("hello")
  puts(signedTx)
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

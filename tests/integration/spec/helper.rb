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
  hash = '7E5DF2DAF3463FEFA633EC1B45ADC434AAE92A55823E210837E975F1FE289BA7'
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
        'sender': "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6",
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

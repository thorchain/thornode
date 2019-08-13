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

def get_rand(len)
  (1..len).map{ rand(36).to_s(36) }.join
end

# generates a random ticker
def ticker()
  return "#{get_rand(3).upcase}-#{get_rand(3).upcase}"
end

def processTx(memo, hash=nil, sender=nil, mode='block', coins=nil, user="jack")
  request = Net::HTTP::Post.new("/swapservice/binance/tx")
  address = `sscli keys show #{user} -a`.strip!
  hash ||= get_rand(64).upcase
  sender ||= "bnb" + get_rand(39).downcase
  coins ||= [{
    'denom': 'BNB',
    'amount': '1',
  }]

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
        'sender': sender,
        'MEMO': memo,
        'coins': coins,
      }
    ],
  }.to_json

  resp = HTTP.request(request)
  if resp.code != "200" 
    pp resp.body
    return resp
  end

  # write unsigned json to disk
  File.open("/tmp/unSigned.json", "w") { |file| file.puts resp.body}
  signedTx = `echo "password" | sscli tx sign /tmp/unSigned.json --from #{user}`
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

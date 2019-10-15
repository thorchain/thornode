require 'net/http'
require 'pp'
require 'json'
require 'securerandom'
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
  str = SecureRandom.hex(len)
  return str.slice(0, len)
end

# generates a random ticker
def ticker()
  return "#{get_rand(3).upcase}-#{get_rand(3).upcase}"
end

def txid()
  get_rand(64).upcase
end

def bnbAddress()
  "bnb" + get_rand(39).downcase
end

def makeTx(memo:'', hash:nil, sender:nil, coins:nil, poolAddr:nil)
  hash ||= txid()
  sender ||= bnbAddress
  coins ||= [{
    'denom': 'RUNE-B1A',
    'amount': '1',
  }]
  poolAddr ||= TRUST_BNB_ADDRESS
  return {
    'tx': hash,
    'sender': sender,
    'observe_pool_address': poolAddr,
    'MEMO': memo,
    'coins': coins
  }
end

def processTx(txs, user="jack", mode='block')
  request = Net::HTTP::Post.new("/swapservice/binance/tx")
  address = `sscli keys show #{user} -a`.strip!
  txs = [txs].flatten(1) # ensures we are an array, and not just a single hash
  request.body = {
    'blockHeight': '376',
    'count': '1',
    'base_req': {
      'chain_id': "statechain",
      'from': address
    },
    'txArray': txs,
  }.to_json
  #puts(request.body.to_json)

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
  #pp signedJson


  request = Net::HTTP::Post.new("/txs")
  request.body = signedJson.to_json
  resp = HTTP.request(request)
  resp.body = JSON.parse(resp.body)

  return resp
end

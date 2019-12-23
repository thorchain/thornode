require 'net/http'
require 'pp'
require 'json'
require 'securerandom'
# require 'tempfile'

HOST = ENV['APIHOST'] || "localhost"
PORT = ENV['APIPORT'] || 1317
HTTP = Net::HTTP.new(HOST, PORT)
$lastget = Time.now()

def get(path)
  # since THORNode rate limit our API, check its been more than than a second since
  # the last query
  if Time.now() - $lastget < 1
    sleep(1)
  end
  # puts(path)
  resp = Net::HTTP.get_response(HOST, "/thorchain#{path}", PORT)
  resp.body = JSON.parse(resp.body)
  $lastget = Time.now()
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
  [ 
    "bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u",
    "bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38",
    "bnb1yk882gllgv3rt2rqrsudf6kn2agr94etnxu9a7",
    "bnb1t3c49u74fum2gtgekwqqdngg5alt4txrq3txad",
    "bnb1hpa7tfffxadq9nslyu2hu9vc44l2x6ech3767y",
    "bnb1llvmhawaxxjchwmfmj8fjzftvwz4jpdhapp5hr",
    "bnb1s3f8vxaqum3pft6cefyn99px8wq6uk3jdtyarn",
    "bnb1e6y59wuz9qqcnqjhjw0cl6hrp2p8dvsyxyx9jm",
    "bnb1zxseqkfm3en5cw6dh9xgmr85hw6jtwamnd2y2v",
  ].sample

end

def makeTx(memo:'', hash:nil, sender:nil, coins:nil)
  # fetch vault address and pubkey
  vault = get("/pool_addresses").body['current'][0]

  hash ||= txid()
  sender ||= bnbAddress
  gas ||= [{
    'asset': {
      'chain': 'BNB',
      'symbol': 'BNB',
      'ticker': 'BNB',
    },
    'amount': '13750',
  }]
  coins ||= [{
    'asset': {
      'chain': 'BNB',
      'symbol': 'RUNE-B1A',
      'ticker': 'RUNE',
    },
    'amount': '1',
  }]
  return {
    'tx': {
      'id': hash,
      'from_address': sender,
      'chain': 'BNB',
      'to_address': vault['address'],
      'coins': coins,
      'memo': memo,
      'gas': gas,
    },
    'block_height': '376',
    'observed_pub_key': vault['pub_key'],
  }
end

def processTx(txs, user="thorchain", mode='block')
  request = Net::HTTP::Post.new("/thorchain/txs")
  address = `thorcli keys show #{user} -a`.strip!
  txs = [txs].flatten(1) # ensures THORNode are an array, and not just a single hash
  request.body = {
    'blockHeight': '376',
    'chain': 'bnb',
    'count': '1',
    'txs': txs,
    'base_req': {
      'chain_id': "thorchain",
      'from': address,
      'gas': 'auto',
    },
  }.to_json
    #puts(request.body.to_json)

    resp = HTTP.request(request)
    if resp.code != "200" 
      pp resp.body
      return resp
    end

    # write unsigned json to disk
    File.open("/tmp/unSigned.json", "w") { |file| file.puts resp.body}
    signedTx = `echo "password" | thorcli tx sign /tmp/unSigned.json --from #{user}`
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

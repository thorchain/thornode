# check for all bonded nodes that bond https://testnet-seed.thorchain.info/bonded_nodes.json

# loop through the endpoint above and select the PUB_KEY one by one

# use the pubkey to search the http://$PEER:1317/thorchain/nodeaccounts

# run this job daily

import requests
import json
import boto3

thornode_env = 'testnet'

s3_file = 'nodes.json'

seed_bucket = "{}-seed.thorchain.info".format(thornode_env)

seed_endpoint = "https://{}".format(seed_bucket)

bonded_nodes = "{}/bonded_nodes.json".format(seed_endpoint)

seed_request = requests.get(bonded_nodes)

all_bonded_nodes = seed_request.text.splitlines()

number_of_bonded_nodes = len(all_bonded_nodes)

active_nodes = []

standby_nodes = []

ready_nodes = []

white_listed_nodes = []

peer = json.loads(seed_request.text.splitlines()[0])['ip']

node_accounts_url = "http://{}:1317/thorchain/nodeaccounts".format(peer)

number_of_node_accounts = len(json.loads(requests.get(node_accounts_url).text))

for num in range(number_of_bonded_nodes):
    ip = json.loads(seed_request.text.splitlines()[num])['ip']
    date = json.loads(seed_request.text.splitlines()[num])['date']
    pub_key = json.loads(seed_request.text.splitlines()[num])['PUB_KEY']

    for num in range(number_of_node_accounts):
        if json.loads(requests.get(node_accounts_url).text)[num]['pub_key_set']['ed25519'] == pub_key:
            status = json.loads(requests.get(node_accounts_url).text)[num]['status']
            if status == 'active':
                active_nodes.append(ip)
            elif status == 'standby':
                standby_nodes.append(ip)
            elif status == 'whitelisted':
                white_listed_nodes.append(ip)
            elif status == 'ready':
                ready_nodes.append(ip)

print("active nodes = {} ".format(active_nodes))
print("standby nodes = {} ".format(standby_nodes))
print("whitelisted nodes = {} ".format(white_listed_nodes))
print("ready nodes = {} ".format(ready_nodes))

# construct payload

data = {
    "active" : active_nodes,
    "standby" : standby_nodes,
    "ready":    ready_nodes,
    "whitelisted": white_listed_nodes
}

print(data)

with open(s3_file, 'w') as outfile:
    json.dump(data, outfile)

# post payload
data = open(s3_file, 'rb')

s3 = boto3.resource('s3')
s3.meta.client.upload_file(s3_file, seed_bucket, s3_file, ExtraArgs={'ContentType': "application/json", 'ACL': "public-read"} )


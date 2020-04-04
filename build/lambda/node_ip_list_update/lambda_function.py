import json
import requests
import boto3

def get_new_ip_list(ip_addr):
  response = requests.get('http://' + ip_addr + ':26657/net_info')
  peers = response.json()['result']['peers']

  new_ip_list = []
  for peer in peers:
    remote_ip = peer['remote_ip']
    new_ip_list.append(remote_ip)

  new_ip_list = list(set(new_ip_list)) # avoid duplicates

  return new_ip_list

def lambda_handler(event, context):
  s3_resource = boto3.resource('s3')
  bucket = 'testnet-seed.thorchain.info'
  prefix = 'node_ip_'
  thorchain_bucket = s3_resource.Bucket(bucket)

  try:
    for obj in thorchain_bucket.objects.all():
      key = obj.key
      if (key[0:8] == prefix):
        print('key: ' + key)
        body = obj.get()['Body'].read()
        ip_list = json.loads(body)

        updated_ip_list = []
        for ip_addr in ip_list:
          updated_ip_list += get_new_ip_list(ip_addr)
        

        updated_ip_list = list(set(updated_ip_list)) # avoid duplicates
        new_body = json.dumps(updated_ip_list)
        obj.put(Body=new_body)

    return {
      'message': 'successfully updated!'
    }
  except:
    return {
      'message': 'exception occured!'
    }


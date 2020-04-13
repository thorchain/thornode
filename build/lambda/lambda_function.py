import json
import requests
import boto3

prefix = 'node_ip_'
s3_resource = boto3.resource('s3')
buckets = ['testnet-seed.thorchain.info']

def get_url(ip_addr, path):
  return 'http://' + ip_addr + ':26657' + path

def get_new_ip_list(ip_addr):
  response = requests.get(get_url(ip_addr, "/net_info"))
  peers = [x['remote_ip'] for x in response.json()['result']['peers']]
  peers.append(ip_addr)
  peers = list(set(peers)) # uniqify

  # filter nodes that are not "caught up"
  results = []    
  for peer in peers:
    response = requests.get(get_url(ip_addr, "/status"))
    if not response.json()['result']['sync_info']['catching_up']:
        results.append(peer)

  return results


def lambda_handler(event, context):
  try:
    for bucket in buckets:
      thorchain_bucket = s3_resource.Bucket(bucket)

      for obj in thorchain_bucket.objects.all():
        key = obj.key
        if key.startswith(prefix):
          print('key: ' + key)
          body = obj.get()['Body'].read()
          ip_list = json.loads(body)

          updated_ip_list = []
          for ip_addr in ip_list:
            updated_ip_list += get_new_ip_list(ip_addr)
          

          updated_ip_list = list(set(updated_ip_list)) # avoid duplicates


          if len(updated_ip_list) != 0:
            new_body = json.dumps(updated_ip_list)
            obj.put(Body=new_body)


    return {
      'message': 'successfully updated!'
    }
  except:
    return {
      'message': 'exception occured!'
    }


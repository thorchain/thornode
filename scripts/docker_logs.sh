set -ex

docker ps -a

mkdir /tmp/logs/
for id in $(docker ps -q); do
  docker logs $id > /tmp/logs/$id.log
done

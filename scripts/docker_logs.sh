set -ex

docker ps -a

mkdir logs
for id in $(docker ps -q); do
  docker logs $id > ./logs/$id.log
done

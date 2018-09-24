#docker stack rm myApp
docker rm $(docker ps -a -q) --force
docker image prune
docker build . -t kademlia
sleep 5
#docker stack deploy myApp -c docker-compose.yml
docker-compose up --scale kademliaNodes=3


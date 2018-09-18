docker stack rm myApp
docker build . -t kademlia
sleep 5
docker stack deploy myApp -c docker-compose.yml

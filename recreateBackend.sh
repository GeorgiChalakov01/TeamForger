cd backend
./rebuildImage.sh
cd ..

docker stop teamforger-backend-1
docker rm teamforger-backend-1
docker run -d --name teamforger-backend-1 -e VIRTUAL_HOST=teamforger.gchalakov.com -e LETSENCRYPT_HOST=teamforger.gchalakov.com --network net -p 8080:8080 teamforger-backend
docker logs --follow teamforger-backend-1

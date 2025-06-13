source ./config.sh

cd backend
./rebuildImage.sh
cd ..

docker stop teamforger-backend-1
docker rm teamforger-backend-1

docker run -d \
	--name teamforger-backend-1 \
	-e DB_USER=$DB_USER \
	-e DB_PORT=$DB_PORT \
	-e DB_PWD=$DB_PWD \
	-e DB_SCHEMA=$DB_SCHEMA \
	-e DB_CONTAINER_NAME=$DB_CONTAINER_NAME \
	-e VIRTUAL_HOST=$BE_HOST \
	-e LETSENCRYPT_HOST=$BE_HOST \
	--network net \
	-p "$BE_PORT:$BE_PORT" \
	teamforger-backend

docker logs --follow teamforger-backend-1

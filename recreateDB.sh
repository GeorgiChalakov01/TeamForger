source ./config.sh

docker stop teamforger-db-1
docker rm teamforger-db-1

cd db
./rebuild.sh
cd ..

docker run -d \
	--name $DB_CONTAINER_NAME \
	-p "$DB_PORT:$DB_PORT" \
	-e POSTGRES_PASSWORD=$DB_PWD \
	-e POSTGRES_USER=$DB_USER \
	-e POSTGRES_DB=$DB_SCHEMA \
	--network net \
	-v /home/gchalakov/services/teamforger/db/pgdata:/var/lib/postgresql/ \
	teamforger-postgres

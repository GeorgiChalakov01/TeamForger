docker stop teamforger-db-1
docker rm teamforger-db-1

cd db
./rebuild.sh
cd ..

docker run -d --name teamforger-db-1 -p 5432:5432 -e POSTGRES_PASSWORD=ChangeMe --network net -v /home/gchalakov/services/teamforger/db/pgdata:/var/lib/postgresql/ teamforger-postgres

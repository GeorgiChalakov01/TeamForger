recreateBackend.sh

docker stop teamforger-db-1
docker rm teamforger-db-1
docker run -d --name teamforger-db-1 -p 5432:5432 -e POSTGRES_PASSWORD=password -v /home/gchalakov/services/teamforger/db/pgdata:/var/lib/postgresql/ teamforger-postgres

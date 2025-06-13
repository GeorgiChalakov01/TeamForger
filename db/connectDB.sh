source ../config.sh

docker exec -it teamforger-db-1 psql -h localhost -p 5432 -d $DB_SCHEMA -U $DB_USER

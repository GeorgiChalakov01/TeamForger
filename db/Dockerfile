FROM postgres:17.5-alpine3.22

ENV POSTGRES_USER=postgres
ENV POSTGRES_DB=developerDB
COPY sql/1-createTables.sql /docker-entrypoint-initdb.d/

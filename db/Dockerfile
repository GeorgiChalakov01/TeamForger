FROM postgres:17.5-alpine3.22

# Defaults if not specified in the docker run
ENV POSTGRES_USER=teamforger
ENV POSTGRES_PASSWORD=teamforger
ENV POSTGRES_DB=teamforger

COPY sql/1-createTables.sql /docker-entrypoint-initdb.d/

FROM postgres:17.5-alpine3.22

# Defaults if not specified in the docker run
ENV POSTGRES_USER=teamforger
ENV POSTGRES_PASSWORD=teamforger
ENV POSTGRES_DB=teamforger

# Install the pgvector extension and create symlinks.
RUN apk add --no-cache postgresql-pgvector && \
	VECTOR_PATH=$(find /usr/lib/postgresql* -name vector.so | head -1) && \
	PG_LIB_DIR=$(pg_config --pkglibdir) && \
	ln -sf "$VECTOR_PATH" "$PG_LIB_DIR/vector.so" && \
	ln -s /usr/share/postgresql/extension/vector* /usr/local/share/postgresql/extension/

COPY sql/1-createTables.sql /docker-entrypoint-initdb.d/

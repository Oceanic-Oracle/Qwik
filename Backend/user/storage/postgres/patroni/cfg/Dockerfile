FROM postgres:15

ENV DEBIAN_FRONTEND=noninteractive \
    TZ=UTC

RUN apt-get update && \
    apt-get install -y --no-install-recommends python3 python3-pip python3-venv build-essential libpq-dev && \
    rm -rf /var/lib/apt/lists/*

RUN python3 -m venv /patroni-venv && \
    chown -R postgres:postgres /patroni-venv

ENV PATH="/patroni-venv/bin:$PATH"

RUN pip install --no-cache-dir patroni[zookeeper] psycopg2-binary

RUN mkdir -p /var/lib/postgresql/data/pgdata && \
    chown -R postgres:postgres /var/lib/postgresql/data && \
    chmod 700 /var/lib/postgresql/data/pgdata

USER postgres

EXPOSE 8008

CMD ["patroni", "/etc/patroni.yml"]
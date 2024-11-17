FROM postgres:17.1

COPY ./up.sql  /docker-entrypoint-initdb.d/1.sql

CMD ["postgres"]
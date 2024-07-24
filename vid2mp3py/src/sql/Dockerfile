FROM mysql:latest

COPY ./init_sql.sh /docker-entrypoint-initdb.d/

EXPOSE 3306

ENTRYPOINT [ "init_sql.sh" ]
FROM mariadb:10.1

ADD 00-init.sql /docker-entrypoint-initdb.d/00-init.sql
ADD 01-init.sql /docker-entrypoint-initdb.d/01-init.sql
ENV MYSQL_ALLOW_EMPTY_PASSWORD 1
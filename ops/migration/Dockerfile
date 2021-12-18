FROM alpine:3.13

WORKDIR /app

RUN apk --no-cache add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
RUN mv /app/migrate.linux-amd64 /usr/local/bin/migrate

COPY migrations ./migrations

COPY ops/migration/wait-for.sh .
RUN chmod +x wait-for.sh

COPY ops/migration/start.sh .
RUN chmod +x start.sh

CMD ["/app/wait-for.sh", "postgres:5432", "--", "/app/start.sh" ]

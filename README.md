# life_easy_api_backed

## PostgreSQL and DataGrip

`.env.example` is intentionally blank and must not be used directly to start
the project. Create a private `.env` with random credentials instead:

```sh
./scripts/init-env.sh
```

The script refuses to overwrite an existing `.env`. The private file exposes
two connection URLs:

- `DATABASE_URL`: used by containers; its host is `postgres`.
- `DATABASE_URL_LOCAL`: used by DataGrip on the host machine; its host is
  `localhost`.
- `PSQL_USER_DB_URL`: alias of `DATABASE_URL_LOCAL` for local PostgreSQL
  clients, including DataGrip.

PostgreSQL listens at `127.0.0.1:${POSTGRES_PORT}` only, so it is not exposed
to other devices on the network. Start the stack with `docker compose up -d`.
Then paste the resolved value of `DATABASE_URL_LOCAL` into DataGrip, or use
host `localhost`, port `5432`, and the corresponding PostgreSQL credentials.

If a username or password contains URL-reserved characters such as `@`, `:`,
`/`, `?`, or `#`, percent-encode that value in both connection URLs.

# life_easy_api_backed

## PostgreSQL and DataGrip

Copy `.env.example` to `.env` and set the passwords. The file exposes two
connection URLs:

- `DATABASE_URL`: used by containers; its host is `postgres`.
- `DATABASE_URL_LOCAL`: used by DataGrip on the host machine; its host is
  `localhost`.

PostgreSQL listens at `127.0.0.1:${POSTGRES_PORT}` only, so it is not exposed
to other devices on the network. Start the stack with `docker compose up -d`.
Then paste the resolved value of `DATABASE_URL_LOCAL` into DataGrip, or use
host `localhost`, port `5432`, and the corresponding PostgreSQL credentials.

If a username or password contains URL-reserved characters such as `@`, `:`,
`/`, `?`, or `#`, percent-encode that value in both connection URLs.

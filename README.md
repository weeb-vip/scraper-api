# scraper-api

Anime metadata from [TheTVDB](https://thetvdb.com), and the links between their
records and weeb.vip's own.

A Go GraphQL service built with [gqlgen](https://gqlgen.com) with federation
support. PostgreSQL through GORM, [Chi](https://go-chi.io) for the HTTP layer,
and Cobra for the CLI around it.

Its endpoints are authenticated: this is an internal service for enriching the
catalogue, not something the site calls on a reader's behalf.

## What it links

The migrations tell the story — a TheTVDB link, then seasons, then creation and
name records against it. That mapping is what lets a series in the catalogue
pick up artwork, episode lists and air dates from TheTVDB without the two ever
sharing an id.

`internal/services/thetvdb_api` is the client; the resolvers around it expose
the results.

## Running it

Requires Go and PostgreSQL, and TheTVDB API credentials.

```sh
make migrate                  # bring the database up to date
go run cmd/main.go server     # the GraphQL server
```

`config/config.dev.json` is the local config; `config/config.go` has the shape,
and environment variables override it.

## Schema and generated code

```sh
make gql        # regenerate resolvers from graph/schema.graphqls
make mocks      # regenerate test mocks
make generate   # both
```

## Migrations

```sh
make create-migration name=add_something
make migrate
```

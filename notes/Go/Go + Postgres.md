### pgx

`pgx` is a Go package built for PostgresQL which is more efficient and better suited for communicating with specifically PostgresQL databases.

Add pgx and a CLI manager to build the task manager, since we need a way to execute CLI commands to actually do things with and to the DB:
```
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/urfave/cli/v2
```


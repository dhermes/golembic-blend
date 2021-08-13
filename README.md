# `golembic-blend`

> SQL Schema Management in Go, inspired by `sqlalchemy/alembic`

[![GoDoc][1]][2]

The goal of this repository is to reframe `github.com/dhermes/golembic` in
terms of the primitives in `github.com/blend/go-sdk/db`.

## Examples

If the database is empty, all migrations will be run:

```
$ make restart-postgres
...
$ go run ./examples/cmd/
2021-08-13T04:41:26.039033Z    [db.migration] -- applied -- Check table does not exist: golembic_migrations
2021-08-13T04:41:26.042609Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T04:41:26.04766Z     [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T04:41:26.058502Z    [db.migration] -- applied -- 3f34bd961f15: Create users table
2021-08-13T04:41:26.066453Z    [db.migration] -- applied -- 464bc456c630: Seed data in users table
2021-08-13T04:41:26.072415Z    [db.migration] -- applied -- 959456a8af88: Add city column to users table
2021-08-13T04:41:26.078768Z    [db.migration] -- applied -- 57393d6ddb95: Rename the root user [MILESTONE]
2021-08-13T04:41:26.100146Z    [db.migration] -- applied -- 4d07dd6af28d: Add index on user emails (concurrently)
2021-08-13T04:41:26.113593Z    [db.migration] -- applied -- 2a35ccd628bc: Create books table
2021-08-13T04:41:26.121265Z    [db.migration] -- applied -- 3196713ca7e6: Create movies table
2021-08-13T04:41:26.123174Z    [db.migration.stats] 9 applied 0 skipped 0 failed 9 total
```

After creation, the next run does nothing

```
$ go run ./examples/cmd/
2021-08-13T04:41:56.152839Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T04:41:56.155591Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T04:41:56.160677Z    [db.migration] -- plan -- No migrations to run; latest revision: 3196713ca7e6
2021-08-13T04:41:56.1607Z      [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T04:41:56.16245Z     [db.migration.stats] 1 applied 1 skipped 0 failed 2 total
```

If we manually delete one, the last migration will get run

```
$ make psql
...
golembic=> DELETE FROM golembic_migrations WHERE revision = '3196713ca7e6';
DELETE 1
golembic=> DROP TABLE movies;
DROP TABLE
golembic=> \q
$
$
$ go run ./examples/cmd/
2021-08-13T04:43:19.237834Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T04:43:19.240721Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T04:43:19.246607Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T04:43:19.255715Z    [db.migration] -- applied -- 3196713ca7e6: Create movies table
2021-08-13T04:43:19.257395Z    [db.migration.stats] 2 applied 1 skipped 0 failed 3 total
```

[1]: https://godoc.org/github.com/dhermes/golembic-blend?status.svg
[2]: https://godoc.org/github.com/dhermes/golembic-blend

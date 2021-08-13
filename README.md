# `golembic-blend`

> SQL Schema Management in Go, inspired by `sqlalchemy/alembic`

[![GoDoc][1]][2]

The goal of this repository is to reframe `github.com/dhermes/golembic` in
terms of the primitives in `github.com/blend/go-sdk/db`.

## Examples

### Typical Usage

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

### Error Mode: Stale Checkout

During development or in `sandbox` environments, sometimes a team member
will try to run migrations with an out-of-date branch. Conversely, a team
member could run a migration that has not been checked into the mainline
branch. In either of these cases, the migrations table will contain an
**unknown** migration.

We can artificially introduce a "new" migration to see what such a failure
would look like:

```
$ go run ./examples/cmd/
...
2021-08-13T17:31:45.242616Z    [db.migration] -- applied -- 3196713ca7e6: Create movies table
2021-08-13T17:31:45.245323Z    [db.migration.stats] 9 applied 0 skipped 0 failed 9 total
$
$ make psql
...
golembic=> INSERT INTO golembic_migrations (serial_id, revision, previous) VALUES (7, 'not-in-sequence', '3196713ca7e6');
INSERT 0 1
golembic=> \q
$
$ go run ./examples/cmd/
2021-08-13T17:32:56.971927Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:32:56.97558Z     [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:32:56.981763Z    [db.migration] -- failed -- Sequence has 7 migrations but 8 are stored in the table
2021-08-13T17:32:56.981836Z    [db.migration] -- failed Finished planning migrations sequence -- Migration stored in SQL doesn't match sequence
2021-08-13T17:32:56.983864Z    [db.migration.stats] 0 applied 1 skipped 1 failed 2 total
Migration stored in SQL doesn't match sequence
exit status 1
```

### Error Mode: Milestone

During typical development, new migrations will be added over time. Sometimes
a "milestone" migration must be applied before the application can move forward
with rolling updates.

Let's simulate the time period at which the first 3 example migrations were
checked in:

```
$ make psql-reset
...
$
$
$ go run ./examples/cmd/ --length 3
2021-08-13T14:43:40.070186Z    [db.migration] -- applied -- Check table does not exist: golembic_migrations
2021-08-13T14:43:40.072975Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T14:43:40.076887Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T14:43:40.0851Z      [db.migration] -- applied -- 3f34bd961f15: Create users table
2021-08-13T14:43:40.092395Z    [db.migration] -- applied -- 464bc456c630: Seed data in users table
2021-08-13T14:43:40.099733Z    [db.migration] -- applied -- 959456a8af88: Add city column to users table
2021-08-13T14:43:40.101636Z    [db.migration.stats] 5 applied 0 skipped 0 failed 5 total
```

Then 2 more example migrations (`57393d6ddb95` and `4d07dd6af28d`) were merged
in, the first of which was a milestone:

```
$ go run ./examples/cmd/ --length 5
2021-08-13T14:46:06.134379Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T14:46:06.137363Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T14:46:06.142908Z    [db.migration] -- failed -- Revision 57393d6ddb95 (1 / 2 migrations)
2021-08-13T14:46:06.143023Z    [db.migration] -- failed Finished planning migrations sequence -- If a migration sequence contains a milestone, it must be the last migration
2021-08-13T14:46:06.144803Z    [db.migration.stats] 0 applied 1 skipped 1 failed 2 total
If a migration sequence contains a milestone, it must be the last migration
exit status 1
```

This will fail because revision `57393d6ddb95` is a milestone and so can't
be applied with **any** revisions after it. (The application needs to deploy
and stabilize with a milestone migration before the next rolling update
deploy can happen.) So the application must be deployed at a point where only
1 new migration is present:

```
$ go run ./examples/cmd/ --length 4
2021-08-13T14:48:32.80814Z     [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T14:48:32.811251Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T14:48:32.818541Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T14:48:32.830876Z    [db.migration] -- applied -- 57393d6ddb95: Rename the root user [MILESTONE]
2021-08-13T14:48:32.833259Z    [db.migration.stats] 2 applied 1 skipped 0 failed 3 total
```

and then deploy the latest changes after the previous deploy stabilizes:

```
$ go run ./examples/cmd/ --length 5
2021-08-13T14:48:50.750336Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T14:48:50.753732Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T14:48:50.76018Z     [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T14:48:50.785616Z    [db.migration] -- applied -- 4d07dd6af28d: Add index on user emails (concurrently)
2021-08-13T14:48:50.787387Z    [db.migration.stats] 2 applied 1 skipped 0 failed 3 total
```

[1]: https://godoc.org/github.com/dhermes/golembic-blend?status.svg
[2]: https://godoc.org/github.com/dhermes/golembic-blend

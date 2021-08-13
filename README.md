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
2021-08-13T17:40:01.857253Z    [db.migration] -- applied -- Check table does not exist: golembic_migrations
2021-08-13T17:40:01.860942Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:40:01.86595Z     [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T17:40:01.875813Z    [db.migration] -- applied -- 3f34bd961f15: Create users table
2021-08-13T17:40:01.884822Z    [db.migration] -- applied -- 464bc456c630: Seed data in users table
2021-08-13T17:40:01.892263Z    [db.migration] -- applied -- 959456a8af88: Add city column to users table
2021-08-13T17:40:01.900737Z    [db.migration] -- applied -- 57393d6ddb95: Rename the root user [MILESTONE]
2021-08-13T17:40:01.924805Z    [db.migration] -- applied -- 4d07dd6af28d: Add index on user emails (concurrently)
2021-08-13T17:40:01.937113Z    [db.migration] -- applied -- 2a35ccd628bc: Create books table
2021-08-13T17:40:01.944698Z    [db.migration] -- applied -- 3196713ca7e6: Create movies table
2021-08-13T17:40:01.946226Z    [db.migration.stats] 9 applied 0 skipped 0 failed 9 total
```

After creation, the next run does nothing

```
$ go run ./examples/cmd/
2021-08-13T17:40:21.716154Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:40:21.720037Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:40:21.726482Z    [db.migration] -- plan -- No migrations to run; latest revision: 3196713ca7e6
2021-08-13T17:40:21.726511Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T17:40:21.728566Z    [db.migration.stats] 1 applied 1 skipped 0 failed 2 total
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
2021-08-13T17:40:43.830474Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:40:43.833663Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:40:43.839139Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T17:40:43.849448Z    [db.migration] -- applied -- 3196713ca7e6: Create movies table
2021-08-13T17:40:43.850975Z    [db.migration.stats] 2 applied 1 skipped 0 failed 3 total
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
$ make psql-reset
...
$ go run ./examples/cmd/
...
2021-08-13T17:41:27.957464Z    [db.migration] -- applied -- 3196713ca7e6: Create movies table
2021-08-13T17:41:27.959059Z    [db.migration.stats] 9 applied 0 skipped 0 failed 9 total
$
$ make psql
...
golembic=> INSERT INTO golembic_migrations (serial_id, revision, previous) VALUES (7, 'not-in-sequence', '3196713ca7e6');
INSERT 0 1
golembic=> \q
$
$ go run ./examples/cmd/
2021-08-13T17:41:54.157961Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:41:54.161573Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:41:54.167462Z    [db.migration] -- failed Finished planning migrations sequence -- No migration registered for revision; Revision: "not-in-sequence"
2021-08-13T17:41:54.169252Z    [db.migration.stats] 0 applied 1 skipped 1 failed 2 total
No migration registered for revision; Revision: "not-in-sequence"
exit status 1
```

If the "verify history" option is used (off by default) we get slightly more
information at the cost of some more SQL queries used for verification:

```
$ go run ./examples/cmd/ --verify-history
2021-08-13T17:42:51.532835Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:42:51.535955Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:42:51.542234Z    [db.migration] -- failed -- Sequence has 7 migrations but 8 are stored in the table
2021-08-13T17:42:51.542296Z    [db.migration] -- failed Finished planning migrations sequence -- Migration stored in SQL doesn't match sequence
2021-08-13T17:42:51.543918Z    [db.migration.stats] 0 applied 1 skipped 1 failed 2 total
Migration stored in SQL doesn't match sequence
exit status 1
```

Similarly, if we can modify an existing entry in the sequence so it
becomes unknown (vs. appending to the sequence):

```
$ make psql
...
golembic=> DELETE FROM golembic_migrations WHERE revision IN ('not-in-sequence', '3196713ca7e6');
DELETE 2
golembic=> INSERT INTO golembic_migrations (serial_id, revision, previous) VALUES (6, 'not-in-sequence', '2a35ccd628bc');
INSERT 0 1
golembic=> \q
$
$ go run ./examples/cmd/
2021-08-13T17:43:26.942161Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:43:26.946179Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:43:26.95287Z     [db.migration] -- failed Finished planning migrations sequence -- No migration registered for revision; Revision: "not-in-sequence"
2021-08-13T17:43:26.955163Z    [db.migration.stats] 0 applied 1 skipped 1 failed 2 total
No migration registered for revision; Revision: "not-in-sequence"
exit status 1
$
$
$ go run ./examples/cmd/ --verify-history
2021-08-13T17:43:36.111506Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:43:36.114598Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:43:36.119419Z    [db.migration] -- failed -- Stored migration 6: "not-in-sequence:2a35ccd628bc" does not match migration "3196713ca7e6:2a35ccd628bc" in sequence
2021-08-13T17:43:36.119526Z    [db.migration] -- failed Finished planning migrations sequence -- Migration stored in SQL doesn't match sequence
2021-08-13T17:43:36.121124Z    [db.migration.stats] 0 applied 1 skipped 1 failed 2 total
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
$ go run ./examples/cmd/ --length 3
2021-08-13T17:43:55.509783Z    [db.migration] -- applied -- Check table does not exist: golembic_migrations
2021-08-13T17:43:55.512859Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:43:55.517287Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T17:43:55.52841Z     [db.migration] -- applied -- 3f34bd961f15: Create users table
2021-08-13T17:43:55.536941Z    [db.migration] -- applied -- 464bc456c630: Seed data in users table
2021-08-13T17:43:55.543169Z    [db.migration] -- applied -- 959456a8af88: Add city column to users table
2021-08-13T17:43:55.544858Z    [db.migration.stats] 5 applied 0 skipped 0 failed 5 total
```

Then 2 more example migrations (`57393d6ddb95` and `4d07dd6af28d`) were merged
in, the first of which was a milestone:

```
$ go run ./examples/cmd/ --length 5
2021-08-13T17:44:08.249128Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:44:08.253153Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:44:08.261251Z    [db.migration] -- failed -- Revision 57393d6ddb95 (1 / 2 migrations)
2021-08-13T17:44:08.261325Z    [db.migration] -- failed Finished planning migrations sequence -- If a migration sequence contains a milestone, it must be the last migration
2021-08-13T17:44:08.263539Z    [db.migration.stats] 0 applied 1 skipped 1 failed 2 total
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
2021-08-13T17:44:21.688567Z    [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:44:21.693548Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:44:21.701493Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T17:44:21.715395Z    [db.migration] -- applied -- 57393d6ddb95: Rename the root user [MILESTONE]
2021-08-13T17:44:21.717678Z    [db.migration.stats] 2 applied 1 skipped 0 failed 3 total
```

and then deploy the latest changes after the previous deploy stabilizes:

```
$ go run ./examples/cmd/ --length 5
2021-08-13T17:44:25.22514Z     [db.migration] -- skipped -- Check table does not exist: golembic_migrations
2021-08-13T17:44:25.229023Z    [db.migration] -- plan -- Determine migrations that need to be applied
2021-08-13T17:44:25.234496Z    [db.migration] -- applied -- Finished planning migrations sequence
2021-08-13T17:44:25.260623Z    [db.migration] -- applied -- 4d07dd6af28d: Add index on user emails (concurrently)
2021-08-13T17:44:25.26268Z     [db.migration.stats] 2 applied 1 skipped 0 failed 3 total
```

[1]: https://godoc.org/github.com/dhermes/golembic-blend?status.svg
[2]: https://godoc.org/github.com/dhermes/golembic-blend

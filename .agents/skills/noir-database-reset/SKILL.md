---
name: noir-database-reset
description: Reset Noir's configured PostgreSQL database and rotate its shared-world seed when the user explicitly requests a clean game-data rebuild.
---

# Noir database reset

Use this skill only for an explicit request to clear or reseed Noir's configured game database. The helper clears all 19 current game tables, including player accounts, sessions, characters, cases, and shared-world state. It preserves schemas, database credentials, and other environment settings. It rotates only `NOIR_WORLD_SEED`; starting the API later recreates the shared world and opening cases.

The reset is destructive and operates on the database selected by the API's environment configuration. Never print or expose `DATABASE_URL` or other credentials.

## Procedure

1. Locate the Noir project root and inspect its API environment-file selection in `apps/api/cmd/server/main.go`. Do not change `DATABASE_URL` or `NOIR_WORLD_SERVER_ID` as part of a reset.
2. Run the helper's dry run from `apps/api`:

   ```sh
   go run ./cmd/resetdb
   ```

   It prints the database host, database name, the number and scope of tables, and the seed file it will update. Compare the target with the user's request and the project configuration. If the target is unclear, differs from the requested database, or is a production branch without explicit authorization to erase that target, stop and ask before continuing.
3. Confirm the local API is stopped. The helper refuses to mutate while something is listening on `127.0.0.1:8080`. Do not start or restart the API as part of the reset; the user may start it afterward to generate content.
4. Proceed with `go run ./cmd/resetdb --yes` only when the user's current request explicitly authorizes wiping the displayed target and the full table scope above. A dry run alone never changes the database.
5. Report the target host and database, that all game tables were cleared, the newly written seed, and that the API was left stopped. Do not report credentials or the previous seed.

If the connection, transaction, or seed-file update fails, stop and report the error. Do not try another database, branch, or destructive method. The helper transactionally truncates the game tables and rolls back the seed-file change if the database commit fails.

## Helper

The executable helper is `apps/api/cmd/resetdb/main.go`. Keep its table allowlist aligned with schemas in `apps/api/internal/*/schema.sql`. Compile and inspect its dry-run output when changing the helper; never run it with `--yes` as a validation step.

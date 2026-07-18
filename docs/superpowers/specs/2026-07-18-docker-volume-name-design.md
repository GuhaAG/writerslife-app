# Docker Volume Name Design

## Goal

Make the WritersLife PostgreSQL data volume identifiable in Docker listings.

## Design

Keep the existing Compose service volume reference, `postgres_data`, and assign its explicit Docker engine name `writerslife_postgres_data` in the root volume declaration.

Existing automatically named volumes are not renamed or migrated. A new Compose startup creates and uses the explicit volume.

#!/bin/bash
# Run database migrations via Docker
# Usage:
#   ./migrate.sh up        - Apply all pending migrations
#   ./migrate.sh down      - Rollback last migration
#   ./migrate.sh down-all  - Rollback all migrations

CONTAINER="aiox-postgres"
DB_USER="aiox"
DB_NAME="aiox"
MIGRATIONS_DIR="./migrations"

run_sql() {
    docker exec -i "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" "$@"
}

# Ensure schema_migrations tracking table exists
run_sql <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INT PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);
SQL

case "${1:-up}" in
    up)
        for f in "$MIGRATIONS_DIR"/*.up.sql; do
            version=$(basename "$f" | cut -d_ -f1 | sed 's/^0*//')
            already=$(run_sql -tAc "SELECT 1 FROM schema_migrations WHERE version = $version" 2>/dev/null)
            if [ "$already" = "1" ]; then
                continue
            fi
            echo "Applying: $(basename "$f")"
            if run_sql < "$f"; then
                run_sql -c "INSERT INTO schema_migrations (version) VALUES ($version)" > /dev/null
            else
                echo "FAILED: $(basename "$f")"
                exit 1
            fi
        done
        echo "All migrations applied."
        ;;

    down)
        last=$(run_sql -tAc "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1" 2>/dev/null)
        if [ -z "$last" ]; then
            echo "No migrations to rollback."
            exit 0
        fi
        padded=$(printf "%06d" "$last")
        down_file=$(ls "$MIGRATIONS_DIR"/${padded}_*.down.sql 2>/dev/null)
        if [ -z "$down_file" ]; then
            echo "No down migration found for version $padded"
            exit 1
        fi
        echo "Rolling back: $(basename "$down_file")"
        if run_sql < "$down_file"; then
            run_sql -c "DELETE FROM schema_migrations WHERE version = $last" > /dev/null
            echo "Rolled back."
        else
            echo "FAILED"
            exit 1
        fi
        ;;

    down-all)
        while true; do
            last=$(run_sql -tAc "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1" 2>/dev/null)
            if [ -z "$last" ]; then
                break
            fi
            padded=$(printf "%06d" "$last")
            down_file=$(ls "$MIGRATIONS_DIR"/${padded}_*.down.sql 2>/dev/null)
            if [ -z "$down_file" ]; then
                echo "No down migration found for version $padded, stopping."
                break
            fi
            echo "Rolling back: $(basename "$down_file")"
            if run_sql < "$down_file"; then
                run_sql -c "DELETE FROM schema_migrations WHERE version = $last" > /dev/null
            else
                echo "FAILED"
                exit 1
            fi
        done
        echo "All migrations rolled back."
        ;;

    status)
        echo "Applied migrations:"
        run_sql -c "SELECT version, applied_at FROM schema_migrations ORDER BY version"
        ;;

    *)
        echo "Usage: $0 {up|down|down-all|status}"
        exit 1
        ;;
esac

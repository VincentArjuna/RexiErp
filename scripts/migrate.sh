#!/bin/bash

# RexiERP Database Migration Script
# This script handles database migrations and seeding

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-rexi_erp}
DB_USER=${DB_USER:-rexi}
DB_PASSWORD=${DB_PASSWORD:-password}
MIGRATIONS_DIR=${MIGRATIONS_DIR:-./migrations/master}

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if PostgreSQL is ready
wait_for_postgres() {
    print_status "Waiting for PostgreSQL to be ready..."

    max_attempts=30
    attempt=1

    while [ $attempt -le $max_attempts ]; do
        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" &>/dev/null; then
            print_success "PostgreSQL is ready!"
            return 0
        fi

        print_status "Attempt $attempt/$max_attempts: PostgreSQL not ready yet..."
        sleep 2
        ((attempt++))
    done

    print_error "PostgreSQL is not ready after $max_attempts attempts"
    exit 1
}

# Function to check if database exists
check_database() {
    print_status "Checking if database '$DB_NAME' exists..."

    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
        print_success "Database '$DB_NAME' exists"
        return 0
    else
        print_warning "Database '$DB_NAME' does not exist"
        return 1
    fi
}

# Function to create database
create_database() {
    print_status "Creating database '$DB_NAME'..."

    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"

    if [ $? -eq 0 ]; then
        print_success "Database '$DB_NAME' created successfully"
    else
        print_error "Failed to create database '$DB_NAME'"
        exit 1
    fi
}

# Function to run migrations
run_migrations() {
    print_status "Running database migrations..."

    # Check if migrations directory exists
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        print_error "Migrations directory '$MIGRATIONS_DIR' does not exist"
        exit 1
    fi

    # Get list of migration files sorted by filename
    migration_files=$(find "$MIGRATIONS_DIR" -name "*.sql" -type f | sort)

    if [ -z "$migration_files" ]; then
        print_warning "No migration files found in '$MIGRATIONS_DIR'"
        return 0
    fi

    # Create migrations table if it doesn't exist
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            filename VARCHAR(255) PRIMARY KEY,
            executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
    " || {
        print_error "Failed to create migrations table"
        exit 1
    }

    # Run each migration file that hasn't been executed yet
    for migration_file in $migration_files; do
        filename=$(basename "$migration_file")

        # Check if migration has already been executed
        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT filename FROM schema_migrations WHERE filename = '$filename';" | grep -q "$filename"; then
            print_status "Migration '$filename' already executed, skipping..."
            continue
        fi

        print_status "Running migration: $filename"

        # Execute migration file
        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration_file"; then
            # Record migration as executed
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "INSERT INTO schema_migrations (filename) VALUES ('$filename');"
            print_success "Migration '$filename' executed successfully"
        else
            print_error "Failed to execute migration '$filename'"
            exit 1
        fi
    done

    print_success "All migrations executed successfully"
}

# Function to run seed data
run_seed_data() {
    print_status "Running seed data..."

    # Check if seed file exists
    seed_file="$MIGRATIONS_DIR/002_seed_data.sql"
    if [ ! -f "$seed_file" ]; then
        print_warning "Seed data file not found: $seed_file"
        return 0
    fi

    # Check if seed data has already been run
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM tenants;" | grep -q "0"; then
        print_status "Running seed data..."

        if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$seed_file"; then
            print_success "Seed data executed successfully"
        else
            print_error "Failed to execute seed data"
            exit 1
        fi
    else
        print_status "Seed data already exists, skipping..."
    fi
}

# Function to reset database
reset_database() {
    print_warning "Resetting database '$DB_NAME' (all data will be lost)..."

    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Database reset cancelled"
        return 0
    fi

    # Drop and recreate database
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
    create_database
    run_migrations
    run_seed_data

    print_success "Database reset completed"
}

# Function to show migration status
show_migration_status() {
    print_status "Migration status:"

    if ! check_database; then
        print_error "Database '$DB_NAME' does not exist"
        return 1
    fi

    # Check if migrations table exists
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt schema_migrations;" | grep -q "schema_migrations"; then
        echo
        print_status "Executed migrations:"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT filename, executed_at FROM schema_migrations ORDER BY executed_at;"

        echo
        print_status "Pending migrations:"
        migration_files=$(find "$MIGRATIONS_DIR" -name "*.sql" -type f | sort)
        for migration_file in $migration_files; do
            filename=$(basename "$migration_file")
            if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT filename FROM schema_migrations WHERE filename = '$filename';" | grep -q "$filename"; then
                echo "  - $filename"
            fi
        done
    else
        print_warning "No migrations have been executed yet"
    fi
}

# Function to backup database
backup_database() {
    backup_file="backup_${DB_NAME}_$(date +%Y%m%d_%H%M%S).sql"
    print_status "Creating backup: $backup_file"

    if PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > "$backup_file"; then
        print_success "Backup created: $backup_file"

        # Compress backup
        gzip "$backup_file"
        print_success "Backup compressed: ${backup_file}.gz"
    else
        print_error "Failed to create backup"
        exit 1
    fi
}

# Function to restore database
restore_database() {
    if [ -z "$1" ]; then
        print_error "Please provide backup file path"
        echo "Usage: $0 restore <backup_file>"
        exit 1
    fi

    backup_file="$1"

    if [ ! -f "$backup_file" ]; then
        print_error "Backup file not found: $backup_file"
        exit 1
    fi

    print_warning "Restoring database from: $backup_file"
    print_warning "This will overwrite all existing data in '$DB_NAME'"

    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Database restore cancelled"
        return 0
    fi

    # Drop and recreate database
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
    create_database

    # Restore from backup
    if [[ $backup_file == *.gz ]]; then
        gunzip -c "$backup_file" | PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < "$backup_file"
    fi

    if [ $? -eq 0 ]; then
        print_success "Database restored successfully"
    else
        print_error "Failed to restore database"
        exit 1
    fi
}

# Function to show usage
show_usage() {
    echo "RexiERP Database Migration Script"
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  migrate              Run database migrations"
    echo "  seed                 Run seed data only"
    echo "  reset                Reset database (drop and recreate)"
    echo "  status               Show migration status"
    echo "  backup               Create database backup"
    echo "  restore <file>       Restore database from backup"
    echo "  init                 Initialize database (create and migrate)"
    echo "  help                 Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  DB_HOST              Database host (default: localhost)"
    echo "  DB_PORT              Database port (default: 5432)"
    echo "  DB_NAME              Database name (default: rexi_erp)"
    echo "  DB_USER              Database user (default: rexi)"
    echo "  DB_PASSWORD          Database password (default: password)"
    echo "  MIGRATIONS_DIR       Migrations directory (default: ./migrations/master)"
    echo ""
    echo "Examples:"
    echo "  $0 migrate                           # Run migrations"
    echo "  $0 seed                              # Run seed data"
    echo "  $0 init                              # Initialize database"
    echo "  $0 reset                             # Reset database"
    echo "  $0 backup                            # Create backup"
    echo "  $0 restore backup_20241001_120000.sql # Restore from backup"
}

# Main script logic
case "${1:-help}" in
    "migrate")
        wait_for_postgres
        if ! check_database; then
            create_database
        fi
        run_migrations
        ;;
    "seed")
        wait_for_postgres
        if check_database; then
            run_seed_data
        else
            print_error "Database '$DB_NAME' does not exist. Run '$0 init' first."
            exit 1
        fi
        ;;
    "init")
        wait_for_postgres
        if ! check_database; then
            create_database
        fi
        run_migrations
        run_seed_data
        print_success "Database initialization completed"
        ;;
    "reset")
        wait_for_postgres
        reset_database
        ;;
    "status")
        show_migration_status
        ;;
    "backup")
        wait_for_postgres
        if check_database; then
            backup_database
        else
            print_error "Database '$DB_NAME' does not exist"
            exit 1
        fi
        ;;
    "restore")
        wait_for_postgres
        restore_database "$2"
        ;;
    "help"|*)
        show_usage
        ;;
esac
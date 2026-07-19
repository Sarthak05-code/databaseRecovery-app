# Database Backup & Recovery Hub (`db-backup-cli`)

A production-ready, high-performance database management utility written in Go for automated MySQL/MariaDB backups and disaster recovery.

`db-backup-cli` supports two operating modes:

- A flexible, Cobra-based command-line interface for manual and cron-friendly operations.
- A local web dashboard for interactive backup and restore management.

## Features

- **Dual run modes** — Run immediate, automation-friendly commands from a terminal or start a local web dashboard.
- **Native database pipelines** — Uses the system `mysqldump` and `mysql` client utilities for database structure and data processing.
- **Streaming gzip compression** — Streams dump output directly into `.sql.gz` archives, avoiding large intermediate SQL files.
- **Secure credential handling** — Supports explicit parameters, the `DB_PASSWORD` environment variable, hidden terminal prompts, and masked web-form inputs.
- **Concurrency protection** — Uses `dbbackup_operation.lock` to prevent simultaneous operations from corrupting shared files or database state.
- **Graceful interruption handling** — Captures operating-system interrupts, including `Ctrl+C`, and cleans up locks and temporary files before exiting.
- **Automatic database provisioning** — Creates the target database during restoration when it does not already exist.
- **Fault-tolerant retries** — Retries eligible backup and restore operations up to three times when transient failures occur.
- **Structured audit logging** — Writes newline-delimited JSON records containing action types, statuses, durations, and errors to `backup_activity.log`.
- **Slack notifications** — Optionally sends operation completion notifications to a configured Slack incoming webhook.

## Project Architecture

```text
db-backup-cli/
├── cmd/
│   ├── backup.go        # CLI backup command, parameters, and flags
│   ├── restore.go       # CLI restore command and configuration
│   ├── root.go          # Cobra application initialization
│   └── server.go        # Web server and dashboard API endpoints
├── pkg/
│   ├── db/
│   │   ├── db.go        # Database interfaces and core types
│   │   └── mysql.go     # MySQL/MariaDB client execution logic
│   ├── storage/
│   │   └── storage.go   # Backup storage and directory management
│   └── utils/
│       ├── compress.go  # Streaming gzip compression and extraction
│       ├── logger.go    # Structured JSON audit logging
│       └── notifier.go  # Slack notification delivery
├── web/
│   ├── index.html       # Web dashboard interface
│   └── app.js           # Frontend logic and API requests
├── main.go              # Application entry point
├── go.mod               # Go module definition
└── backup_activity.log  # Generated audit log; created at runtime
```

> The exact structure may vary if additional packages, tests, or configuration files are added.

## Requirements

### Go

Install a current stable release of the [Go programming language](https://go.dev/dl/).

### MySQL or MariaDB client tools

The following executables must be installed and available through the operating system's `PATH`:

- `mysql`
- `mysqldump`

Verify the installation with:

```bash
mysql --version
mysqldump --version
```

### Linux — Debian or Ubuntu

```bash
sudo apt update
sudo apt install -y mysql-client
```

### macOS — Homebrew

```bash
brew install mysql-client
```

Homebrew may print instructions for adding the client binaries to your `PATH`. Follow those instructions if the commands are not found after installation.

### Windows

Install MySQL Community Server, MySQL Workbench, or another compatible MySQL client package. Add the directory containing `mysql.exe` and `mysqldump.exe` to `PATH`.

A common installation path is:

```text
C:\Program Files\MySQL\MySQL Server 8.0\bin
```

The path may differ depending on the installed MySQL version.

## Build and Installation

Clone or copy the project, then change to its root directory—the directory containing `main.go` and `go.mod`.

Download and resolve dependencies:

```bash
go mod tidy
```

If the project uses terminal password prompts through `golang.org/x/term` and the dependency is not already declared in `go.mod`, add it with:

```bash
go get golang.org/x/term
go mod tidy
```

### Build for Windows

```bash
go build -o dbbackup.exe .
```

### Build for Linux or macOS

```bash
go build -o dbbackup .
```

## Authentication and Password Security

Avoid passing passwords with `--pass` whenever possible. Command-line arguments can be visible to other processes or stored in shell history.

### Environment variable

Set `DB_PASSWORD` for the current shell session.

#### Linux/macOS

```bash
export DB_PASSWORD='secure_password_here'
```

#### Windows PowerShell

```powershell
$env:DB_PASSWORD = 'secure_password_here'
```

#### Windows Command Prompt

```cmd
set DB_PASSWORD=secure_password_here
```

### Hidden interactive prompt

If `DB_PASSWORD` is not set and no password is provided with `--pass`, the application prompts for the password without displaying the typed characters:

```text
Enter Database Password:
```

## Usage

After building the project, use `./dbbackup` on Linux/macOS or `dbbackup.exe` on Windows.

### Start the web dashboard

Start the local web application with:

```bash
./dbbackup server
```

On Windows:

```powershell
.\dbbackup.exe server
```

Expected console output:

```text
Dashboard UI running at http://localhost:8080
```

Open [http://localhost:8080](http://localhost:8080) in a browser.

> The dashboard is intended to run locally unless the application explicitly provides authentication and network-access controls. Do not expose an unauthenticated administrative interface to an untrusted network.

### Create a compressed backup

```bash
./dbbackup backup -u <username> -d <database_name> [flags]
```

On Windows:

```powershell
.\dbbackup.exe backup -u <username> -d <database_name> [flags]
```

#### Backup options

| Short | Long | Required | Default | Description |
|---|---|:---:|---|---|
| `-u` | `--user` | Yes | — | Database username, such as `root`. |
| `-d` | `--db` | Yes | — | Name of the database to back up. |
| `-p` | `--pass` | No | — | Database password. Prefer `DB_PASSWORD` or the hidden prompt. |
| `-H` | `--host` | No | `localhost` | Database server hostname or IP address. |
| `-P` | `--port` | No | `3306` | Database server port. |
| `-o` | `--output` | No | `./backups` | Directory in which to store backup archives. |
| `-s` | `--slack` | No | — | Slack incoming webhook URL for completion notifications. |

Example:

```bash
./dbbackup backup -u root -d shop_db -o ./my_backups
```

Example output:

```text
Enter Database Password:
Success! Backup archive safely stored at: my_backups/shop_db_20260719_183501.sql.gz (Time taken: 1.1425s)
```

Backup archives use a timestamped naming convention similar to:

```text
<database>_YYYYMMDD_HHMMSS.sql.gz
```

### Restore a database archive

Restore either an uncompressed `.sql` file or a gzip-compressed `.sql.gz` file:

```bash
./dbbackup restore -u <username> -d <database_name> -f <archive_path>
```

On Windows:

```powershell
.\dbbackup.exe restore -u <username> -d <database_name> -f <archive_path>
```

#### Restore options

| Short | Long | Required | Default | Description |
|---|---|:---:|---|---|
| `-u` | `--user` | Yes | — | Database username. |
| `-d` | `--db` | Yes | — | Target database for restoration. |
| `-f` | `--file` | Yes | — | Path to a `.sql` or `.sql.gz` archive. |
| `-p` | `--pass` | No | — | Database password. Prefer `DB_PASSWORD` or the hidden prompt. |

Example:

```bash
./dbbackup restore \
  -u root \
  -d shop_db \
  -f ./backups/shop_db_20260719_183501.sql.gz
```

Example output:

```text
Enter Database Password:
Attempting database restore...
Database restored successfully
```

Before restoring production data, verify that the archive is valid and that the target database and credentials are correct. A restore may overwrite existing data depending on the contents of the SQL dump.

## Audit Logging

Each operation appends a newline-delimited JSON object to `backup_activity.log`:

```json
{"timestamp":"2026-07-19T18:35:02+05:45","action":"BACKUP","status":"SUCCESS","duration":"1.142512s"}
{"timestamp":"2026-07-19T18:42:11+05:45","action":"RESTORE","status":"SUCCESS","duration":"2.410985s"}
```

Each log record may contain additional fields when an operation fails, such as an error message or retry information.

Inspect the log with:

#### Windows Command Prompt

```cmd
type backup_activity.log
```

#### Linux/macOS

```bash
cat backup_activity.log
```

Protect the audit log because operational errors or notification details may contain sensitive environment information.

## Locking and Cleanup

The application uses `dbbackup_operation.lock` to prevent multiple backup or restore operations from running concurrently.

If a lock remains after an interrupted process:

1. Confirm that no `dbbackup` process is still running.
2. Confirm that no backup or restore operation is actively writing data.
3. Remove the stale lock only after both checks are complete.

## Notifications

A Slack incoming webhook can be supplied for backup completion notifications using the `--slack` option:

```bash
./dbbackup backup \
  -u root \
  -d shop_db \
  --slack "$SLACK_WEBHOOK_URL"
```

Store webhook URLs in protected environment variables or a secret manager. Do not commit webhook URLs to source control.

## Operational Recommendations

- Grant the database user only the permissions required for backup and restoration.
- Store backups on a separate disk, host, or remote storage provider.
- Test restoration regularly instead of assuming that a successful backup is restorable.
- Restrict permissions on the backup directory, lock file, and audit log.
- Exclude `.sql`, `.sql.gz`, `.lock`, and sensitive log files from source control.
- Do not expose the web dashboard outside a trusted network without authentication and appropriate access controls.

## Roadmap.sh Challenge

This project was designed with reference to the [Roadmap.sh Database Backup Utility Project](https://roadmap.sh/projects/database-backup-utility).

## License

This project is licensed under the MIT License. See the [`LICENSE`](LICENSE) file for the full license text.

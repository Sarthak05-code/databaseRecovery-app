# Database Backup & Recovery CLI Tool (`dbbackup`)

A production-ready, high-performance command-line utility written in Go for automated MySQL/MariaDB backups and disaster recovery. `dbbackup` supports full database dumps, streaming gzip compression, structured JSON audit trails, retry handling, secure password input, and restore workflows.

## Features

- **Native database dumps** — Uses the native `mysqldump` and `mysql` client utilities for reliable schema and data extraction.
- **Streaming gzip compression** — Compresses backup data on the fly and writes directly to `.sql.gz` archives without creating large intermediate SQL files in memory.
- **Secure password handling** — Resolves credentials from the `DB_PASSWORD` environment variable or prompts for the password through hidden terminal input. Passwords should not be passed on the command line in production.
- **Concurrency protection** — Uses an active `.lock` file to prevent concurrent operations from causing race conditions or data corruption.
- **Graceful interruption handling** — Listens for operating-system interrupts, including `Ctrl+C`, and cleans up temporary resources and locks before exiting.
- **Automatic database provisioning** — Creates the target database during restoration when it does not already exist.
- **Transient fault resilience** — Retries backup and restore operations up to three times with exponential backoff after temporary failures.
- **Structured audit logging** — Records actions, timestamps, statuses, durations, and performance metrics in `backup_activity.log` as JSON entries.
- **Slack notifications** — Optionally sends completion notifications to a Slack incoming webhook.

## Requirements

### Go

Install a recent version of [Go](https://go.dev/dl/).

### MySQL or MariaDB client utilities

The following commands must be available on your system's `PATH`:

- `mysql`
- `mysqldump`

#### Windows

Install MySQL or MySQL Workbench and ensure the directory containing `mysql.exe` and `mysqldump.exe` is included in `PATH`. For example:

```text
C:\Program Files\MySQL\MySQL Server 8.0\bin
```

> The exact installation path may vary depending on the MySQL version and installation method.

#### Linux

Install the MySQL client package using your distribution's package manager. For Debian or Ubuntu:

```bash
sudo apt update
sudo apt install mysql-client
```

#### macOS

Install the MySQL client with Homebrew:

```bash
brew install mysql-client
```

You may need to add the Homebrew client binaries to your `PATH`.

## Installation and Compilation

From the project root—the directory containing `main.go`—download the terminal input dependency and build the executable:

```bash
go get golang.org/x/term
go mod tidy
```

### Windows

```bash
go build -o dbbackup.exe .
```

### Linux and macOS

```bash
go build -o dbbackup .
```

## Authentication

The password can be supplied in one of two secure ways.

### Environment variable

Set `DB_PASSWORD` before running the command:

#### Windows PowerShell

```powershell
$env:DB_PASSWORD = "your-database-password"
```

#### Windows Command Prompt

```cmd
set DB_PASSWORD=your-database-password
```

#### Linux/macOS

```bash
export DB_PASSWORD='your-database-password'
```

### Hidden interactive prompt

If `DB_PASSWORD` is not set and no password is provided with `--pass`, the tool prompts for the password using hidden terminal input:

```text
Enter Database Password:
```

> Avoid passing passwords with `--pass` where possible because command-line arguments may be visible to other processes or stored in shell history.

## Usage

### Create a compressed backup

Run a full backup with a timestamped `.sql.gz` archive:

```bash
dbbackup backup -u <username> -d <database_name>
```

#### Backup options

| Option | Long form  | Required | Default     | Description                                                   |
| ------ | ---------- | :------: | ----------- | ------------------------------------------------------------- |
| `-u`   | `--user`   |   Yes    | —           | Database username, such as `root`.                            |
| `-d`   | `--db`     |   Yes    | —           | Name of the database to back up.                              |
| `-p`   | `--pass`   |    No    | —           | Database password. Prefer `DB_PASSWORD` or the hidden prompt. |
| `-H`   | `--host`   |    No    | `localhost` | Database server host.                                         |
| `-P`   | `--port`   |    No    | `3306`      | Database server port.                                         |
| `-o`   | `--output` |    No    | `./backups` | Directory in which to store backup archives.                  |
| `-s`   | `--slack`  |    No    | —           | Slack incoming webhook URL for completion notifications.      |

#### Example

```bash
dbbackup backup -u root -d company_db
```

Example output:

```text
Enter Database Password:
Success! Backup archive safely stored at: backups/company_db_20260719_125816.sql.gz (Time taken: 3.9753092s)
```

To specify a custom host, port, and output directory:

```bash
dbbackup backup \
  --user backup_user \
  --db company_db \
  --host 127.0.0.1 \
  --port 3306 \
  --output ./backups
```

### Restore an archive

Restore an existing `.sql` or `.sql.gz` archive into a database:

```bash
dbbackup restore -u <username> -d <database_name> -f <archive_path>
```

#### Restore options

| Option | Long form | Required | Default | Description                                                   |
| ------ | --------- | :------: | ------- | ------------------------------------------------------------- |
| `-u`   | `--user`  |   Yes    | —       | Database username, such as `root`.                            |
| `-d`   | `--db`    |   Yes    | —       | Target database into which the archive will be restored.      |
| `-f`   | `--file`  |   Yes    | —       | Path to a `.sql` or `.sql.gz` archive.                        |
| `-p`   | `--pass`  |    No    | —       | Database password. Prefer `DB_PASSWORD` or the hidden prompt. |

#### Example

```bash
dbbackup restore \
  -u root \
  -d company_db \
  -f backups/company_db_20260719_125816.sql.gz
```

Example output:

```text
Enter Database Password:
Attempting database restore...
Database restored successfully
```

During restoration, the tool automatically creates the target database if it does not exist, then streams and imports the archive contents.

## Audit Logging

Each backup and restore operation appends a structured JSON record to `backup_activity.log`:

```json
{
  "timestamp": "2026-07-19T12:58:17+05:45",
  "action": "BACKUP",
  "status": "SUCCESS",
  "duration": "3.9753092s"
}
```

Inspect the audit log with the appropriate terminal command:

### Windows Command Prompt

```cmd
type backup_activity.log
```

### Linux/macOS

```bash
cat backup_activity.log
```

The log can be ingested by monitoring, compliance, or centralized logging systems that support newline-delimited JSON records.

## Backup File Naming

Backups are written using a timestamped naming convention similar to:

```text
<database>_YYYYMMDD_HHMMSS.sql.gz
```

Example:

```text
backups/company_db_20260719_125816.sql.gz
```

## Operational Safety

- Verify that the database user has the permissions required for dumping, creating databases, and restoring data.
- Protect the backup directory and audit log with appropriate filesystem permissions.
- Store backups on a separate disk or remote storage location for protection against host failure.
- Test restoration regularly; a backup is only useful if it can be restored successfully.
- Do not commit `.sql`, `.sql.gz`, password files, or audit logs containing sensitive information to source control.
- Configure Slack webhook URLs through a secret manager or protected environment variable rather than committing them to scripts or repositories.

## Troubleshooting

### `mysql` or `mysqldump` cannot be found

Confirm that both client binaries are installed and available on `PATH`:

```bash
mysql --version
mysqldump --version
```

### Authentication fails

Check the username, host, port, and password. If using an environment variable, verify that it is set in the same shell session that launches `dbbackup`.

### A lock file remains after an interrupted run

First confirm that no other `dbbackup` process is still running. Only remove the lock file after verifying that the previous process has terminated and that no operation is still writing to the backup or restore target.

## Roadmap

Did this project as a challennge for roadmap
https://roadmap.sh/projects/database-backup-utility

## License

empty for now

# Database Backup & Recovery CLI Tool (`dbbackup`)

A production-ready, high-performance command-line utility built in Go designed to handle automated full backups, on-the-fly streaming compression, structured JSON audit trails, and instant disaster recovery pipelines for MySQL/MariaDB databases.

## 🚀 Key Features

*   **Zero-Overhead Dumps:** Hooks directly into native client utilities (`mysqldump` and `mysql`) for reliable schema structural locks and data extracts.
*   **On-the-Fly Gzip Pipelines:** Uses low-memory streaming compression and inflation algorithms to read and write directly to `.sql.gz` archives without creating massive intermediate raw text buffers.
*   **Structured Audit Trails:** Logs every action, timeframe, execution status, and granular performance metric directly into a localized JSON compliance file (`backup_activity.log`).
*   **Built-in Disaster Recovery:** Automated detection and transparent extraction of binary gzip streams straight into target schemas for seamless recovery loops.
*   **Multi-Channel Integrations:** Built-in hooks to push completion notifications directly to external applications (like Slack webhooks).

---

## 🛠️ Requirements & Installation

### 1. Prerequisite Dependencies
Ensure that the database client binaries (`mysql` and `mysqldump`) are available on your machine's environment variables (`PATH`).
*   **Windows (MySQL Workbench):** Ensure the path `C:\Program Files\MySQL\MySQL Workbench 8.0` is registered.
*   **Linux/macOS:** Ensure client packages (`mysql-client`) are available via your system package managers (`apt`, `brew`).

### 2. Compilation
Compile the project from the root folder containing the `main.go` entry point:

```bash
# Compile for Windows environments
go build -o dbbackup.exe .

# Compile for Unix/Linux environments
go build -o dbbackup .
```

---

## 💻 CLI Commands & Syntax

### 1. Perform a Full Compressed Backup
Triggers a backup job, compresses the output to a timestamped file inside the tracking folder, and writes metrics to the audit trail log.

```bash
dbbackup.exe backup -u <username> -p <password> -d <database_name> [flags]
```

**Parameters & Flags:**
*   `-u`, `--user`: Database user account (e.g., `root`)
*   `-p`, `--pass`: Database connection password
*   `-d`, `--db`: Name of the target schema database to back up
*   `-H`, `--host`: Database engine host address *(Default: `localhost`)*
*   `-P`, `--port`: Database connection listening port *(Default: `3306`)*
*   `-o`, `--output`: Target path directory to store compressed archives *(Default: `./backups`)*
*   `-s`, `--slack`: Optional incoming Slack Webhook App URL for remote tracking alerts

#### 📝 Example:
```bash
dbbackup.exe backup -u root -p 123456 -d company_db
```
**Output Response:**
```text
Success! Backup archive safely stored at: backups\company_db_20260719_122758.sql.gz (Time taken: 101.2288ms)
```

---

### 2. Restore an Archive (Disaster Recovery)
Automatically evaluates the target archive string format, transparently extracts the data stream, and imports the structure directly back into the target schema database.

```bash
dbbackup.exe restore -u <username> -p <password> -d <database_name> -f <archive_path>
```

**Parameters & Flags:**
*   `-u`, `--user`: Database user account (e.g., `root`)
*   `-p`, `--pass`: Database connection password
*   `-d`, `--db`: Name of the target schema database to pour data into
*   `-f`, `--file`: Path to the targeted `.sql` or `.sql.gz` backup archive file to execute

#### 📝 Example:
```bash
dbbackup.exe restore -u root -p 123456 -d company_db -f backups\company_db_20260719_122758.sql.gz
```
**Output Response:**
```text
Attempting database restore...
Database restored successfully
```

> ⚠️ *Note: The destination database shell must exist before executing a data restore loop. If the database was fully dropped, run `CREATE DATABASE database_name;` inside your engine manager prior to triggering the recovery CLI tool.*

---

## 📊 Compliance Auditing & Log Structures

Every lifecycle action translates metrics cleanly into a structured JSON entry within `backup_activity.log` to track operation windows and performance trends:

```json
{"timestamp":"2026-07-19T12:27:59+05:45","action":"BACKUP","status":"SUCCESS","duration":"101.2288ms"}
```

Use standard terminal inspection utilities to check operational health histories:
*   **Windows CMD:** `type backup_activity.log`
*   **Unix/Bash:** `cat backup_activity.log`
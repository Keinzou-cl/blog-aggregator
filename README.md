# Gator

Gator is a command-line RSS feed aggregator written in Go. It allows users to register accounts, add RSS feeds, aggregate posts from those feeds, and browse the collected posts from the terminal.

## Requirements

Before running Gator, make sure you have:

- [Go](https://go.dev/) installed
- [PostgreSQL](https://www.postgresql.org/) installed and running
- A PostgreSQL database for Gator

## Installation

Clone the repository and navigate to its root directory:

```bash
git clone <your-repository-url>
cd gator
```

Install the Gator CLI with Go:

```bash
go install .
```

After installation, make sure Go's binary directory is included in your `PATH`. You can then run:

```bash
gator
```

## Configuration

Gator requires a configuration file containing your PostgreSQL connection string and current username.

Create the configuration file expected by Gator and provide your PostgreSQL connection string.

The configuration should contain values similar to:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace the database connection details with those for your PostgreSQL installation.

The `current_user_name` value is updated when you use the appropriate Gator user command.

## Running Gator

Once PostgreSQL is running and your configuration is set up, you can run Gator from your terminal.

### Register a user

```bash
gator register <username>
```

### Log in as a user

```bash
gator login <username>
```

### Add a feed

```bash
gator addfeed <name> <url>
```

### View available feeds

```bash
gator feeds
```

### Aggregate feeds

The `agg` command periodically fetches posts from the registered feeds.

For example, to fetch feeds every minute:

```bash
gator agg 1m
```

You can use other Go duration formats, such as:

```bash
gator agg 1s
gator agg 1m
gator agg 1h
```

The aggregator runs continuously until the program is stopped.

### Browse posts

Browse the latest posts collected for the current user:

```bash
gator browse
```

By default, `browse` displays 2 posts.

You can optionally specify the number of posts:

```bash
gator browse 5
```

## Example Workflow

A typical first-time setup might look like:

```bash
gator register alice
gator login alice
gator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml
gator feeds
gator agg 1m
```

While the aggregator is running, Gator periodically fetches posts from the configured feeds.

In another terminal, you can browse the collected posts:

```bash
gator browse
```

Or specify a larger number:

```bash
gator browse 10
```

## Stopping the Aggregator

The `agg` command runs continuously. Press:

```text
Ctrl+C
```

to stop it.

## Built With

- Go
- PostgreSQL
- SQLC
- Goose
- RSS
- UUID
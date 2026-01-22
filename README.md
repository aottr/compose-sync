# compose-sync

A tool to automatically sync and deploy Docker Compose stacks from a git repository, with multi-host support.

## Overview

compose-sync pulls changes from a git repository containing Docker Compose files and only deploys stacks that:
1. Have changed since the last pull
2. Are assigned to the current host

## Repository Structure

Your git repository should have the following structure:

```
stacks/
    traefik/compose.yml
    uptime-kuma/compose.yml
    home-assistant/compose.yml
    ...

inventory.yml
```

The `inventory.yml` file at the root contains all hosts and their assigned stacks. For example:

```yaml
hosts:
  vps-1:
    - traefik
    - uptime-kuma
  nas-1:
    - home-assistant
    - nextcloud
```

This format is compatible with Ansible inventory structures and provides a centralized view of all host assignments.

## Installation

```bash
go install github.com/aottr/compose-sync@latest
```

#### Alternatively build from source

1. Clone this repository
2. Build the application:
   ```bash
   go build -o compose-sync
   ```

## Configuration

1. Copy `config.yml.example` to `config.yml`
2. Edit `config.yml` and set `repo_path` to the local path of your git repository

## Usage

### Basic Usage

```bash
./compose-sync
```

This will:
- Detect the current hostname
- Pull the latest changes from git
- Find changed stacks
- Deploy only changed stacks assigned to this host

### Dry Run

To see what would be deployed without actually deploying:

```bash
./compose-sync -dry-run
```

### Custom Config Path

```bash
./compose-sync -config /path/to/config.yml
```

## How It Works

1. **Host Detection**: The tool uses the system hostname to identify the current host
2. **Stack Assignment**: Reads `inventory.yml` to determine which stacks should be deployed on this host
3. **Change Detection**: Compares git commits before and after pulling to find changed stacks
4. **Selective Deployment**: Only deploys stacks that both changed AND are assigned to this host

## Requirements

- Go 1.21 or later
- Git
- Docker and Docker Compose
- A git repository with the structure described above

## License

MIT


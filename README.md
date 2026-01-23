# compose-sync

A tool to automatically sync and deploy Docker Compose stacks from a git repository, with multi-host support.

## Overview

compose-sync pulls changes from a git repository containing Docker Compose files and only deploys stacks that:
1. Have changed since the last pull
2. Are assigned to the current host

## How It Works

1. **Host Detection**: The tool uses the system hostname to identify the current host
2. **Stack Assignment**: Reads `inventory.yml` to determine which stacks should be deployed on this host
3. **Change Detection**: Compares git commits before and after pulling to find changed stacks
4. **Selective Deployment**: Only deploys stacks that both changed AND are assigned to this host

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

## Requirements

- Go 1.21 or later
- Git
- Docker and Docker Compose
- A git repository with the structure described above

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

### Custom Config Path

```bash
./compose-sync -config /path/to/config.yml
```

## Automatic Execution

### Systemd Service and Timer

To run compose-sync automatically using systemd:

1. **Install the binary** (if not already installed):
   ```bash
   sudo cp compose-sync /usr/local/bin/
   ```

2. **Edit the service file** (`compose-sync.service`):
   - Replace `youruser` with the actual user that should run compose-sync (make sure this user has the correct permissions)
   - Adjust paths if needed (binary location, config file, etc.)

3. **Copy the service and timer files**:
   ```bash
   sudo cp compose-sync.service compose-sync.timer /etc/systemd/system/
   ```

4. **Reload systemd and enable the timer**:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable compose-sync.timer
   sudo systemctl start compose-sync.timer
   ```

5. **Check status**:
   ```bash
   sudo systemctl status compose-sync.timer
   sudo systemctl status compose-sync.service
   ```

6. **View logs**:
   ```bash
   sudo journalctl -u compose-sync.service -f
   ```

The timer will run compose-sync every 5 minutes. To change the interval, edit `compose-sync.timer` and modify the `OnUnitActiveSec` value.

### Cron Job (Alternative)

To run compose-sync using cron instead of systemd:

1. **Install the cron job**:
   ```bash
   crontab -e
   ```
   Then add the desired crontab job, for example (every 5min, defaults, logging):
   ```bash
   */5 * * * * /usr/local/bin/compose-sync >> /var/log/compose-sync.log 2>&1
   ```

## License

MIT


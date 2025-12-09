# Pyre

Polymarket position tracker that monitors positions and PnL for configured users.

## Running

```bash
./pyre --config config.yaml
```

## Configuration

Create a `config.yaml` file:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "./data/pyre.db"

sync:
  intervalMinutes: 5

personas:
  stuart:
    displayName: "UserA"
    usernames:
      SomePolyMarketUser:
        - "0xfd....." # Replace with your own address
```

## Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v ./config.yaml:/config.yaml \
  -v ./data:/data \
  pyre --config /config.yaml
```

### Volumes

| Path | Description |
|------|-------------|
| `/config.yaml` | Configuration file |
| `/data` | SQLite database directory (persist this to retain data) |

## Building

```bash
# Build frontend
cd frontend && pnpm install && pnpm build

# Build backend (embeds frontend)
cd backend && go build -o ../pyre ./cmd/server
```

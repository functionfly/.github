# IoT Runtime

The IoT runtime bridges constrained and edge devices (sensors, actuators, gateways, robots) to the FunctionFly orchestrator and SAR Rust runtime.

## Architecture

```
IoT Device → MQTT/COAP → IoT Bridge → GBA Auth → NATS → SAR Runtime
                                                      ↓
                                                Go Orchestrator
```

## NATS Subject Schema

The bridge publishes events to subjects with the pattern `iot.device.<device_id>.<event_type>`:

| Subject | Purpose |
|---------|---------|
| `iot.device.<id>.telemetry` | Sensor readings |
| `iot.device.<id>.command`  | Outbound commands |
| `iot.device.<id>.status`   | Online/offline state |
| `iot.device.<id>.error`    | Error events |
| `iot.device.<id>.state`    | Device state updates |
| `iot.device.<id>.observe`  | COAP observation updates |

SAR can subscribe to all device events with the wildcard pattern: `iot.device.*.*`.

## Components

| Package | Purpose |
|---------|---------|
| `internal/auth/iot/` | GBA plugin: X.509, JWT, PSK device auth |
| `internal/iot/`      | MQTT/COAP bridge with NATS publishing |
| `internal/wasm/iot_fixtures.go` | Pre-compiled WASM test fixtures |
| `internal/storage/sql/iot/` | Device registry persistence |

## Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `IOT_MQTT_PORT` | 1883 | MQTT listen port |
| `IOT_COAP_PORT` | 5683 | COAP listen port |
| `IOT_USE_EMBEDDED_BROKER` | true | Use mochi-mqtt vs external broker |
| `IOT_EXTERNAL_BROKER` | - | External MQTT URL when not embedded |
| `IOT_AUTH_ENABLED` | false | Enable X.509/JWT/PSK device auth |
| `IOT_JWT_SECRET` | - | Secret for JWT validation |
| `NATS_URL` | nats://localhost:4222 | NATS server URL |

## Device Auth Tiers

| Device Type | Auth Method |
|-------------|-------------|
| Edge gateways | X.509 + mTLS |
| Intermediate devices | JWT tokens |
| Constrained sensors | PSK |

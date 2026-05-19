# Enterprise / Self-Hosted Configuration

qctx has no defaults to public-cloud endpoints; every URL is configurable.

| Source | Env var | Flag |
|---|---|---|
| SonarQube | `SONAR_HOST_URL` | `--sonar-url` |
| GitLab | `GITLAB_HOST_URL` | `--gitlab-url` |

## Tokens

| Source | Env | Flag | File |
|---|---|---|---|
| Sonar | `SONAR_TOKEN` | `--sonar-token` | `--sonar-token-file` |
| GitLab | `GITLAB_TOKEN` | `--gitlab-token` | `--gitlab-token-file` |

## Corporate CA

```bash
qctx fetch --ca-cert /etc/ssl/corp/ca-bundle.pem ...
# or:
export SSL_CERT_FILE=/etc/ssl/corp/ca-bundle.pem
```

## Proxy

`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` honored automatically.

## SSO proxies / extra headers

```bash
qctx fetch --header 'X-Forwarded-User: ci-bot' --header 'X-SSO-Group: ci' ...
```

## Insecure (dev only)

`--insecure` disables TLS verification. Stderr warning every run.

## Config file

Long-lived per-environment defaults can live in a YAML file. Resolution priority is **flag > env > file > error**.

```yaml
# ~/.qctx.yaml (or pass via --config PATH / QCTX_CONFIG=PATH)
sonar_url: https://sonar.example.com
gitlab_url: https://gitlab.example.com
sonar_token: ${SONAR_TOKEN:-}   # YAML doesn't expand; leave secrets to env or --*-token-file
ca_cert: /etc/ssl/corp/ca-bundle.pem
insecure: false
```

| Source | Flag | Env |
|---|---|---|
| Config path | `--config PATH` | `QCTX_CONFIG` (default: `~/.qctx.yaml`) |

Tokens in a config file are loaded verbatim — prefer `--*-token-file` or env vars in production so they do not land on disk in plaintext.

## Log level

```bash
qctx fetch --log-level debug ...
# or:
QCTX_LOG_LEVEL=debug qctx fetch ...
```

Levels: `debug`, `info` (default), `warn`, `error`. Logs go to stderr; stdout stays clean for the JSON payload.

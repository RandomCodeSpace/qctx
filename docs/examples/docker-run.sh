#!/usr/bin/env bash
set -euo pipefail

# Example: run qctx in Docker against a self-hosted SonarQube/GitLab.

docker run --rm \
  -e SONAR_HOST_URL=https://sonar.example.com \
  -e SONAR_TOKEN \
  -e GITLAB_HOST_URL=https://gitlab.example.com \
  -e GITLAB_TOKEN \
  -v "$(pwd)/nexus-iq-report.json:/in/nexus.json:ro" \
  -v "$(pwd)/ca-bundle.pem:/etc/ssl/corp/ca.pem:ro" \
  registry.example.com/qctx:0.1.0 \
  fetch \
  --ca-cert /etc/ssl/corp/ca.pem \
  --mr "${1:-https://gitlab.example.com/team/my-svc/-/merge_requests/42}" \
  --nexus-report /in/nexus.json

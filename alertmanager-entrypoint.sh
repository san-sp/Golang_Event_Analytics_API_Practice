#!/bin/sh
set -eu

: "${SMTP_FROM:?SMTP_FROM is required}"
: "${SMTP_USERNAME:?SMTP_USERNAME is required}"
: "${SMTP_PASSWORD:?SMTP_PASSWORD is required}"
: "${ALERT_EMAIL_TO:?ALERT_EMAIL_TO is required}"

cat > /tmp/alertmanager.yml <<EOF
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: '${SMTP_FROM}'
  smtp_auth_username: '${SMTP_USERNAME}'
  smtp_auth_password: '${SMTP_PASSWORD}'
  smtp_require_tls: true

route:
  receiver: 'email-alerts'
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 1m
  repeat_interval: 4h

receivers:
  - name: 'email-alerts'
    email_configs:
      - to: '${ALERT_EMAIL_TO}'
        send_resolved: true
EOF

exec /bin/alertmanager \
  --config.file=/tmp/alertmanager.yml \
  --web.listen-address=0.0.0.0:9093

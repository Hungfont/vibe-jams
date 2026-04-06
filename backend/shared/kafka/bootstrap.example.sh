#!/usr/bin/env bash
set -euo pipefail

# Example bootstrap flow for Phase 1 Kafka topics and ACLs.
# Required env vars:
#   KAFKA_BOOTSTRAP_SERVERS
#   KAFKA_ADMIN_CONFIG (client properties file)

create_topic() {
  local name="$1"
  local partitions="$2"
  local retention_ms="$3"

  kafka-topics.sh \
    --bootstrap-server "${KAFKA_BOOTSTRAP_SERVERS}" \
    --command-config "${KAFKA_ADMIN_CONFIG}" \
    --create --if-not-exists \
    --topic "${name}" \
    --partitions "${partitions}" \
    --replication-factor 3 \
    --config "retention.ms=${retention_ms}"
}

add_acl() {
  local principal="$1"
  local operation="$2"
  local topic="$3"

  kafka-acls.sh \
    --bootstrap-server "${KAFKA_BOOTSTRAP_SERVERS}" \
    --command-config "${KAFKA_ADMIN_CONFIG}" \
    --add \
    --allow-principal "User:${principal}" \
    --operation "${operation}" \
    --topic "${topic}"
}

create_topic "jam.session.events" "12" "604800000"
create_topic "jam.queue.events" "24" "604800000"
create_topic "jam.playback.events" "12" "604800000"
create_topic "jam.moderation.events" "12" "604800000"
create_topic "analytics.user.actions" "24" "1209600000"

# Producer ACLs
add_acl "svc-jam-service" "WRITE" "jam.session.events"
add_acl "svc-jam-service" "WRITE" "jam.queue.events"
add_acl "svc-jam-service" "WRITE" "jam.moderation.events"
add_acl "svc-playback-service" "WRITE" "jam.playback.events"
add_acl "svc-api-service" "WRITE" "analytics.user.actions"

# Consumer ACLs
add_acl "svc-rt-gateway" "READ" "jam.queue.events"
add_acl "svc-rt-gateway" "READ" "jam.playback.events"
add_acl "svc-rt-gateway" "READ" "jam.moderation.events"

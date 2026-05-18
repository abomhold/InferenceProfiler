#!/usr/bin/env bash
ENV_FILE=configs/default.env
for l in {1..100}; do
  echo "Starting Loop: ${l}"
  for az in {0..2}; do
    echo "Starting AZ: ${az}"
    if ! make infra-down; then
      echo "Failed infra teardown"
    fi

    sed -i "s|^TF_VAR_AVAILABILITY_ZONE=.*|TF_VAR_AVAILABILITY_ZONE=${az}|" "${ENV_FILE}"
    if ! timeout 5m make infra-up; then
      echo "Failed infra Bootstrap"
      continue
    fi

    ALL_IPS=$(tofu output -raw all_ips | tr ',' ' ')
    for IP in $ALL_IPS; do
      echo "Waiting for port 22 on ${IP}..."
      until timeout 1 bash -c "</dev/tcp/${IP}/22" 2>/dev/null; do
        sleep 2
      done
    done

    if ! timeout 5m make deploy; then
      echo "Failed infra deploy"
      continue
    fi

    echo "Waiting on cloud-init and vllm start up"
    SERVER_IP=$(tofu output -raw server_ip)
    if ! curl -s -f --retry-connrefused --retry-all-errors --retry 500 --retry-delay 2 --retry-max-time 1000 \
        "http://${SERVER_IP}:8000/health"; then
      echo "Failed service check"
      continue
    fi

    if ! make start_bench; then
      echo "Failed benchmark"
      continue
    fi

    if ! make pull-results; then
      echo "Failed result collection"
      continue
    fi
  done
done
if ! make infra-down; then
  echo "Failed infra teardown"
fi
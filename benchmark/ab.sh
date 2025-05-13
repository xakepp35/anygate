#!/bin/bash

# 🎯 Точки тестирования (через опубликованные хост-порты!)
TARGETS=(
  "echo:80"
  "anygate:80"
  "haproxy:80"
  "nginx-dark:80"
  "nginx:80"
)

# 🔧 Настройки
TOTAL_REQUESTS=200000
CONCURRENCY=1000

for target in "${TARGETS[@]}"; do
  echo -e "\n🎯 TARGET: $target"
  docker compose exec ab ab -n "$TOTAL_REQUESTS" -c "$CONCURRENCY" "http://$target/"
done

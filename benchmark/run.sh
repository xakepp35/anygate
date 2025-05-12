#!/bin/bash

# 🧠 Вариативность — наш стиль
TARGETS=(
  "echo:9000"
  "anygate:8000"
  "haproxy:80"
  "nginx-dark:80"
  "nginx:80"
)

# ⚡️ Уровень тёмного лорда DevOps'а — корото, мощно, с переменными, которые легко дергать под себя.
THREADS=4
CONNECTIONS=1024
DURATION=10s

# 🚀 Цикл ярости
for target in "${TARGETS[@]}"; do
  echo -e "\n🎯 TARGET: $target"
  docker compose exec wrk wrk -t"$THREADS" -c"$CONNECTIONS" -d"$DURATION" "http://$target/"
done


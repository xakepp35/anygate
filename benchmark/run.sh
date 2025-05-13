#!/bin/bash

# 🧠 Вариативность — наш стиль
TARGETS=(
  "echo:80"
  "anygate:80"
  "haproxy:80"
  "nginx-dark:80"
  "nginx:80"
)

# ⚡️ Уровень тёмного лорда DevOps'а — корото, мощно, с переменными, которые легко дергать под себя.
THREADS=4
CONNECTIONS=100 # Настройки бенчмарка со скринов: 4/100/10s
DURATION=10s

# 🚀 Цикл ярости
for target in "${TARGETS[@]}"; do
  echo -e "\n🎯 TARGET: $target"
  docker compose exec wrk wrk -t"$THREADS" -c"$CONNECTIONS" -d"$DURATION" "http://$target/"
done

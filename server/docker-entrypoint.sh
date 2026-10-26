#!/bin/bash
set -e

# 从配置文件中提取数据库信息
DB_HOST="postgres"
DB_PORT=5432
DB_USER="postgres"
DB_PASSWORD="wangbadan"
DB_NAME="soybean"

# 从配置文件中提取Redis信息
REDIS_HOST="redis"
REDIS_PORT=6379
REDIS_PASSWORD="wangbadan"

# 等待PostgreSQL服务可用
echo "Waiting for PostgreSQL to be available..."
until PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c '\q'; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - executing command"

# 检查数据库是否存在，如果不存在则创建
if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
  echo "Creating database $DB_NAME"
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"
fi

# 等待Redis服务可用
echo "Waiting for Redis to be available..."
until redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" ping | grep -q PONG; do
  echo "Redis is unavailable - sleeping"
  sleep 1
done

echo "Redis is up - executing command"

# 运行初始化数据的脚本
echo "Initializing data..."
/app/init-data

# 启动主应用
echo "Starting application..."
exec /app/soybean-admin-go
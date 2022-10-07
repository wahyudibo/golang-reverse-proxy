#!/usr/bin/env sh

docker run --rm --name redis -p 6379:6379 -d redis:7 redis-server --requirepass "${CACHE_REDIS_PASSWORD}"
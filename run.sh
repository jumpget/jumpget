#!/bin/bash
docker stop jumpget && docker rm jumpget || true
docker build -t jumpget .
docker run --name jumpget -p 3100:3100 \
  -p 127.0.0.1:4100:4100 -d \
  -e JUPMGET_PUBLIC_PORT=3100 -e JUMPGET_LOCAL_PORT=4100 \
  -e JUMPGET_PUBLIC_URL="example.com:3100"  lsgrep/jumpget

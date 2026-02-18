#!/bin/bash
# go test -v -run TestRunServer -timeout 30m
curl -v -X GET "http://127.0.0.1:9000/hello" -H "Cookie: user_id=12345" -H "Auth-Token: abc-xyz"
echo "(END)"
curl -X POST "http://127.0.0.1:9000/submit?source=web" -d "username=admin&password=123"
echo "(END)"
curl "http://127.0.0.1:9000/stats"
echo "(END)"
curl -s -X GET "http://127.0.0.1:9000/ping?t=$(date +%s%3N)"
echo "(END)"

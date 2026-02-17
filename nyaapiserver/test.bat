START go test -v -run TestRunServer -timeout 30m
timeout 3
curl -v -X GET "http://127.0.0.1:9000/hello" -H "Cookie: user_id=12345" -H "Auth-Token: abc-xyz"
curl -X POST "http://127.0.0.1:9000/submit?source=web" -d "username=admin&password=123"
curl "http://127.0.0.1:9000/stats"

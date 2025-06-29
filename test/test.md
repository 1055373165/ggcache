1. brew install mysql
2. 运行 mysql_config_editor set --login-path=local --host=127.0.0.1 --user=root --password
3. 会被提示输入密码，输入后密码将被安全地存储。
4. brew services start mysql
5. chmod +x create.sh
6. ./create.sh
7. 如果表创建成功将输出：数据库 students 创建成功
8. 在 main.go 中注释掉 pprof、metrics 相关代码
9. go run main.go -port 9999
10. go run main.go -port 10000 
11. go run main.go -port 10001
12. enter root directory of ggcache 
13. ./test/grpc/run_clients.sh

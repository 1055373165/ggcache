1. cd test/sql
2. chmod +x create_sql.sh
3. ./create.sh
4. 如果表创建成功将输出：数据库 ggcache 创建成功
5. brew install etcd
6. go install github.com/mattn/goreman@latest
7. cd pkg/etcd/cluster
8. goreman -f pkg/etcd/cluster/Procfile start
9. 回到主目录
10. 启动 grpc server
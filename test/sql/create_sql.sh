#!/bin/bash

# MySQL 连接信息
USER="root"
HOST="127.0.0.1"
PORT="3306"
DBNAME="cachetest"

# 使用 mysql_config_editor 存储 MySQL 用户名和密码
# 运行以下命令来配置 MySQL 用户名和密码（只需运行一次）
# mysql_config_editor set --login-path=local --host=$HOST --user=$USER --password

# 检查数据库是否存在
DB_EXISTS=$(mysql --login-path=local -h$HOST -P$PORT -e "SHOW DATABASES LIKE '$DBNAME';" | grep "$DBNAME" > /dev/null; echo "$?")

# 如果数据库不存在，则创建它
if [ $DB_EXISTS -ne 0 ]; then
  mysql --login-path=local -h$HOST -P$PORT -e "CREATE DATABASE $DBNAME;"
  echo "数据库 $DBNAME 创建成功"
else
  echo "数据库 $DBNAME 已经存在，无需重复创建"
fi
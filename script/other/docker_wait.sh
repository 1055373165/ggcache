#!/bin/bash  
  
# 定义尝试连接的次数和间隔  
MAX_RETRIES=10  
RETRY_INTERVAL=10 # 秒  
COUNTER=0  
  
# MySQL 服务的主机名和端口  
MYSQL_HOST=mysql  
MYSQL_PORT=3306  
MYSQL_USER=root  
MYSQL_PASS=1234  
  
# 检查并等待 MySQL 服务就绪  
until mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASS -e "status;" &> /dev/null; do  
    let COUNTER=COUNTER+1  
    if [ $COUNTER -ge $MAX_RETRIES ]; then  
        echo "无法连接到 MySQL 服务，超过最大重试次数"  
        exit 1  
    fi  
    echo "等待 MySQL 服务就绪..."  
    sleep $RETRY_INTERVAL  
done  
  
# MySQL 服务就绪，运行 main 命令，并传递所有参数  
exec ./main "$@"
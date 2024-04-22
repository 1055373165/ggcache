# 测试数据库创建

## 方法 1
在项目根目录下执行以下命令

```
mysql -u root -p < assets/sql/init_db.sql
```

## 方法 2

1. mysql -u root -p 进入 mysql cli
2. 输入 `source assets/sql/init_db.sql`

### 数据库驱动及连接串

| 数据库 | 连接串 |
---------| ------ |
| `mysql`| `usrname:password@tcp(server:port)/dbname?option1=value1&...`|
| `postgres`| `postgresql://username:password@server:port/dbname?option1=value1`|
|           | `user=<dbuser> password=<password> dbname=<dbname> sslmode=disable connect_timeout=3 host=<db host>` |
| `sqlite3`| `/path/to/db.sqlite?option1=value1`|
| `sqlserver` | `sqlserver://username:password@host/instance?param1=value&param2=value` |
|             | `sqlserver://username:password@host:port?param1=value&param2=value`|
|             | `sqlserver://sa@localhost/SQLExpress?database=master&connection+timeout=30`|
| `mssql` | `server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName`|
|         | `server=localhost;user id=sa;database=master;app name=MyAppName`|
|         | `odbc:server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName` |
|         | `odbc:server=localhost;user id=sa;database=master;app name=MyAppName` |
| `hdb` (SAP HANA) |   `hdb://user:password@host:port` |
| `clickhouse` (Yandex ClickHouse) |   `tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000` |

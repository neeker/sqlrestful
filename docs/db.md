# 数据库支持列表

> 使用`Go`的`https://github.com/jmoiron/sqlx`的库实现`SQL`执行，因此大部分有`database/sql`驱动实现的数据都可以纳入支持。

| 驱动名 | 数据库 | DSN |
|---------| ------ |
| `mysql` | MySQL | `usrname:password@tcp(server:port)/dbname?option1=value1&...`|
| `postgres` | PostgresQL | `postgresql://username:password@server:port/dbname?option1=value1`|
|           | | `user=<dbuser> password=<password> `<br>`dbname=<dbname> sslmode=disable connect_timeout=3 host=<db host>` |
| `sqlite3` | SQLite3 | `/path/to/db.sqlite?option1=value1`|
| `mssql`| SQLServer | `server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName`|
|        | | `server=localhost;user id=sa;database=master;app name=MyAppName`|
|        | | `odbc:server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName` |
|        | | `odbc:server=localhost;user id=sa;database=master;app name=MyAppName` |
| `hdb` | SAP HANA |   `hdb://user:password@host:port` |
| `clickhouse` | Yandex ClickHouse | `tcp://host1:9000?username=user&password=`<br>`qwerty&database=clicks&read_timeout=10&`<br>`write_timeout=20&alt_hosts=host2:9000,host3:9000` |
| `oci8` | Oracle | `username/password@host:port/sid` |

> 目前`Oracle`驱动基于`Oracle Instant Client 12.2.0.1.0`编译，因此需要`oci`库支持。

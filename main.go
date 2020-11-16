package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mylxsw/go-utils/diff"
	"github.com/mylxsw/go-utils/file"
	"github.com/mylxsw/mysql-diff/mysql"
)

var dbHost, dbUser, dbPassword string
var dbPort int
var diffVars, diffUsers, diffDatabases, withTables bool
var excludeVars, excludeDatabases string
var dataDir string
var contextLine, keepVersion uint

func main() {
	flag.StringVar(&dbHost, "db-host", "127.0.0.1", "MySQL Host")
	flag.IntVar(&dbPort, "db-port", 3306, "MySQL Port")
	flag.StringVar(&dbUser, "db-user", "root", "MySQL User")
	flag.StringVar(&dbPassword, "db-password", "", "MySQL Password")

	flag.BoolVar(&diffVars, "diff-vars", true, "是否对比系统变量差异")
	flag.BoolVar(&diffUsers, "diff-users", true, "是否对比用户差异")
	flag.BoolVar(&diffDatabases, "diff-databases", true, "是否对比数据库差异")
	flag.BoolVar(&withTables, "with-tables", false, "对比数据库差异时，是否启用表名差异对比")

	flag.StringVar(&excludeVars, "exclude-vars", "gtid_binlog_pos,gtid_binlog_state,gtid_current_pos", "需要排除的系统变量")
	flag.StringVar(&excludeDatabases, "exclude-dbs", "performance_schema,information_schema,mysql,sys", "需要排除的系统变量")

	flag.StringVar(&dataDir, "data-dir", "./tmp", "diff 状态数据存储目录")
	flag.UintVar(&contextLine, "context-line", 2, "diff 上下文信息数量")
	flag.UintVar(&keepVersion, "keep-version", 100, "保留多少个版本的历史记录")

	flag.Parse()

	fs := file.LocalFS{}
	if err := fs.MkDir(dataDir); err != nil {
		panic(err)
	}

	differ := diff.NewDiffer(fs, dataDir, int(contextLine))

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql?loc=Local&parseTime=true", dbUser, dbPassword, dbHost, dbPort))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ms := mysql.NewMySQLServer(db)

	if diffVars {
		variables, err := ms.Variables(strings.Split(excludeVars, ","))
		if err != nil {
			panic(err)
			return
		}

		df := differ.DiffLatest("variables", variables.String())
		if err := df.PrintAndSave(os.Stdout); err != nil {
			panic(err)
		}
		_ = df.Clean(keepVersion)
	}

	if diffUsers {
		users, err := ms.UsersWithPrivileges()
		if err != nil {
			panic(err)
			return
		}

		df := differ.DiffLatest("users", users.String())
		if err := df.PrintAndSave(os.Stdout); err != nil {
			panic(err)
		}
		_ = df.Clean(keepVersion)
	}

	if diffDatabases {
		databases, err := ms.Databases(strings.Split(excludeDatabases, ","), withTables)
		if err != nil {
			panic(err)
			return
		}

		df := differ.DiffLatest("databases", databases.String())
		if err := df.PrintAndSave(os.Stdout); err != nil {
			panic(err)
		}
		_ = df.Clean(keepVersion)
	}
}

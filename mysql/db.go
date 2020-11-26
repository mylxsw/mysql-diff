package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"

	"github.com/mylxsw/go-utils/str"
)

type Variables []Variable

func (vs Variables) Len() int {
	return len(vs)
}

func (vs Variables) Less(i, j int) bool {
	return vs[i].Key < vs[j].Key
}

func (vs Variables) Swap(i, j int) {
	vs[i], vs[j] = vs[j], vs[i]
}

func (vs Variables) String() string {
	strbuf := bytes.NewBuffer(nil)
	for _, v := range vs {
		strbuf.WriteString(fmt.Sprintf("VAR -> %s = %s\n", v.Key, v.Value))
	}

	return strbuf.String()
}

type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type User struct {
	User       string   `json:"user"`
	Host       string   `json:"host"`
	Privileges []string `json:"privileges"`
}

type Users []User

func (users Users) String() string {
	strbuf := bytes.NewBuffer(nil)
	for _, u := range users {
		strbuf.WriteString(fmt.Sprintf("USER -> '%s'@'%s'\n", u.User, u.Host))
		for _, p := range u.Privileges {
			strbuf.WriteString(fmt.Sprintf("PRIVILEGE -> '%s'@'%s': %s\n", u.User, u.Host, p))
		}
	}

	return strbuf.String()
}

type Database struct {
	Name   string   `json:"name"`
	Tables []string `json:"tables"`
}

type Databases []Database

func (databases Databases) String() string {
	strbuf := bytes.NewBuffer(nil)
	for _, d := range databases {
		strbuf.WriteString(fmt.Sprintf("DATABASE -> %s\n", d.Name))
		for _, t := range d.Tables {
			strbuf.WriteString(fmt.Sprintf("TABLE -> %s.%s\n", d.Name, t))
		}
	}

	return strbuf.String()
}

type MySQLServer struct {
	db *sql.DB
}

func NewMySQLServer(db *sql.DB) *MySQLServer {
	return &MySQLServer{db: db}
}

// TablesInDB 获取数据库中所有表名
func (ms *MySQLServer) TablesInDB(dbname string) ([]string, error) {
	rows, err := ms.db.Query(fmt.Sprintf("SHOW TABLES IN %s", dbname))
	if err != nil {
		return nil, err
	}

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}

		tables = append(tables, table)
	}

	sort.Strings(tables)
	return tables, nil
}

func (ms *MySQLServer) Databases(excludeDatabases []string, withTables bool) (Databases, error) {
	names, err := ms.DatabaseNames(excludeDatabases)
	if err != nil {
		return nil, err
	}

	databases := make(Databases, 0)
	for _, name := range names {
		database := Database{Name: name}

		if withTables {
			tables, err := ms.TablesInDB(name)
			if err != nil {
				return nil, err
			}
			database.Tables = tables
		}

		databases = append(databases, database)
	}

	return databases, nil
}

// DatabaseNames 获取所有数据库名称
func (ms *MySQLServer) DatabaseNames(excludeDatabases []string) ([]string, error) {
	rows, err := ms.db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}

	databases := make([]string, 0)
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			return nil, err
		}

		if str.InIgnoreCase(dbname, excludeDatabases) {
			continue
		}

		databases = append(databases, dbname)
	}

	sort.Strings(databases)
	return databases, nil
}

// UsersWithPrivileges 查询用户列表，包含用户权限
func (ms *MySQLServer) UsersWithPrivileges() (Users, error) {
	rows, err := ms.db.Query("SELECT user, host FROM mysql.user ORDER BY user, host")
	if err != nil {
		return nil, err
	}

	users := make(Users, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.User, &user.Host); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	for i, u := range users {
		privileges, err := ms.userPrivileges(u.User, u.Host)
		if err != nil {
			return nil, err
		}
		users[i].Privileges = privileges
	}

	return users, nil
}

// userPrivileges 查询用户权限列表
func (ms *MySQLServer) userPrivileges(username string, host string) ([]string, error) {
	rows, err := ms.db.Query(fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", username, host))
	if err != nil {
		return nil, err
	}

	privileges := make([]string, 0)
	for rows.Next() {
		var prv string
		if err := rows.Scan(&prv); err != nil {
			return nil, err
		}

		privileges = append(privileges, prv)
	}

	sort.Strings(privileges)
	return privileges, nil
}

// Variables 获取数据库中所有的配置信息
func (ms *MySQLServer) Variables(excludeVariables []string) (Variables, error) {
	rows, err := ms.db.Query("SHOW GLOBAL VARIABLES")
	if err != nil {
		return nil, err
	}

	results := make(Variables, 0)
	for rows.Next() {
		var variable Variable
		if err := rows.Scan(&variable.Key, &variable.Value); err != nil {
			return nil, err
		}

		if str.InIgnoreCase(variable.Key, excludeVariables) {
			continue
		}

		results = append(results, variable)
	}

	sort.Sort(results)
	return results, nil
}

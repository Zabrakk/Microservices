package mysqlconf

import (
	"os"
	"testing"
)

// Creates the environment variables used by MySQLConf
func setMySqlEnv() {
	os.Setenv("MYSQL_HOST", "HOST")
	os.Setenv("MYSQL_DB", "DB")
	os.Setenv("MYSQL_PORT", "1000")
	os.Setenv("MYSQL_USER", "USER")
	os.Setenv("MYSQL_PASSWORD", "PASSWORD")
}

func TestNewMySQLConf(t *testing.T) {
	setMySqlEnv()
	c := NewMySQLConf()
	if c.Host 		!= "HOST" { t.Fatal("MySQL host was incorrect") }
	if c.DB 		!= "DB" { t.Fatal("MySQL DB name was incorrect") }
	if c.Port 		!= "1000" { t.Fatal("MySQL port was incorrect") }
	if c.User 		!= "USER" { t.Fatal("MySQL user was incorrect") }
	if c.Password	!= "PASSWORD" { t.Fatal("MySQL password was incorrect") }
}

func TestGetDataSourceName(t *testing.T) {
	setMySqlEnv()
	c := NewMySQLConf()
	dataSourceName := c.GetDataSourceName()
	if dataSourceName != "USER:PASSWORD@tcp(HOST:1000)/DB" {
		t.Fatalf("DataSourceName was icorrect, got %s\n", dataSourceName)
	}
}

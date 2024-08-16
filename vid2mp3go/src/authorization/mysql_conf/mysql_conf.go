package mysqlconf

import "os"

type MySQLConf struct {
	Host     string
	DB       string
	Port     string
	User     string
	Password string
}

func NewMySQLConf() MySQLConf {
	return MySQLConf{
		Host:     os.Getenv("MYSQL_HOST"),
		DB:       os.Getenv("MYSQL_DB"),
		Port:     os.Getenv("MYSQL_PORT"),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
	}
}

// Creates a string that can be used as the dataSourceName for db.Open()
// based on the MySQLConf's field values
func (c MySQLConf) GetDataSourceName() (dataSourceName string) {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.DB
}
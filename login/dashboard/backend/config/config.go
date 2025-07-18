package config

import "time"

type Config struct {
	LDAPServerHost        string
	LDAPServerPort        int
	LDAPAdminDN           string
	LDAPAdminPassword     string
	LDAPSearchBaseDN      string
	LDAPUserSearchFilter  string
	SlurmAPIHost          string
	JWTSecretKey          string
	JWTIssuer             string
	JWTDuration           time.Duration
	SlurmTokenLifespanSec string
	ServerPort            string
}

// LoadConfig 加载并返回所有配置
func LoadConfig() *Config {
	return &Config{
		LDAPServerHost:       "10.20.20.20",
		LDAPServerPort:       389,
		LDAPAdminDN:          "cn=admin,dc=pushihao,dc=com",
		LDAPAdminPassword:    "root",
		LDAPSearchBaseDN:     "dc=pushihao,dc=com",
		LDAPUserSearchFilter: "(uid=%s)",

		JWTSecretKey: "2f650e45-73e4-4b36-88f7-psh-521-c3b3829d3542",
		JWTIssuer:    "slurm-dashboard-backend",
		JWTDuration:  time.Hour * 24,

		SlurmAPIHost:          "http://10.20.20.2:6820",
		SlurmTokenLifespanSec: "90000", // 25h

		ServerPort: "80",
	}
}

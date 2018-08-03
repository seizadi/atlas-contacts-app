package db

type DatabaseConfig struct {
	Type, Host, Port, Name, User, Pwd, Path, SslMode string
	CaCertPath                                       string
	ClientKeyPath                                    string
	ClientCertPath                                   string
	ServerCertName                                   string
	MaxOpenConn                                      int
	MaxIdleConn                                      int
}

package cmd

import (
	"fmt"
	"github.com/infobloxopen/atlas-contacts-app/cmd/setting"
	"github.com/infobloxopen/atlas-contacts-app/db"
)

//const (
//	// ServerAddress is the default address for the gRPC server, if no override is specified in the flags
//	ServerAddress = "0.0.0.0:9090"
//	// GatewayAddress is the default address for the gateway server, if no override is specified in the flags
//	GatewayAddress = "0.0.0.0:8080"
//	// InternalAddress is the default address for the internal http server, if no override is specified in the flags
//	InternalAddress = "0.0.0.0:8081"
//	// DatabaseAddress is the default address for the database, if no override is specified in the flags
//	DBConnectionString = "host=localhost port=5432 user=postgres password=postgres sslmode=disable dbname=atlas_contacts_app"
//	// SwaggerFile is the file location of the swagger file to serve
//	SwaggerFile = "./pkg/pb/contacts.swagger.json"
//	// ApplicationID associates a microservice with an application. The atlas
//	// contacts application consists of only one service, so we identify both the
//	// service and the application as "atlas-contacts-app"
//	ApplicationID = "atlas-contacts-app"
//)

var (
	// ServerAddress is the default address for the gRPC server
	ServerAddress string
	// GatewayAddress is the default address for the gateway server
	GatewayAddress string
	// InternalAddress is the default address for the internal http server
	InternalAddress string
	// DatabaseAddress is the default address for the database
	DbCfg db.DatabaseConfig
	DBConnectionString string
	// SwaggerFile is the file location of the swagger file to serve
	SwaggerFile string
	// ApplicationID associates a microservice with an application. The atlas
	// contacts application consists of only one service, so we identify both the
	// service and the application as "atlas-contacts-app"
	ApplicationID string
	// Log Level
	LogLevel string
	// Address of the authorization service
	AuthzAddr string
)

func LoadConfig() {
	
	LoadAppConfig()
	LoadServerConfig()
	LoadDbConfig()
	if (LogLevel == "debug") {
		setting.ShowConfigSettings()
	}
}

func LoadAppConfig() {
	ApplicationID = setting.Cfg.Section("").Key("app_id").MustString("atlas-contacts-app")
	SwaggerFile = setting.Cfg.Section("paths").Key("swagger").MustString("./pkg/pb/contacts.swagger.json")
	LogLevel = setting.Cfg.Section("log").Key("level").MustString("info")
}

func LoadServerConfig() {
	server := setting.Cfg.Section("server")
	domain := server.Key("domain").MustString("localhost")
	serverPort := server.Key("server_port").MustString("9090")
	gatewayPort := server.Key("gateway_port").MustString("8080")
	internalPort := server.Key("internal_port").MustString("8081")
	ServerAddress = fmt.Sprintf("%s:%s",domain, serverPort)
	GatewayAddress = fmt.Sprintf("%s:%s",domain, gatewayPort)
	InternalAddress = fmt.Sprintf("%s:%s",domain, internalPort)
	AuthzAddr = server.Key("authz_addr").MustString("")
}

func LoadDbConfig(){
	sec := setting.Cfg.Section("database")
	DbCfg.Host = sec.Key("host").String()
	DbCfg.Port = sec.Key("port").String()
	DbCfg.Name = sec.Key("name").String()
	DbCfg.User = sec.Key("user").String()
	DbCfg.Pwd = sec.Key("password").String()
	DbCfg.MaxOpenConn = sec.Key("max_open_conn").MustInt(0)
	DbCfg.MaxIdleConn = sec.Key("max_idle_conn").MustInt(0)
	DbCfg.SslMode = sec.Key("ssl_mode").String()
	DbCfg.CaCertPath = sec.Key("ca_cert_path").String()
	DbCfg.ClientKeyPath = sec.Key("client_key_path").String()
	DbCfg.ClientCertPath = sec.Key("client_cert_path").String()
	DbCfg.ServerCertName = sec.Key("server_cert_name").String()
	SetDbConnectionString()
}

func SetDbConnectionString() {
	if DbCfg.Pwd == "" {
		DbCfg.Pwd = "''"
	}
	if DbCfg.User == "" {
		DbCfg.User = "''"
	}
	DBConnectionString = fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
		DbCfg.User, DbCfg.Pwd, DbCfg.Host, DbCfg.Port, DbCfg.Name, DbCfg.SslMode, DbCfg.ClientCertPath, DbCfg.ClientKeyPath, DbCfg.CaCertPath)
}
package koanfexample

var defaultConf = map[string]interface{}{
	// User server
	"servers.userserver.port":         8080,
	"servers.userserver.readTimeout":  "5s",
	"servers.userserver.writeTimeout": "3m",

	// Product server
	"servers.productserver.port":         8090,
	"servers.productserver.readTimeout":  "1s",
	"servers.productserver.writeTimeout": "1m",

	// DB
	"db.dataSourceName":   "root:password@tcp(127.0.0.1:3306)/local_db?charset=utf8&parseTime=True&multiStatements=true",
	"db.pool.maxIdle":     20,
	"db.pool.maxLifetime": "86400s",
	"db.pool.maxOpen":     10,
}

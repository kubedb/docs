package docker

const (
	ImageOperator          = "aerokite/operator"
	ImagePostgresOperator  = "kubedb/pg-operator"
	ImagePostgres          = "aerokite/postgres"
	ImageMySQLOperator     = "kubedb/mysql-operator"
	ImageMySQL             = "library/mysql"
	ImageElasticOperator   = "kubedb/es-operator"
	ImageElasticsearch     = "aerokite/elasticsearch"
	ImageElasticdump       = "aerokite/elasticdump"
	ImageMongoDBOperator   = "kubedb/mongodb-operator"
	ImageMongoDB           = "library/mongo"
	ImageRedisOperator     = "kubedb/redis-operator"
	ImageRedis             = "library/redis"
	ImageMemcachedOperator = "kubedb/mc-operator"
	ImageMemcached         = "library/memcached"
)

const (
	OperatorName       = "kubedb-operator"
	OperatorContainer  = "operator"
	OperatorPortName   = "web"
	OperatorPortNumber = 8080
)

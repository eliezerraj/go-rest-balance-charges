package main

import(
	"time"
	"os"
	"strings"
	"strconv"
	"net"
	"io/ioutil"
	"context"
	"crypto/tls"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
    "github.com/aws/aws-sdk-go-v2/config"

	"github.com/go-rest-balance-charges/internal/circuitbreaker"
	"github.com/go-rest-balance-charges/internal/handler"
	"github.com/go-rest-balance-charges/internal/core"
	"github.com/go-rest-balance-charges/internal/service"
	"github.com/go-rest-balance-charges/internal/repository/postgre"
	"github.com/go-rest-balance-charges/internal/repository/cache"
	"github.com/go-rest-balance-charges/internal/adapter/restapi"
	redis "github.com/redis/go-redis/v9"
	
)

var(
	logLevel 	= zerolog.DebugLevel
	version 	= "GO CRUD BALANCE_CHARGE 1.0"
	serverUrlDomain	string
	path			string
	noAZ		=	true // set only if you get to split the xray trace per AZ

	infoPod					core.InfoPod
	envDB	 				core.DatabaseRDS
	httpAppServerConfig 	core.HttpAppServer
	server					core.Server
	dataBaseHelper 			db_postgre.DatabaseHelper
	repoDB					db_postgre.WorkerRepository
	restApiBalance			restapi.RestApiSConfig
	cache					cache_redis.CacheService
	envCacheCluster			redis.ClusterOptions
	envCache				redis.Options
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	// Just for easy test
	envDB.Host = "127.0.0.1" //"host.docker.internal"
	envDB.Port = "5432"
	envDB.Schema = "public"
	envDB.DatabaseName = "postgres"
	//envDB.User  = "postgres"
	//envDB.Password  = "pass123"
	serverUrlDomain 	= "http://localhost:5000"
	path				= "/get"
	envDB.Db_timeout = 90
	envDB.Postgres_Driver = "postgres"
	server.Port = 5001

	envCacheCluster.Username = ""
	envCacheCluster.Password = ""
	envCacheCluster.Addrs = strings.Split("clustercfg.memdb-arch.vovqz2.memorydb.us-east-2.amazonaws.com:6379", ",")

	envCache.Username = ""
	envCache.Password = ""
	envCache.Addr = "127.0.0.1:6379"
	envCache.DB	= 0
	//Just for easy test

	server.ReadTimeout = 60
	server.WriteTimeout = 60
	server.IdleTimeout = 60
	server.CtxTimeout = 60

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error().Err(err).Msg("Error to get the POD IP address !!!")
		os.Exit(3)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				infoPod.IPAddress = ipnet.IP.String()
			}
		}
	}
	infoPod.OSPID = strconv.Itoa(os.Getpid())

	file_user, err := ioutil.ReadFile("/var/pod/secret/username")
    if err != nil {
        log.Error().Err(err).Msg("ERRO FATAL recuperacao secret-user")
		os.Exit(3)
    }
	file_pass, err := ioutil.ReadFile("/var/pod/secret/password")
    if err != nil {
        log.Error().Err(err).Msg("ERRO FATAL recuperacao secret-pass")
		os.Exit(3)
    }
	envDB.User = string(file_user)
	envDB.Password = string(file_pass)

	getEnv()

	// Get AZ only if localtest is true
	if (noAZ != true) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Error().Err(err).Msg("ERRO FATAL get Context !!!")
			os.Exit(3)
		}
		client := imds.NewFromConfig(cfg)
		response, err := client.GetInstanceIdentityDocument(context.TODO(), &imds.GetInstanceIdentityDocumentInput{})
		if err != nil {
			log.Error().Err(err).Msg("Unable to retrieve the region from the EC2 instance !!!")
			os.Exit(3)
		}
		infoPod.AvailabilityZone = response.AvailabilityZone	
	} else {
		infoPod.AvailabilityZone = "LOCALHOST_NO_AZ"
	}
}

func getEnv() {
	log.Debug().Msg("getEnv")

	if os.Getenv("API_VERSION") !=  "" {
		infoPod.ApiVersion = os.Getenv("API_VERSION")
	}
	if os.Getenv("POD_NAME") !=  "" {
		infoPod.PodName = os.Getenv("POD_NAME")
	}

	if os.Getenv("PORT") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("PORT"))
		server.Port = intVar
	}

	if os.Getenv("DB_HOST") !=  "" {
		envDB.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") !=  "" {
		envDB.Port = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_NAME") !=  "" {	
		envDB.DatabaseName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_SCHEMA") !=  "" {	
		envDB.Schema = os.Getenv("DB_SCHEMA")
	}

	if os.Getenv("SERVER_URL_DOMAIN") !=  "" {	
		serverUrlDomain = os.Getenv("SERVER_URL_DOMAIN")
	}
	if os.Getenv("SERVER_PATH") !=  "" {	
		path = os.Getenv("SERVER_PATH")
	}

	if os.Getenv("NO_AZ") == "false" {	
		noAZ = false
	} else {
		noAZ = true
	}

	if os.Getenv("REDIS_ADDRESS") !=  "" {	
		envCache.Addr =  os.Getenv("REDIS_ADDRESS")
	}
	if os.Getenv("REDIS_CLUSTER_ADDRESS") !=  "" {	
		envCacheCluster.Addrs =  strings.Split(os.Getenv("REDIS_CLUSTER_ADDRESS"), ",") 
	}
	if os.Getenv("REDIS_DB_NAME") !=  "" {	
		intVar, _ := strconv.Atoi(os.Getenv("REDIS_DB_NAME"))
		envCache.DB = intVar
	}
	if os.Getenv("REDIS_PASSWORD") !=  "" {	
		envCache.Password = os.Getenv("REDIS_PASSWORD")
	}

}

func main() {
	log.Debug().Msg("main")
	log.Debug().Interface("",envDB).Msg("getEnv")
	log.Debug().Msg("--------")
	log.Debug().Interface("",server).Msg("server")
	log.Debug().Msg("--------")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration( server.ReadTimeout ) * time.Second)
	defer cancel()

	count := 1
	var err error
	for {
		dataBaseHelper, err = db_postgre.NewDatabaseHelper(ctx, envDB)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro na abertura do Database")
			} else {
				log.Error().Err(err).Msg("ERRO FATAL na abertura do Database aborting")
				panic(err)	
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		break
	}

	if !strings.Contains(envCache.Addr, "127.0.0.1") {
		envCache.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	if !strings.Contains(envCacheCluster.Addrs[0], "127.0.0.1") {
		log.Debug().Msg("tls ok")
		envCacheCluster.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	cache := cache_redis.NewClusterCache(ctx, &envCacheCluster)
	//cache := cache_redis.NewCache(ctx, &envCache)
	_, err = cache.Ping(ctx)
	if err != nil{
		log.Error().Err(err).Msg("Erro na abertura do Redis")
	}

	circuitBreaker := circuitbreaker.CircuitBreakerConfig()
	restapi	:= restapi.NewRestApi(serverUrlDomain, path)
	httpAppServerConfig.Server = server
	repoDB = db_postgre.NewWorkerRepository(dataBaseHelper)
	workerService := service.NewWorkerService(&repoDB, restapi, circuitBreaker, cache)
	httpWorkerAdapter := handler.NewHttpWorkerAdapter(workerService)

	httpAppServerConfig.InfoPod = &infoPod
	httpServer := handler.NewHttpAppServer(httpAppServerConfig)

	httpServer.StartHttpAppServer(ctx, httpWorkerAdapter)
}
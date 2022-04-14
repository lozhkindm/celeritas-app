package celeritas

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/lozhkindm/celeritas/cache"
	"github.com/lozhkindm/celeritas/render"
	"github.com/lozhkindm/celeritas/session"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

type Celeritas struct {
	AppName       string
	Debug         bool
	Version       string
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            database
	Cache         cache.Cache
	JetViews      *jet.Set
	EncryptionKey string
	config        config
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	db          dbConfig
	redis       redisConfig
}

func (c *Celeritas) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middlewares"},
	}

	if err := c.init(pathConfig); err != nil {
		return err
	}
	if err := c.checkDotEnv(rootPath); err != nil {
		return err
	}
	if err := godotenv.Load(fmt.Sprintf("%s/.env", rootPath)); err != nil {
		return err
	}

	c.createLoggers()
	c.createDB()
	c.createCache()
	c.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	c.Version = version
	c.RootPath = rootPath
	c.createConfig()
	c.Routes = c.routes().(*chi.Mux)
	c.prepareJetViews(rootPath)
	c.createSession()
	c.createRenderer()
	c.EncryptionKey = os.Getenv("KEY")

	return nil
}

func (c *Celeritas) prepareJetViews(rootPath string) {
	if c.Debug {
		c.JetViews = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)
	} else {
		c.JetViews = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)
	}
}

func (c *Celeritas) ListenAndServe() {
	defer func() {
		if err := c.DB.Pool.Close(); err != nil {
			c.ErrorLog.Fatal(err)
		}
	}()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      c.Routes,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
		IdleTimeout:  30 * time.Second,
		ErrorLog:     c.ErrorLog,
	}
	c.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	if err := srv.ListenAndServe(); err != nil {
		c.ErrorLog.Fatal(err)
	}
}

func (c *Celeritas) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"),
		)
		if pass := os.Getenv("DATABASE_PASS"); pass != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, pass)
		}
	default:

	}

	return dsn
}

func (c *Celeritas) init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		if err := c.CreateDirIfNotExists(fmt.Sprintf("%s/%s", root, path)); err != nil {
			return err
		}
	}
	return nil
}

func (c *Celeritas) checkDotEnv(path string) error {
	if err := c.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path)); err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) createLoggers() {
	c.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	c.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
}

func (c *Celeritas) createConfig() {
	c.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		db: dbConfig{
			dsn:      c.BuildDSN(),
			database: os.Getenv("DATABASE_TYPE"),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}
}

func (c *Celeritas) createRenderer() {
	c.Render = &render.Render{
		Renderer: c.config.renderer,
		RootPath: c.RootPath,
		Port:     c.config.port,
		JetViews: c.JetViews,
		Session:  c.Session,
	}
}

func (c *Celeritas) createSession() {
	s := session.Session{
		CookieLifetime: c.config.cookie.lifetime,
		CookiePersist:  c.config.cookie.persist,
		CookieName:     c.config.cookie.name,
		CookieDomain:   c.config.cookie.domain,
		CookieSecure:   c.config.cookie.secure,
		SessionType:    c.config.sessionType,
	}
	switch s.SessionType {
	case "redis":
		s.RedisPool = c.Cache.(*cache.RedisCache).Conn
	case "mysql", "mariadb", "postgres", "postgresql":
		s.DBPool = c.DB.Pool
	}
	c.Session = s.Init()
}

func (c *Celeritas) createDB() {
	if dbType := os.Getenv("DATABASE_TYPE"); dbType != "" {
		db, err := c.OpenDB(dbType, c.BuildDSN())
		if err != nil {
			c.ErrorLog.Fatalln(err)
		}
		c.DB = database{DataType: dbType, Pool: db}
	}
}

func (c *Celeritas) createCache() {
	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		c.Cache = c.createRedisCacheClient()
	}
}

func (c *Celeritas) createRedisCacheClient() *cache.RedisCache {
	return &cache.RedisCache{
		Conn:   c.createRedisPool(),
		Prefix: c.config.redis.prefix,
	}
}

func (c *Celeritas) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				c.config.redis.host,
				redis.DialPassword(c.config.redis.password),
			)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

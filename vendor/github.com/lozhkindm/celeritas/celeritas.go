package celeritas

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lozhkindm/celeritas/cache"
	"github.com/lozhkindm/celeritas/filesystem/minio"
	"github.com/lozhkindm/celeritas/filesystem/s3"
	"github.com/lozhkindm/celeritas/filesystem/sftp"
	"github.com/lozhkindm/celeritas/filesystem/webdav"
	"github.com/lozhkindm/celeritas/mailer"
	"github.com/lozhkindm/celeritas/render"
	"github.com/lozhkindm/celeritas/session"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
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
	DB            Database
	Cache         cache.Cache
	JetViews      *jet.Set
	EncryptionKey string
	Scheduler     *cron.Cron
	Mail          mailer.Mail
	Server        Server
	FileSystems   map[string]interface{}
	S3            s3.S3
	SFTP          sftp.SFTP
	WebDAV        webdav.WebDAV
	Minio         minio.Minio
	config        config
	redisPool     *redis.Pool
	badgerConn    *badger.DB
}

type Server struct {
	Name   string
	Port   string
	Secure bool
	URL    string
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	db          dbConfig
	redis       redisConfig
	upload      uploadConfig
}

func (c *Celeritas) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath: rootPath,
		folderNames: []string{
			"handlers", "migrations", "views", "mails", "data", "public", "tmp", "logs", "middlewares", "screenshots",
		},
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
	c.Scheduler = cron.New()
	if err := c.createCache(); err != nil {
		return err
	}
	c.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	c.Version = version
	c.RootPath = rootPath
	if err := c.createConfig(); err != nil {
		return err
	}
	c.Routes = c.routes().(*chi.Mux)
	c.prepareJetViews(rootPath)
	c.createSession()
	c.createRenderer()
	c.EncryptionKey = os.Getenv("KEY")
	c.createMailer()
	c.createServer()
	c.createFileSystem()

	go c.Mail.ListenForMail()

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

func (c *Celeritas) createConfig() error {
	uploadSize, err := strconv.Atoi(os.Getenv("MAX_UPLOAD_SIZE"))
	if err != nil {
		return err
	}
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
		upload: uploadConfig{
			allowedMimes: strings.Split(os.Getenv("ALLOWED_MIMETYPES"), ","),
			maxSize:      int64(uploadSize),
		},
	}
	return nil
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
		s.RedisPool = c.redisPool
	case "mysql", "mariadb", "postgres", "postgresql":
		s.DBPool = c.DB.Pool
	}
	c.Session = s.Init()
}

func (c *Celeritas) createDB() {
	if dbType := os.Getenv("DATABASE_TYPE"); dbType != "" {
		db, err := c.OpenDB(dbType, c.BuildDSN())
		if err != nil {
			c.ErrorLog.Fatal(err)
		}
		c.DB = Database{DataType: dbType, Pool: db}
	}
}

func (c *Celeritas) createCache() error {
	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		c.Cache = c.createRedisCacheClient()
	}
	if os.Getenv("CACHE") == "badger" {
		client, err := c.createBadgerCacheClient()
		if err != nil {
			return err
		}

		_, err = c.Scheduler.AddFunc("@daily", func() {
			if err := client.Conn.RunValueLogGC(0.7); err != nil {
				c.ErrorLog.Println(err)
			}
		})
		if err != nil {
			return err
		}

		c.Cache = client
	}
	return nil
}

func (c *Celeritas) createRedisCacheClient() *cache.RedisCache {
	c.redisPool = c.createRedisPool()
	return &cache.RedisCache{
		Conn:   c.redisPool,
		Prefix: c.config.redis.prefix,
	}
}

func (c *Celeritas) createBadgerCacheClient() (*cache.BadgerCache, error) {
	conn, err := c.createBadgerConn()
	if err != nil {
		return nil, err
	}
	c.badgerConn = conn
	return &cache.BadgerCache{Conn: c.badgerConn}, nil
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

func (c *Celeritas) createBadgerConn() (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions(fmt.Sprintf("%s/tmp/badger", c.RootPath)))
}

func (c *Celeritas) createMailer() {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		c.ErrorLog.Fatal(err)
	}

	c.Mail = mailer.Mail{
		Domain:       os.Getenv("MAIL_DOMAIN"),
		TemplatesDir: fmt.Sprintf("%s/mails", c.RootPath),
		Host:         os.Getenv("SMTP_HOST"),
		Port:         port,
		Username:     os.Getenv("SMTP_USERNAME"),
		Password:     os.Getenv("SMTP_PASSWORD"),
		Encryption:   os.Getenv("SMTP_ENCRYPTION"),
		FromAddress:  os.Getenv("FROM_NAME"),
		FromName:     os.Getenv("FROM_ADDRESS"),
		Jobs:         make(chan mailer.Message, 20),
		Results:      make(chan mailer.Result, 20),
		API:          os.Getenv("MAILER_API"),
		APIKey:       os.Getenv("MAILER_KEY"),
		APIUrl:       os.Getenv("MAILER_URL"),
	}
}

func (c *Celeritas) createServer() {
	secure, _ := strconv.ParseBool(os.Getenv("SECURE"))
	c.Server = Server{
		Name:   os.Getenv("SERVER_NAME"),
		Port:   os.Getenv("PORT"),
		Secure: secure,
		URL:    os.Getenv("APP_URL"),
	}
}

func (c *Celeritas) createFileSystem() {
	c.FileSystems = make(map[string]interface{})
	if os.Getenv("MINIO_SECRET") != "" {
		usessl, _ := strconv.ParseBool(os.Getenv("MINIO_USESSL"))
		c.Minio = minio.Minio{
			Endpoint: os.Getenv("MINIO_ENDPOINT"),
			Key:      os.Getenv("MINIO_KEY"),
			Secret:   os.Getenv("MINIO_SECRET"),
			UseSSL:   usessl,
			Region:   os.Getenv("MINIO_REGION"),
			Bucket:   os.Getenv("MINIO_BUCKET"),
		}
		c.FileSystems["MINIO"] = c.Minio
	}
	if os.Getenv("SFTP_HOST") != "" {
		c.SFTP = sftp.SFTP{
			Host:     os.Getenv("SFTP_HOST"),
			User:     os.Getenv("SFTP_USER"),
			Password: os.Getenv("SFTP_PASSWORD"),
			Port:     os.Getenv("SFTP_PORT"),
		}
		c.FileSystems["SFTP"] = c.SFTP
	}
	if os.Getenv("WEBDAV_HOST") != "" {
		c.WebDAV = webdav.WebDAV{
			Host:     os.Getenv("WEBDAV_HOST"),
			User:     os.Getenv("WEBDAV_USER"),
			Password: os.Getenv("WEBDAV_PASSWORD"),
		}
		c.FileSystems["WEBDAV"] = c.WebDAV
	}
	if os.Getenv("S3_KEY") != "" {
		c.S3 = s3.S3{
			Key:      os.Getenv("S3_KEY"),
			Secret:   os.Getenv("S3_SECRET"),
			Region:   os.Getenv("S3_REGION"),
			Endpoint: os.Getenv("S3_ENDPOINT"),
			Bucket:   os.Getenv("S3_BUCKET"),
		}
		c.FileSystems["S3"] = c.S3
	}
}

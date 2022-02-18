package config

import (
	_ "embed"

	"flag"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/ini.v1"
)

type Config struct {
	*ini.File
	sync.RWMutex

	Mode        string
	Initialized bool

	Meta
	Database
	Redis
	Security
	Server
	Service
	Cache
	Directories
}

type Meta struct {
	BaseURL     string `json:"baseURL"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
}

type Database struct {
	Host    string
	Port    int
	Name    string
	User    string
	Passwd  string
	SSLMode string
}

type Redis struct {
	Host   string
	Port   int
	DB     int
	Passwd string
}

type Security struct {
	JWTSessionSecret []byte
	JWTRefreshSecret []byte
}

type Server struct {
	Port int
}

type Service struct {
	DisableRegistration bool `json:"disableRegistration"`
	CoverMaxFileSize    int  `json:"coverMaxFileSize"`
	PageMaxFileSize     int  `json:"pageMaxFileSize"`
}

type Cache struct {
	DefaultTTL   time.Duration
	TemplatesTTL time.Duration
}

type Directories struct {
	Root string
}

//go:embed config.ini
var buf []byte

var config *Config
var path string

var (
	User    *user.User
	UserID  int
	GroupID int
)

func init() {
	p := flag.String("config", "", "Path to config file")
	m := flag.String("mode", "production", "App mode")

	flag.Parse()

	path = *p
	if len(path) == 0 {
		ex, err := os.Executable()
		if err != nil {
			log.Fatalln(err)
		}
		path = filepath.Join(filepath.Dir(ex), "config.ini")
	}

	var err error
	User, err = user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	UserID, _ = strconv.Atoi(User.Uid)
	GroupID, _ = strconv.Atoi(User.Gid)

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalln(err)
			} else if err := os.Chown(dir, UserID, GroupID); err != nil {
				log.Fatalln(err)
			}
		}
		if err := os.WriteFile(path, buf, 0755); err != nil {
			log.Fatalln(err)
		} else if err := os.Chown(path, UserID, GroupID); err != nil {
			log.Fatalln(err)
		}
	}

	var file *ini.File
	if file, err = ini.Load(path); err != nil {
		log.Fatalln(err)
	}

	config = &Config{
		File: file,

		Mode:        file.Section("").Key("mode").MustString("production"),
		Initialized: file.Section("").Key("initialized").MustBool(false),

		Meta: Meta{
			BaseURL:     file.Section("meta").Key("base_url").MustString("http://localhost:42072"),
			Title:       file.Section("meta").Key("title").MustString("Kasen"),
			Description: file.Section("meta").Key("description").MustString("CMS for scanlators"),
			Language:    file.Section("meta").Key("language").MustString("en-US"),
		},

		Database: Database{
			Host:    file.Section("database").Key("host").MustString("localhost"),
			Port:    file.Section("database").Key("port").MustInt(5432),
			Name:    file.Section("database").Key("name").MustString("kasen"),
			User:    file.Section("database").Key("user").MustString("kasen"),
			Passwd:  file.Section("database").Key("passwd").MustString("kasen"),
			SSLMode: file.Section("database").Key("ssl_mode").MustString("disable"),
		},

		Redis: Redis{
			Host:   file.Section("redis").Key("host").MustString("localhost"),
			Port:   file.Section("redis").Key("port").MustInt(6379),
			DB:     file.Section("redis").Key("db").MustInt(0),
			Passwd: file.Section("redis").Key("passwd").String(),
		},

		Security: Security{
			JWTSessionSecret: []byte(file.Section("security").Key("jwt_session_secret").String()),
			JWTRefreshSecret: []byte(file.Section("security").Key("jwt_refresh_secret").String()),
		},

		Server: Server{
			Port: file.Section("server").Key("port").MustInt(42072),
		},

		Service: Service{
			DisableRegistration: file.Section("service").Key("disable_registration").MustBool(true),
			CoverMaxFileSize:    file.Section("service").Key("cover_max_file_size").MustInt(10485760),
			PageMaxFileSize:     file.Section("service").Key("page_max_file_size").MustInt(20971520),
		},

		Cache: Cache{
			DefaultTTL:   time.Duration(file.Section("cache").Key("default_ttl").MustInt(86400000000000)),
			TemplatesTTL: time.Duration(file.Section("cache").Key("templates_ttl").MustInt(300000000000)),
		},

		Directories: Directories{
			Root: file.Section("directories").Key("root").MustString("/var/lib/kasen"),
		},
	}

	if len(*m) > 0 {
		config.Mode = *m
	}

	if len(config.Security.JWTSessionSecret) == 0 {
		config.Security.JWTSessionSecret = []byte(uuid.New().String())
	}

	if len(config.Security.JWTRefreshSecret) == 0 {
		config.Security.JWTRefreshSecret = []byte(uuid.New().String())
	}

	Save()
}

func GetMode() string {
	config.RLock()
	defer config.RUnlock()
	return config.Mode
}

func SetMode(v string) {
	config.Lock()
	defer config.Unlock()
	config.Mode = v
}

func GetInitialized() bool {
	config.RLock()
	defer config.RUnlock()
	return config.Initialized
}

func SetInitialized(v bool) {
	config.Lock()
	defer config.Unlock()
	config.Initialized = v
}

func GetMeta() Meta {
	config.RLock()
	defer config.RUnlock()
	return config.Meta
}

func SetMeta(v Meta) {
	config.Lock()
	defer config.Unlock()
	config.Meta = v
}

func GetDatabase() Database {
	config.RLock()
	defer config.RUnlock()
	return config.Database
}

func SetDatabase(v Database) {
	config.Lock()
	defer config.Unlock()
	config.Database = v
}

func GetRedis() Redis {
	config.RLock()
	defer config.RUnlock()
	return config.Redis
}

func SetRedis(v Redis) {
	config.Lock()
	defer config.Unlock()
	config.Redis = v
}

func GetSecurity() Security {
	config.RLock()
	defer config.RUnlock()
	return config.Security
}

func SetSecurity(v Security) {
	config.Lock()
	defer config.Unlock()
	config.Security = v
}

func GetServer() Server {
	config.RLock()
	defer config.RUnlock()
	return config.Server
}

func SetServer(v Server) {
	config.Lock()
	defer config.Unlock()
	config.Server = v
}

func GetService() Service {
	config.RLock()
	defer config.RUnlock()
	return config.Service
}

func SetService(v Service) {
	config.Lock()
	defer config.Unlock()
	config.Service = v
}

func GetCache() Cache {
	config.RLock()
	defer config.RUnlock()
	return config.Cache
}

func SetCache(v Cache) {
	config.Lock()
	defer config.Unlock()
	config.Cache = v
}

func GetDirectories() Directories {
	config.RLock()
	defer config.RUnlock()
	return config.Directories
}

func SetDirectories(v Directories) {
	config.Lock()
	defer config.Unlock()
	config.Directories = v
}

func Save() error {
	config.Lock()
	defer config.Unlock()

	config.Section("").Key("mode").SetValue(config.Mode)
	config.Section("").Key("initialized").SetValue(strconv.FormatBool(config.Initialized))

	config.Section("meta").Key("base_url").SetValue(config.Meta.BaseURL)
	config.Section("meta").Key("description").SetValue(config.Meta.Description)
	config.Section("meta").Key("title").SetValue(config.Meta.Title)
	config.Section("meta").Key("language").SetValue(config.Meta.Language)

	config.Section("database").Key("host").SetValue(config.Database.Host)
	config.Section("database").Key("port").SetValue(strconv.Itoa(config.Database.Port))
	config.Section("database").Key("name").SetValue(config.Database.Name)
	config.Section("database").Key("user").SetValue(config.Database.User)
	config.Section("database").Key("passwd").SetValue(config.Database.Passwd)
	config.Section("database").Key("ssl_mode").SetValue(config.Database.SSLMode)

	config.Section("redis").Key("host").SetValue(config.Redis.Host)
	config.Section("redis").Key("port").SetValue(strconv.Itoa(config.Redis.Port))
	config.Section("redis").Key("db").SetValue(strconv.Itoa(config.Redis.DB))
	config.Section("redis").Key("passwd").SetValue(config.Redis.Passwd)

	config.Section("security").Key("jwt_session_secret").SetValue(string(config.Security.JWTSessionSecret))
	config.Section("security").Key("jwt_refresh_secret").SetValue(string(config.Security.JWTRefreshSecret))

	config.Section("server").Key("port").SetValue(strconv.Itoa(config.Server.Port))

	config.Section("service").Key("disable_registration").SetValue(strconv.FormatBool(config.Service.DisableRegistration))
	config.Section("service").Key("cover_max_file_size").SetValue(strconv.Itoa(config.Service.CoverMaxFileSize))
	config.Section("service").Key("page_max_file_size").SetValue(strconv.Itoa(config.Service.PageMaxFileSize))

	config.Section("cache").Key("default_ttl").SetValue(strconv.Itoa(int(config.Cache.DefaultTTL)))
	config.Section("cache").Key("templates_ttl").SetValue(strconv.Itoa(int(config.Cache.TemplatesTTL)))

	config.Section("directories").Key("root").SetValue(config.Directories.Root)

	return config.SaveTo(path)
}

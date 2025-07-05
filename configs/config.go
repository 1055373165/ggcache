package configs

import (
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var Conf *Config
var DefaultEtcdConfig clientv3.Config

type Config struct {
	Mysql        *MySQL              `yaml:"mysql"`
	Etcd         *Etcd               `yaml:"etcd"`
	Services     map[string]*Service `yaml:"services"`
	Domain       map[string]*Domain  `yaml:"domain"`
	Server       *Server            `yaml:"server"`
	GroupManager *GroupManager       `yaml:"groupManager"`
}

type MySQL struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
}

type Etcd struct {
	Address []string `yaml:"address"`
	TTL     int      `yaml:"ttl"`
}

type Service struct {
	Name        string   `yaml:"name"`
	LoadBalance bool     `yaml:"loadBalance"`
	Addr        []string `yaml:"addr"`
	TTL         int      `yaml:"ttl"`
}

type Domain struct {
	Name string `yaml:"name"`
}

type Server struct {
	Port        int `yaml:"port"`
	MetricsPort int `yaml:"metricsPort"`
	PprofPort   int `yaml:"pprofPort"`
}

type GroupManager struct {
	Strategy     string `yaml:"strategy"`
	MaxCacheSize int64  `yaml:"maxCacheSize"`
}

// InitConfig loads configuration. If customPath is provided it will be used, otherwise
// we search for `configs/config.yml` under project root.
func InitConfig(customPath ...string) error {
	if len(customPath) > 0 && customPath[0] != "" {
		viper.SetConfigFile(customPath[0])
	} else {
		rootDir := findRootDir()
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
		viper.AddConfigPath(path.Join(rootDir, "configs"))
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	// parse into Conf object
	if err := viper.Unmarshal(&Conf); err != nil {
		return err
	}
	InitClientV3Config()
	return nil
}

func InitClientV3Config() {
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   Conf.Etcd.Address,
		DialTimeout: 5 * time.Second,
	}
}

func findRootDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Traverse upward until you find the go.mod file
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			panic("reached top of file system without finding go.mod")
		}
		currentDir = parentDir
	}
}

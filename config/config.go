package config

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
	Mysql    *Mysql              `yaml:"server"`
	Etcd     *Etcd               `yaml:"etcd"`
	Services map[string]*Service `yaml:"services"`
	Domain   map[string]*Domain  `yaml:"domain"`
	Server   Server
}

type Mysql struct {
	DriverName string `yaml:"port"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Database   string `yaml:"database"`
	UserName   string `yaml:"username"`
	Password   string `yaml:"password"`
	Charset    string `yaml:"charset"`
	ParseTime  string `yaml:"parsetime"`
	Local      string `yaml:"loc"`
}

type Etcd struct {
	AddressList []string `yaml:"addresslist"`
	TimeToLive  int      `yaml:"ttl"`
}

type Service struct {
	Name        string   `yaml:"name"`
	LoadBalance bool     `yaml:"loadBalance"`
	Addr        []string `yaml:"addr"`
}

type Domain struct {
	Name string `yaml:"name"`
}

type Server struct {
	Port      string `yaml:"port"`
	Version   string `yaml:"version"`
	JwtSecret string `yaml:"jwtSecret"`
}

func InitConfig() {
	rootDir := findRootDir()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(path.Join(rootDir, "config"))
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// parse into Conf object
	err = viper.Unmarshal(&Conf)
	if err != nil {
		panic(err)
	}
	InitClientV3Config()
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

func InitClientV3Config() {
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   Conf.Etcd.AddressList,
		DialTimeout: 5 * time.Second,
	}
}

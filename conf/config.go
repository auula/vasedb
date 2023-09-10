package conf

import (
	"encoding/json"
	"io/fs"

	"github.com/auula/vasedb/clog"
	"github.com/spf13/viper"
)

const (
	cfSuffix = "yaml"

	defaultFilePath = "./config.yaml"

	// DefaultConfigJSON configure json string
	DefaultConfigJSON = `
	{
    "vasedb": {
        "port": 2468,
        "mode": "mmap",
        "url": "/cql",
        "filesize": 1024,
        "path": "/tmp/vasedb",
        "auth": "password@123",
        "log": "/tmp/vasedb/out.log",
        "debug": false,
        "encoder": {
            "enable": true,
            "secret": "/tmp/vasedb/etc/encrypt.wasm"
        },
        "compressor": {
            "enable": true,
            "mode": "cycle",
            "second": 15000
        }
    }
}
`
)

// Settings global configure options
var Settings *ServerConfig = new(ServerConfig)

// DefaultConfig is the default configuration
var DefaultConfig *ServerConfig = new(ServerConfig)

// Dirs 标准目录结构
var Dirs = []string{"etc", "temp", "index", "data"}

func init() {

	// 先读内置默认配置，设置为全局的配置
	if err := DefaultConfig.Unmarshal([]byte(DefaultConfigJSON)); err != nil {
		// 读取失败直接退出进程
		clog.Failed(err)
	}

	// 设置默认的配置文件路径
	DefaultConfig.FilePath = defaultFilePath

	// 设置默认文件系统权限
	DefaultConfig.Permissions = fs.FileMode(0600)

	// 当初始化完成之后应该使用此 Settings 配置
	if err := Settings.Unmarshal([]byte(DefaultConfigJSON)); err != nil {
		// 读取失败直接退出进程
		clog.Failed(err)
	}

}

// Load through a configuration file
func Load(file string, opt *ServerConfig) error {

	v := viper.New()
	v.SetConfigType(cfSuffix)
	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(&opt)
}

func (opt *ServerConfig) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &opt)
}

func (opt *ServerConfig) Marshal() ([]byte, error) {
	return json.Marshal(opt)
}

type ServerConfig struct {
	VaseDB      `json:"vasedb"`
	FilePath    string
	Permissions fs.FileMode
}

type VaseDB struct {
	Port       int32      `json:"port"`
	Path       string     `json:"path"`
	Mode       string     `json:"mode"`
	Debug      bool       `json:"debug"`
	FileSize   int64      `json:"filesize"`
	Logging    string     `json:"log"`
	Password   string     `json:"auth"`
	Encoder    Encoder    `json:"encoder"`
	Compressor Compressor `json:"compressor"`
}

type Compressor struct {
	Enable  bool   `json:"enable"`
	Mode    string `json:"mode"`
	Hotspot int64  `json:"hotspot"`
	Second  int64  `json:"second"`
}

type Encoder struct {
	Enable bool   `json:"enable"`
	Secret string `json:"secret"`
}

package conf

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	extension       = "yaml"
	fileName        = "config"
	defaultFilePath = ""

	// 设置默认文件系统权限
	FsPerm = fs.FileMode(0755)

	// DefaultConfigJSON configure json string
	DefaultConfigJSON = `
{
	"port": 2468,
	"path": "/tmp/vasedb",
	"auth": "",
	"log_path": "/tmp/vasedb/out.log",
	"debug": false,
	"compressor": {
		"enable": true,
		"second": 15000
	}
}
`
)

var (
	// Settings global configure options
	Settings *ServerConfig = new(ServerConfig)
	// Default is the default configuration
	Default *ServerConfig = new(ServerConfig)
)

func init() {
	// 先读内置默认配置，设置为全局的配置
	_ = Default.Unmarshal([]byte(DefaultConfigJSON))

	// 当初始化完成之后应该使用此 Settings 配置
	_ = Settings.Unmarshal([]byte(DefaultConfigJSON))
}

// HasCustom checked enable custom config
func HasCustom(path string) bool {
	return path != defaultFilePath
}

func Vaildated(opt *ServerConfig) error {
	if opt.Password == "" {
		return errors.New("auth password is empty")
	}
	if opt.Path == "" {
		return errors.New("data directory path is empty")
	}
	if !(opt.Port > 1024 && opt.Port < 65535) {
		return errors.New("port range not legal")
	}
	if opt.LogPath == "" {
		return errors.New("logging output path is empty")
	}
	return nil
}

// Load through a configuration file
func Load(file string, opt *ServerConfig) error {
	// 检查文件是否存在
	_, err := os.Stat(file)
	if err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType(extension)
	v.SetConfigFile(file)

	err = v.ReadInConfig()
	if err != nil {
		return err
	}

	return v.Unmarshal(opt)
}

func saved(path string, opt *ServerConfig) error {
	// 将配置对象转换为 YAML 格式的字节数组
	yamlData, _ := yaml.Marshal(&opt)
	return os.WriteFile(path, yamlData, FsPerm)
}

// SavedAs Settings.Path 存储到磁盘文件中
func (opt *ServerConfig) SavedAs(path string) error {
	return saved(path, opt)
}

// Saved Settings.Path 存储到配置好的目录中
func (opt *ServerConfig) Saved() error {
	return saved(filepath.Join(opt.Path, fileName+"."+extension), opt)
}

func (opt *ServerConfig) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &opt)
}

func (opt *ServerConfig) Marshal() ([]byte, error) {
	return json.Marshal(opt)
}

func (opt *ServerConfig) String() string {
	return toString(opt)
}

func toString(opt *ServerConfig) string {
	bs, _ := opt.Marshal()
	return string(bs)
}

type ServerConfig struct {
	Port       int        `json:"port"`
	Path       string     `json:"path"`
	Debug      bool       `json:"debug"`
	LogPath    string     `json:"log_path"`
	Password   string     `json:"auth"`
	Compressor Compressor `json:"compressor"`
}

type Compressor struct {
	Enable bool  `json:"enable"`
	Second int64 `json:"second"`
}

// 尽量减少 conf 的配置参数侵入到其他包中
// conf 包只限于 cmd 包下使用
package conf

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	cfSuffix        = "yaml"
	defaultFileName = "config"
	defaultFilePath = ""

	// 设置默认文件系统权限
	Permissions = fs.FileMode(0755)

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

var (
	// Settings global configure options
	Settings *ServerConfig = new(ServerConfig)
	// DefaultConfig is the default configuration
	DefaultConfig *ServerConfig = new(ServerConfig)
)

func init() {
	// 先读内置默认配置，设置为全局的配置
	_ = DefaultConfig.Unmarshal([]byte(DefaultConfigJSON))

	// 当初始化完成之后应该使用此 Settings 配置
	_ = Settings.Unmarshal([]byte(DefaultConfigJSON))
}

// HasCustom checked enable custom config
func HasCustom(path string) bool {
	return path != defaultFilePath
}

// Load through a configuration file
func Load(file string, opt *ServerConfig) error {

	v := viper.New()
	v.SetConfigType(cfSuffix)
	v.SetConfigFile(file)

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	return v.Unmarshal(&opt)
}

func saved(path string, opt *ServerConfig) error {
	// 将配置对象转换为 YAML 格式的字节数组
	yamlData, _ := yaml.Marshal(&opt)
	return os.WriteFile(path, yamlData, Permissions)
}

// SavedAs Settings.Path 存储到磁盘文件中
func (opt *ServerConfig) SavedAs(path string) error {
	return saved(path, opt)
}

// Saved Settings.Path 存储到配置好的目录中
func (opt *ServerConfig) Saved() error {
	return saved(filepath.Join(opt.Path, defaultFileName+"."+cfSuffix), opt)
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
	VaseDB `json:"vasedb"`
}

type VaseDB struct {
	Port       int        `json:"port"`
	Path       string     `json:"path"`
	Mode       string     `json:"mode"`
	Debug      bool       `json:"debug"`
	FileSize   int64      `json:"filesize"`
	LogPath    string     `json:"log"`
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

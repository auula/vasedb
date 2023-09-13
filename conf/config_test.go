package conf

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	// 创建一个临时目录用于测试
	tmpDir := t.TempDir()

	// 设置 Settings.Path 为临时目录
	Settings.Path = tmpDir

	// 创建一个配置文件并写入测试数据
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	testConfigData := []byte(`
vasedb:
  port: 8080
  path: "/test/path"
  debug: true
`)

	err := os.WriteFile(configFile, testConfigData, 0644)
	if err != nil {
		t.Fatalf("Error writing test config file: %v", err)
	}

	// 调用 Load 函数
	loadedConfig := new(ServerConfig)
	err = Load(configFile, loadedConfig)
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	// 检查加载的配置是否正确
	expectedConfig := &ServerConfig{
		VaseDB: VaseDB{
			Port:  8080,
			Path:  "/test/path",
			Debug: true,
		},
	}

	// 检查比较是否一致
	if !reflect.DeepEqual(loadedConfig, expectedConfig) {
		t.Errorf("Loaded config is not as expected.\nGot: %+v\nExpected: %+v", loadedConfig, expectedConfig)
	}
}

func TestReloadConfig(t *testing.T) {

	// 创建一个临时目录用于测试
	tmpDir := t.TempDir()

	// 设置 Settings.Path 为临时目录
	Settings.Path = tmpDir

	// 创建一个配置文件并写入测试数据
	configFile := filepath.Join(tmpDir, "etc", "config.yaml")
	// 模拟文件中数据
	configData := []byte(`
        {
            "vasedb": {
                "port": 8080,
                "path": "/test/path",
                "debug": true
            }
        }
    `)

	// 设置文件系统权限
	perm := os.FileMode(0755)

	err := os.MkdirAll(filepath.Dir(configFile), perm)
	if err != nil {
		t.Fatalf("Error creating test directory: %v", err)
	}
	err = os.WriteFile(configFile, configData, perm)
	if err != nil {
		t.Fatalf("Error writing test config file: %v", err)
	}

	// 调用 ReloadConfig 函数
	reloadedConfig, err := ReloadConfig()
	if err != nil {
		t.Fatalf("Error reloading config: %v", err)
	}

	// 检查重新加载的配置是否正确
	expectedConfig := &ServerConfig{
		VaseDB: VaseDB{
			Port:  8080,
			Path:  "/test/path",
			Debug: true,
		},
	}

	// 采用深度比较是否一致
	if !reflect.DeepEqual(reloadedConfig, expectedConfig) {
		t.Errorf("Reloaded config is not as expected.\nGot: %+v\nExpected: %+v", reloadedConfig, expectedConfig)
	}
}

func TestSavedConfig(t *testing.T) {

	// 创建一个临时目录用于测试
	tmpDir := t.TempDir()

	// 创建一个 ServerConfig 实例
	config := &ServerConfig{
		VaseDB: VaseDB{
			Port:     8080,
			Path:     tmpDir,
			Debug:    true,
			Mode:     "mmap",
			FileSize: 10248080,
			Logging:  "/tmp/vasedb/out.log",
			Password: "password@123",
			Encoder: Encoder{
				Enable: true,
				Secret: "/tmp/vasedb/etc/encrypt.wasm",
			},
			Compressor: Compressor{
				Enable:  true,
				Mode:    "cycle",
				Hotspot: 1000,
				Second:  15000,
			},
		},
	}

	// 调用 Saved 函数
	err := config.Saved(tmpDir)

	if err != nil {
		t.Fatalf("Error saving config: %v", err)
	}
}

func TestIsDefault(t *testing.T) {
	tests := []struct {
		name string
		flag string
		want bool
	}{
		{
			name: "successful", flag: "default.yaml", want: true,
		},
		{
			name: "failed", flag: "", want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDefault(tt.flag); got != tt.want {
				t.Errorf("IsDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	t.Run("Test DefaultConfig Unmarshal", func(t *testing.T) {
		err := DefaultConfig.Unmarshal([]byte(DefaultConfigJSON))
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	})

	t.Run("Test Settings Unmarshal", func(t *testing.T) {
		err := Settings.Unmarshal([]byte(DefaultConfigJSON))
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	})

	if !reflect.DeepEqual(DefaultConfig, Settings) {
		t.Errorf("default config not equal settings. \nGot: %+v\nExpected: %+v", DefaultConfig, Settings)
	}
}

func TestServerConfig_Marshal(t *testing.T) {

	if err := Settings.Unmarshal([]byte(DefaultConfigJSON)); err != nil {
		t.Error(err)
	}

	bytes, err := Settings.Marshal()

	if err != nil {
		t.Error(err)
	}

	if err := DefaultConfig.Unmarshal(bytes); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(Settings, DefaultConfig) {
		t.Errorf("ServerConfig.Marshal() = %v, want %v", string(bytes), DefaultConfigJSON)
	}

}

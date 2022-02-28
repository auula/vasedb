// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/27 - 12:22 上午 - UTC/GMT+08:00

package bottle

import (
	"fmt"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {

	os.RemoveAll("./testdata/")

	type args struct {
		opt Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successful",
			args: args{
				Option{
					Directory:       "./testdata",
					DataFileMaxSize: defaultMaxFileSize,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Open(tt.args.opt); (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type userinfo struct {
	Name string
	Age  uint8
}

func TestPutANDGet(t *testing.T) {
	os.RemoveAll("./testdata/")

	err := Open(Option{
		Directory:       "./testdata",
		DataFileMaxSize: defaultMaxFileSize,
	})

	if err != nil {
		t.Error(err)
	}

	user := userinfo{
		Name: "Leon Ding",
		Age:  22,
	}

	Put([]byte("foo"), Bson(&user))

	// time.Sleep(5 * time.Second)
	var u userinfo

	Get([]byte("foo")).Unwrap(&u)

	t.Log(u)
	Close()
}

func TestSaveData(t *testing.T) {

	t.Log(active)
	err := Open(Option{
		Directory:       "./testdata",
		DataFileMaxSize: defaultMaxFileSize,
	})

	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("test_key_%d", i)
		v := fmt.Sprintf("test_value_%d", i)
		err := Put([]byte(k), []byte(v))
		if err != nil {
			t.Error(err)
		}
	}

	err = Close()
	if err != nil {
		t.Error(err)
	}
}

func TestRecoverData(t *testing.T) {
	err := Open(Option{
		Directory:       "./testdata",
		DataFileMaxSize: defaultMaxFileSize,
	})

	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("test_key_%d", i)
		if Get([]byte(k)).isError() {
			t.Log(Get([]byte(k)))
		}
	}

	err = Close()
	if err != nil {
		t.Error(err)
	}
}

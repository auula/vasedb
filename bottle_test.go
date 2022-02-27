// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/27 - 12:22 上午 - UTC/GMT+08:00

package bottle

import (
	"os"
	"testing"
	"time"
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

	Put([]byte("foo"), Bson(&user), TTL(3))

	time.Sleep(5 * time.Second)
	var u userinfo

	Get([]byte("foo")).Unwrap(&u)

	t.Log(u)
}

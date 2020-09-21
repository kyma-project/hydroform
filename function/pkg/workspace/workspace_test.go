package workspace

import (
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"reflect"
	"testing"
)

func TestInitialize(t *testing.T) {
	type args struct {
		cfg     Cfg
		dirPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Initialize(tt.args.cfg, tt.args.dirPath); (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_fromRuntime(t *testing.T) {
	type args struct {
		runtime types.Runtime
	}
	tests := []struct {
		name    string
		args    args
		want    workspace
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromRuntime(tt.args.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("fromRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromRuntime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initialize(t *testing.T) {
	type args struct {
		cfg            Cfg
		dirPath        string
		writerProvider WriterProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initialize(tt.args.cfg, tt.args.dirPath, tt.args.writerProvider); (err != nil) != tt.wantErr {
				t.Errorf("initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_workspace_build(t *testing.T) {
	type args struct {
		cfg            Cfg
		dirPath        string
		writerProvider WriterProvider
	}
	tests := []struct {
		name    string
		ws      workspace
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ws.build(tt.args.cfg, tt.args.dirPath, tt.args.writerProvider); (err != nil) != tt.wantErr {
				t.Errorf("build() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

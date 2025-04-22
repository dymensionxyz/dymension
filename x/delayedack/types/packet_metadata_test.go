package types

import (
	"reflect"
	"testing"

	"github.com/dymensionxyz/dymension/v3/utils/utransfer"
)

func Test_parsePacketMetadata(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    *Memo
		wantErr bool
	}{
		{
			"valid",
			args{
				`{"eibc":{"fee":"100"}}`,
			},
			&Memo{
				EIBC: &EIBCMemo{
					Fee: "100",
				},
			},
			false,
		},
		{
			"valid with hook",
			args{
				utransfer.CreateMemo("100", []byte{123}),
			},
			&Memo{
				EIBC: &EIBCMemo{
					Fee:              "100",
					OnCompletionHook: []byte{123},
				},
			},
			false,
		},
		{
			"invalid - misquoted fee",
			args{
				`{"eibc":{"fee":100}}`,
			},
			nil,
			true,
		},
		{
			"invalid - pfm",
			args{
				`{"forward":{}}`,
			},
			nil,
			true,
		},
		{
			"invalid - empty",
			args{
				``,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMemo(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePacketMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePacketMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}

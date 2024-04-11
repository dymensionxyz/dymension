package types

import (
	"reflect"
	"testing"
)

func Test_parsePacketMetadata(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    *PacketMetadata
		wantErr bool
	}{
		{
			"valid",
			args{
				`{"eibc":{"fee":"100"}}`,
			},
			&PacketMetadata{
				EIBC: &EIBCMetadata{
					Fee: "100",
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
			"invalid - emtpy",
			args{
				``,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePacketMetadata(tt.args.input)
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

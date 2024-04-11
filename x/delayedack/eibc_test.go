package delayedack

import (
	"reflect"
	"testing"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func Test_parsePacketMetadata(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.PacketMetadata
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"valid",
			args{
				`{"eibc":{"fee":"100"}}`,
			},
			&types.PacketMetadata{
				EIBC: &types.EIBCMetadata{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePacketMetadata(tt.args.input)
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

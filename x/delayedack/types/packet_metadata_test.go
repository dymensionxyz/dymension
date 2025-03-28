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
			"valid with hook",
			args{
				utransfer.CreateMemo("100", []byte{123}),
			},
			&PacketMetadata{
				EIBC: &EIBCMetadata{
					Fee:         "100",
					FulfillHook: []byte{123},
				},
			},
			false,
		},
		{
			"real hook",
			args{
				` "{"eibc":{"fee":"100","fulfill_hook":"Ct4BEkIweDcyNmY3NTc0NjU3MjVmNjE3MDcwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMTAwMDAwMDAwMDAwMDAwMDEYASJCMHg5MzRiODY3MDUyY2E5YzY1ZTMzMzYyMTEyZjM1ZmI1NDhmODczMmMyZmU0NWYwN2I5YzU5MTk1OGU4NjVkZWYwKgMyOTk6ATBCSgpEaWJjLzlBMUVBQ0Q1M0E2QTE5N0FEQzgxREY5QTQ5RjBDNEEyNkY3RkY2ODVBQ0Y0MTVFRTcyNkQ3RDU5Nzk2RTcxQTcSAjIwIiwKKmR5bTEzOW1xNzUyZGVseHY3OGp2dG13eGhhc3lyeWN1ZnN2cnc0YWthOQ=="}}"`,
			},
			&PacketMetadata{
				EIBC: &EIBCMetadata{
					Fee:         "100",
					FulfillHook: []byte{123},
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

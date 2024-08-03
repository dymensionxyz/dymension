package utils

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/stretchr/testify/require"
)

func TestIsValidBech32AccountAddress(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                            string
		address                         string
		matchAccountAddressBech32Prefix bool
		want                            bool
	}{
		{
			name:                            "valid bech32 account address",
			address:                         "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			matchAccountAddressBech32Prefix: true,
			want:                            true,
		},
		{
			name:                            "valid bech32 account address, Interchain Account",
			address:                         "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul",
			matchAccountAddressBech32Prefix: true,
			want:                            true,
		},
		{
			name: "reject bech32 account address which neither 20 nor 32 bytes, case 19 bytes",
			address: func() string {
				bz := make([]byte, 19)
				addr, err := bech32.ConvertAndEncode("dym", bz)
				require.NoError(t, err)
				return addr
			}(),
			want: false,
		},
		{
			name: "reject bech32 account address which neither 20 nor 32 bytes, case 21 bytes",
			address: func() string {
				bz := make([]byte, 21)
				addr, err := bech32.ConvertAndEncode("dym", bz)
				require.NoError(t, err)
				return addr
			}(),
			want: false,
		},
		{
			name: "reject bech32 account address which neither 20 nor 32 bytes, case 31 bytes",
			address: func() string {
				bz := make([]byte, 31)
				addr, err := bech32.ConvertAndEncode("dym", bz)
				require.NoError(t, err)
				return addr
			}(),
			want: false,
		},
		{
			name: "reject bech32 account address which neither 20 nor 32 bytes, case 33 bytes",
			address: func() string {
				bz := make([]byte, 33)
				addr, err := bech32.ConvertAndEncode("dym", bz)
				require.NoError(t, err)
				return addr
			}(),
			want: false,
		},
		{
			name: "reject bech32 account address which neither 20 nor 32 bytes, case 128 bytes",
			address: func() string {
				bz := make([]byte, 128)
				addr, err := bech32.ConvertAndEncode("dym", bz)
				require.NoError(t, err)
				return addr
			}(),
			want: false,
		},
		{
			name:                            "bad checksum bech32 account address",
			address:                         "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9feu",
			matchAccountAddressBech32Prefix: true,
			want:                            false,
		},
		{
			name:                            "bad bech32 account address",
			address:                         "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3",
			matchAccountAddressBech32Prefix: true,
			want:                            false,
		},
		{
			name:                            "not bech32 address",
			address:                         "0x4fea76427b8345861e80a3540a8a9d936fd39391",
			matchAccountAddressBech32Prefix: true,
			want:                            false,
		},
		{
			name:                            "not bech32 address",
			address:                         "0x4fea76427b8345861e80a3540a8a9d936fd39391",
			matchAccountAddressBech32Prefix: false,
			want:                            false,
		},
		{
			name:                            "valid bech32 account address but mis-match HRP",
			address:                         "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			matchAccountAddressBech32Prefix: true,
			want:                            false,
		},
		{
			name:                            "valid bech32 account address ignore mis-match HRP",
			address:                         "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			matchAccountAddressBech32Prefix: false,
			want:                            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid := IsValidBech32AccountAddress(tt.address, tt.matchAccountAddressBech32Prefix)
			require.Equal(t, tt.want, gotValid)
		})
	}
}

//goland:noinspection SpellCheckingInspection
func TestIsValidHexAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "allow hex address with 20 bytes",
			address: "0x1234567890123456789012345678901234567890",
			want:    true,
		},
		{
			name:    "allow hex address with 32 bytes, Interchain Account",
			address: "0x1234567890123456789012345678901234567890123456789012345678901234",
			want:    true,
		},
		{
			name:    "disallow hex address with 19 bytes",
			address: "0x123456789012345678901234567890123456789",
			want:    false,
		},
		{
			name:    "disallow hex address with 21 bytes",
			address: "0x12345678901234567890123456789012345678901",
			want:    false,
		},
		{
			name:    "disallow hex address with 31 bytes",
			address: "0x123456789012345678901234567890123456789012345678901234567890123",
			want:    false,
		},
		{
			name:    "disallow hex address with 33 bytes",
			address: "0x12345678901234567890123456789012345678901234567890123456789012345",
			want:    false,
		},
		{
			name:    "disallow valid bech32 address",
			address: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			want:    false,
		},
		{
			name:    "disallow valid bech32 address, Interchain Account",
			address: "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsValidHexAddress(tt.address))
		})
	}
}

//goland:noinspection SpellCheckingInspection
func TestGetBytesFromHexAddress(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		wantPanic bool
		want      []byte
	}{
		{
			name:    "20 bytes address",
			address: "0x1234567890123456789012345678901234567890",
			want: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
			},
		},
		{
			name:    "32 bytes address",
			address: "0x1234567890123456789012345678901234567890123456789012345678901234",
			want: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34,
			},
		},
		{
			name:      "panic if invalid hex address",
			address:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9feu",
			wantPanic: true,
		},
		{
			name:      "panic if neither 20 nor 32 bytes address",
			address:   "0x123456789012345678901234567890123456789",
			wantPanic: true,
		},
		{
			name:      "panic if neither 20 nor 32 bytes address",
			address:   "0x12345678901234567890123456789012345678901",
			wantPanic: true,
		},
		{
			name:      "panic if neither 20 nor 32 bytes address",
			address:   "0x123456789012345678901234567890123456789012345678901234567890123",
			wantPanic: true,
		},
		{
			name:      "panic if neither 20 nor 32 bytes address",
			address:   "0x12345678901234567890123456789012345678901234567890123456789012345",
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() {
					_ = GetBytesFromHexAddress(tt.address)
				})
				return
			}

			require.Equal(t, tt.want, GetBytesFromHexAddress(tt.address))
		})
	}
}

//goland:noinspection SpellCheckingInspection
func TestGetHexAddressFromBytes(t *testing.T) {
	tests := []struct {
		name      string
		bytes     []byte
		want      string
		wantPanic bool
	}{
		{
			name: "20 bytes address",
			bytes: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
			},
			want:      "0x1234567890123456789012345678901234567890",
			wantPanic: false,
		},
		{
			name: "32 bytes address",
			bytes: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34,
			},
			want:      "0x1234567890123456789012345678901234567890123456789012345678901234",
			wantPanic: false,
		},
		{
			name: "panic if neither 20 nor 32 bytes address",
			bytes: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78,
			},
			wantPanic: true,
		},
		{
			name: "panic if neither 20 nor 32 bytes address",
			bytes: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12,
			},
			wantPanic: true,
		},
		{
			name: "panic if neither 20 nor 32 bytes address",
			bytes: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12,
			},
			wantPanic: true,
		},
		{
			name: "panic if neither 20 nor 32 bytes address",
			bytes: []byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56,
			},
			wantPanic: true,
		},
		{
			name: "output must be lower case",
			bytes: []byte{
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44,
				0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee,
			},
			want: "0xaabbccddeeff112233445566778899aabbccddee",
		},
		{
			name: "output must be lower case",
			bytes: []byte{
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44,
				0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee,
				0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
				0xaa, 0xbb,
			},
			want: "0xaabbccddeeff112233445566778899aabbccddeeff112233445566778899aabb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() {
					_ = GetHexAddressFromBytes(tt.bytes)
				})

				return
			}

			got := GetHexAddressFromBytes(tt.bytes)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPossibleAccountRegardlessChain(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "pass - Dymension bech32 address",
			address: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			want:    true,
		},
		{
			name:    "pass - Cosmos bech32 address",
			address: "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
			want:    true,
		},
		{
			name:    "pass - bech32 address - Interchain Account",
			address: "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul",
			want:    true,
		},
		{
			name:    "pass - Ethereum - numeric only",
			address: "0x1234567890123456789012345678901234567890",
			want:    true,
		},
		{
			name:    "pass - Ethereum - lowercase address",
			address: "0xaabbccddeeff112233445566778899aabbccddee",
			want:    true,
		},
		{
			name:    "pass - Ethereum - checksum address",
			address: "0xAabBccddeEfF112233445566778899aaBbccDdEe",
			want:    true,
		},
		{
			name:    "pass - 0x address - 32 bytes",
			address: "0x0380d46a00e427d89f35d78b4eacb4270bd5ecfd10b64662dcfe31eb117fc62c68",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2PK (Obsolete)",
			address: "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2PKH (26 chars)",
			address: "11111111111111111111BZbvjr",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2PKH (27 chars)",
			address: "1111111111111111111114oLvT2",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2PKH (34 chars)",
			address: "12higDjoCCNXSA95xZMWUdPvXNmkAduhWv",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2SH",
			address: "342ftSRCvFHfCeFFBuz4xwbeqnDw6BGUey",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2WPKH",
			address: "bc1q34aq5drpuwy3wgl9lhup9892qp6svr8ldzyy7c",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2WSH",
			address: "bc1qeklep85ntjz4605drds6aww9u0qr46qzrv5xswd35uhjuj8ahfcqgf6hak",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - P2TR",
			address: "bc1pxwww0ct9ue7e8tdnlmug5m2tamfn7q06sahstg39ys4c9f3340qqxrdu9k",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - BC1P",
			address: "bc1prwgcpptoxrpfl5go81wpd5qlsig5yt4g7urb45e",
			want:    true,
		},
		{
			name:    "pass - Bitcoin - bech32 (Native SegWit)",
			address: "bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej",
			want:    true,
		},
		{
			name:    "pass - Avalanche - C-Chain",
			address: "0x3cA8ac240F6ebeA8684b3E629A8e8C1f0E3bC0Ff",
			want:    true,
		},
		{
			name:    "pass - Avalanche - X-Chain",
			address: "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5",
			want:    true,
		},
		{
			name:    "pass - Cardano - legacy Byron (Icarus)",
			address: "Ae2tdPwUPEZFSi1cTyL1ZL6bgixhc2vSy5heg6Zg9uP7PpumkAJ82Qprt8b",
			want:    true,
		},
		{
			name:    "pass - Cardano - legacy Byron (Daedalus)",
			address: "DdzFFzCqrhsfZHjaBunVySZBU8i9Zom7Gujham6Jz8scCcAdkDmEbD9XSdXKdBiPoa1fjgL4ksGjQXD8ZkSNHGJfT25ieA9rWNCSA5qc",
			want:    true,
		},
		{
			name:    "pass - Cardano - Shelley",
			address: "addr1q8gg2r3vf9zggn48g7m8vx62rwf6warcs4k7ej8mdzmqmesj30jz7psduyk6n4n2qrud2xlv9fgj53n6ds3t8cs4fvzs05yzmz",
			want:    true,
		},
		{
			name:    "pass - Polkadot - Polkadot SS58",
			address: "1a1LcBX6hGPKg5aQ6DXZpAHCCzWjckhea4sz3P1PvL3oc4F",
			want:    true,
		},
		{
			name:    "pass - Polkadot - Kusama SS58",
			address: "HNZata7iMYWmk5RvZRTiAsSDhV8366zq2YGb3tLH5Upf74F",
			want:    true,
		},
		{
			name:    "pass - Polkadot - Generic Substrate SS58",
			address: "5CdiCGvTEuzut954STAXRfL8Lazs3KCZa5LPpkPeqqJXdTHp",
			want:    true,
		},
		{
			name:    "pass - Polkadot - Account ID",
			address: "0x192c3c7e5789b461fbf1c7f614ba5eed0b22efc507cda60a5e7fda8e046bcdce",
			want:    true,
		},
		{
			name:    "pass - XRP Ledger - base58",
			address: "rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz",
			want:    true,
		},
		{
			name:    "pass - XRP Ledger - X-address",
			address: "XV5sbjUmgPpvXv4ixFWZ5ptAYZ6PD28Sq49uo34VyjnmK5H",
			want:    true,
		},
		{
			name:    "pass - Solana",
			address: "7EcDhSYGxXyscszYEp35KHN8vvw3svAuLKTzXwCFLtV",
			want:    true,
		},
		{
			name:    "pass - Tron - hex",
			address: "414450cf8c8b6a8229b7f628e36b3a658e84441b6f",
			want:    true,
		},
		{
			name:    "pass - Tron - base58",
			address: "TGCRkw1Vq759FBCrwxkZGgqZbRX1WkBHSu",
			want:    true,
		},
		{
			name:    "pass - XinFin - xdc",
			address: "xdc64b3b0a417775cfb441ed064611bf79826649c0f",
			want:    true,
		},
		{
			name:    "pass - XinFin - hex",
			address: "0x64b3b0a417775cfb441ed064611bf79826649c0f",
			want:    true,
		},
		{
			name:    "pass - Stellar",
			address: "GBH4TZYZ4IRCPO44CBOLFUHULU2WGALXTAVESQA6432MBJMABBB4GIYI",
			want:    true,
		},
		{
			name:    "pass - Stellar - Federated addresses",
			address: "jed*stellar.org",
			want:    true,
		},
		{
			name:    "pass - Stellar - Federated addresses",
			address: "maria@gmail.com*stellar.org",
			want:    true,
		},
		{
			name:    "pass - Klaytn",
			address: "0xBe588061d20fe359E69D78824EC45EA98C87069A",
			want:    true,
		},
		{
			name:    "pass - Neo N3",
			address: "NVeu7XqbZ6WiL1prhChC1jMWgicuWtneDP",
			want:    true,
		},
		{
			name:    "pass - Neo Legacy",
			address: "ALuhj3QNoxvAnMZsA2oKP5UxYsBmRwjwHL",
			want:    true,
		},
		{
			name:    "pass - Tezos - user tz1",
			address: "tz1YWK1gDPQx9N1Jh4JnmVre7xN6xhGGM4uC",
			want:    true,
		},
		{
			name:    "pass - Tezos - user tz3",
			address: "tz3T8djchG5FDwt7H6wEUU3sRFJwonYPqMJe",
			want:    true,
		},
		{
			name:    "pass - Tezos - Smart Contract",
			address: "KT1S5hgipNSTFehZo7v81gq6fcLChbRwptqy",
			want:    true,
		},
		{
			name:    "pass - Litecoin - Legacy",
			address: "LMHEFMwRsQ3nHDfb9zZqynLHxjuJ2hgyyW",
			want:    true,
		},
		{
			name:    "pass - Litecoin - P2SH",
			address: "MC2JYMPVWaxqUb9qUkUbjtUwoNMo1tPaLF",
			want:    true,
		},
		{
			name:    "pass - Litecoin - bech32",
			address: "ltc1qhzjptwpym9afcdjhs7jcz6fd0jma0l0rc0e5yr",
			want:    true,
		},
		{
			name:    "pass - Litecoin - bech32",
			address: "ltc1qzvcgmntglcuv4smv3lzj6k8szcvsrmvk0phrr9wfq8w493r096ssm2fgsw",
			want:    true,
		},
		{
			name:    "pass - Bitcoin Cash",
			address: "qrvax3jgtwqssnkpctlqdl0rq7rjn0l0hgny8pt0hp",
			want:    true,
		},
		{
			name:    "pass - Bitcoin Cash - with prefix",
			address: "bitcoincash:qrvax3jgtwqssnkpctlqdl0rq7rjn0l0hgny8pt0hp",
			want:    true,
		},
		{
			name:    "pass - Dogecoin",
			address: "D7wbmbjBWG5HPkT6d4gh6SdQPp6z25vcF2",
			want:    true,
		},
		{
			name:    "pass - Monero",
			address: "4BDtRc8Ym9wGFyEBzDWMSZ7iuUcNJ1ssiRkU6LjQgHURD4PGAMsZnzxAz2SGmNhinLxPF111N41bTHQBiu6QTmaZwKngDWrH",
			want:    true,
		},
		{
			name:    "pass - Zcash - Transparent Address",
			address: "t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
			want:    true,
		},
		{
			name:    "pass - Zcash - Sapling Shielded Address",
			address: "zs1z7rejlpsa98s2rrrfkwmaxu53e4ue0ulcrw0h4x5g8jl04tak0d3mm47vdtahatqrlkngh9sly",
			want:    true,
		},
		{
			name:    "pass - Zcash - Legacy Sprout Shielded Address",
			address: "zcU1Cd6zYyZCd2VJF8yKgmzjxdiiU1rgTTjEwoN1CGUWCziPkUTXUjXmX7TMqdMNsTfuiGN1jQoVN4kGxUR4sAPN4XZ7pxb",
			want:    true,
		},
		{
			name:    "pass - Dash",
			address: "XpLM8qBMd7CqukVzKXkQWuQJmgrAFb87Qr",
			want:    true,
		},
		{
			name:    "pass - Binance Smart Chain",
			address: "0x7f533b5fbf6ef86c3b7df76cc27fc67744a9a760",
			want:    true,
		},
		{
			name:    "pass - Algorand",
			address: "2UEQTE5QDNXPI7M3TU44G6SYKLFWLPQO7EBZM7K7MHMQQMFI4QJPLHQFHM",
			want:    true,
		},
		{
			name:    "pass - Algorand - typed",
			address: "ALGO-2UEQTE5QDNXPI7M3TU44G6SYKLFWLPQO7EBZM7K7MHMQQMFI4QJPLHQFHM",
			want:    true,
		},
		{
			name:    "pass - Hedera Hashgraph - without checksum",
			address: "0.0.123",
			want:    true,
		},
		{
			name:    "pass - Hedera Hashgraph - without checksum",
			address: "0.0.0",
			want:    true,
		},
		{
			name:    "pass - Hedera Hashgraph - with checksum",
			address: "0.0.123-vfmkw",
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.address)
			got := PossibleAccountRegardlessChain(tt.address)
			require.Equal(t, tt.want, got)
		})
	}
}

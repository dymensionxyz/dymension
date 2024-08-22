package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestIsValidDymName(t *testing.T) {
	t.Run("maximum accepted length is 20", func(t *testing.T) {
		require.True(t, IsValidDymName("12345678901234567890"))
		require.False(t, IsValidDymName("123456789012345678901"))
	})

	t.Run("not allow empty", func(t *testing.T) {
		require.False(t, IsValidDymName(""))
	})

	t.Run("single character", func(t *testing.T) {
		for i := 'a'; i <= 'z'; i++ {
			require.Truef(t, IsValidDymName(string(i)), "failed for single char '%c'", i)
		}
		for i := 'A'; i <= 'Z'; i++ {
			require.Falsef(t, IsValidDymName(string(i)), "should not accept '%c'", i)
		}
		for i := '0'; i <= '9'; i++ {
			require.Truef(t, IsValidDymName(string(i)), "failed for single char '%c'", i)
		}
		require.False(t, IsValidDymName("-"), "should not accept single dash")
		require.False(t, IsValidDymName("_"), "should not accept single underscore")
	})

	t.Run("not starts or ends with dash or underscore", func(t *testing.T) {
		for _, prototype := range []string{"a", "aa", "aaa", "8"} {
			check := func(dymName string) {
				require.Falsef(t, IsValidDymName(dymName), "should not accept '%s'", dymName)
			}
			check(prototype + "-")
			check(prototype + "_")
			check("-" + prototype)
			check("_" + prototype)
		}
	})

	tests := []struct {
		dymName string
		invalid bool
	}{
		{
			dymName: "a",
		},
		{
			dymName: "aa",
		},
		{
			dymName: "9",
		},
		{
			dymName: "9999",
		},
		{
			dymName: "-",
			invalid: true,
		},
		{
			dymName: "_",
			invalid: true,
		},
		{
			dymName: "9-",
			invalid: true,
		},
		{
			dymName: "9_",
			invalid: true,
		},
		{
			dymName: "-9",
			invalid: true,
		},
		{
			dymName: "_9",
			invalid: true,
		},
		{
			dymName: "_9",
			invalid: true,
		},
		{
			dymName: "a_9",
		},
		{
			dymName: "a-9",
		},
		{
			dymName: "a--9",
			invalid: true,
		},
		{
			dymName: "a__9",
			invalid: true,
		},
		{
			dymName: "a.dym",
			invalid: true,
		},
		{
			dymName: "a a",
			invalid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.dymName, func(t *testing.T) {
			if tt.invalid {
				require.Falsef(t, IsValidDymName(tt.dymName), "should not accept '%s'", tt.dymName)
			} else {
				require.Truef(t, IsValidDymName(tt.dymName), "should accept '%s'", tt.dymName)
			}
		})
	}

	t.Run("not allow hex address", func(t *testing.T) {
		require.False(t, IsValidDymName("0x1234567890123456789012345678901234567890"))
		require.False(t, IsValidDymName("0x1234567890123456789012345678901234567890123456789012345678901234"))
		require.False(t, IsValidDymName("dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96"))
		require.False(t, IsValidDymName("dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul"))
	})
}

func TestIsValidSubDymName(t *testing.T) {
	t.Run("maximum accepted length is 20", func(t *testing.T) {
		require.True(t, IsValidSubDymName("12345678901234567890"))
		require.False(t, IsValidSubDymName("123456789012345678901.12345678901234567890"))
		require.False(t, IsValidSubDymName("1234567890123456789012345678901234567890123456789012345678901234567.12345678901234567890"))
	})

	t.Run("allow empty", func(t *testing.T) {
		require.True(t, IsValidSubDymName(""))
	})

	t.Run("single character", func(t *testing.T) {
		for i := 'a'; i <= 'z'; i++ {
			require.Truef(t, IsValidSubDymName(string(i)), "failed for single char '%c'", i)
		}
		for i := 'A'; i <= 'Z'; i++ {
			require.Falsef(t, IsValidSubDymName(string(i)), "should not accept '%c'", i)
		}
		for i := '0'; i <= '9'; i++ {
			require.Truef(t, IsValidSubDymName(string(i)), "failed for single char '%c'", i)
		}
		require.False(t, IsValidSubDymName("-"), "should not accept single dash")
		require.False(t, IsValidSubDymName("_"), "should not accept single underscore")
	})

	t.Run("not starts or ends with dash or underscore or dot", func(t *testing.T) {
		for _, prototype := range []string{"a", "aa", "aaa", "8"} {
			check := func(dymName string) {
				require.Falsef(t, IsValidSubDymName(dymName), "should not accept '%s'", dymName)
			}
			check(prototype + "-")
			check(prototype + "_")
			check("-" + prototype)
			check("_" + prototype)
			check(prototype + ".")
			check("." + prototype)
		}
	})

	tests := []struct {
		subDymName string
		invalid    bool
	}{
		{
			subDymName: "a",
		},
		{
			subDymName: "a.a",
		},
		{
			subDymName: "aa",
		},
		{
			subDymName: "aa.aa",
		},
		{
			subDymName: "9",
		},
		{
			subDymName: "9.9",
		},
		{
			subDymName: "9999",
		},
		{
			subDymName: "9999.9999",
		},
		{
			subDymName: "-",
			invalid:    true,
		},
		{
			subDymName: "-.-",
			invalid:    true,
		},
		{
			subDymName: "-.a",
			invalid:    true,
		},
		{
			subDymName: "_",
			invalid:    true,
		},
		{
			subDymName: "_._",
			invalid:    true,
		},
		{
			subDymName: "a._",
			invalid:    true,
		},
		{
			subDymName: "9-",
			invalid:    true,
		},
		{
			subDymName: "a.9-",
			invalid:    true,
		},
		{
			subDymName: "9_",
			invalid:    true,
		},
		{
			subDymName: "9_.a",
			invalid:    true,
		},
		{
			subDymName: "-9",
			invalid:    true,
		},
		{
			subDymName: "-9.a",
			invalid:    true,
		},
		{
			subDymName: "_9",
			invalid:    true,
		},
		{
			subDymName: "a._9",
			invalid:    true,
		},
		{
			subDymName: "_9",
			invalid:    true,
		},
		{
			subDymName: "a_9",
		},
		{
			subDymName: "a_9.a_9",
		},
		{
			subDymName: "a-9",
		},
		{
			subDymName: "a-9.a-9",
		},
		{
			subDymName: "a--9",
			invalid:    true,
		},
		{
			subDymName: "a--9.a",
			invalid:    true,
		},
		{
			subDymName: "a__9",
			invalid:    true,
		},
		{
			subDymName: "a.a__9",
			invalid:    true,
		},
		{
			subDymName: "a.dym",
		},
		{
			subDymName: "a a",
			invalid:    true,
		},
		{
			subDymName: "a a.a",
			invalid:    true,
		},
		{
			subDymName: "aa..a",
			invalid:    true,
		},
		{
			subDymName: "aa. .a",
			invalid:    true,
		},
		{
			subDymName: "a .a",
			invalid:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.subDymName, func(t *testing.T) {
			if tt.invalid {
				require.Falsef(t, IsValidSubDymName(tt.subDymName), "should not accept '%s'", tt.subDymName)
			} else {
				require.Truef(t, IsValidSubDymName(tt.subDymName), "should accept '%s'", tt.subDymName)
			}
		})
	}
}

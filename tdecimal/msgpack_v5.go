//go:build go_tarantool_msgpack_v5
// +build go_tarantool_msgpack_v5

package tdecimal

import (
	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	msgpack.RegisterExt(decimalExtID, (*Decimal)(nil))
}

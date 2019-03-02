package log

import (
	"net"
	"time"

	"github.com/bloom42/rz-go"
)

type Field func(event *rz.Event)

func String(key, val string) Field {
	return func(event *rz.Event) {
		event.String(key, val)
	}
}

func Strings(key string, vals []string) Field {
	return func(event *rz.Event) {
		event.Strings(key, vals)
	}
}

func Bytes(key string, val []byte) Field {
	return func(event *rz.Event) {
		event.Bytes(key, val)
	}
}

func Hex(key string, val []byte) Field {
	return func(event *rz.Event) {
		event.Hex(key, val)
	}
}

func RawJSON(key string, b []byte) Field {
	return func(event *rz.Event) {
		event.RawJSON(key, b)
	}
}

func Error(key string, err error) Field {
	return func(event *rz.Event) {
		event.Error(key, err)
	}
}

func Errors(key string, errs []error) Field {
	return func(event *rz.Event) {
		event.Errors(key, errs)
	}

}

func Err(err error) Field {
	return func(event *rz.Event) {
		event.Err(err)
	}
}

func Stack() Field {
	return func(event *rz.Event) {
		event.Stack()
	}
}

func Bool(key string, b bool) Field {
	return func(event *rz.Event) {
		event.Bool(key, b)
	}
}

func Bools(key string, b []bool) Field {
	return func(event *rz.Event) {
		event.Bools(key, b)
	}
}

func Int(key string, i int) Field {
	return func(event *rz.Event) {
		event.Int(key, i)
	}
}

func Ints(key string, i []int) Field {
	return func(event *rz.Event) {
		event.Ints(key, i)
	}
}

func Int8(key string, i int8) Field {
	return func(event *rz.Event) {
		event.Int8(key, i)
	}
}

func Ints8(key string, i []int8) Field {
	return func(event *rz.Event) {
		event.Ints8(key, i)
	}
}

func Int16(key string, i int16) Field {
	return func(event *rz.Event) {
		event.Int16(key, i)
	}
}

func Ints16(key string, i []int16) Field {
	return func(event *rz.Event) {
		event.Ints16(key, i)
	}
}

func Int32(key string, i int32) Field {
	return func(event *rz.Event) {
		event.Int32(key, i)
	}
}

func Ints32(key string, i []int32) Field {
	return func(event *rz.Event) {
		event.Ints32(key, i)
	}
}

func Int64(key string, i int64) Field {
	return func(event *rz.Event) {
		event.Int64(key, i)
	}
}

func Ints64(key string, i []int64) Field {
	return func(event *rz.Event) {
		event.Ints64(key, i)
	}
}

func Uint(key string, i uint) Field {
	return func(event *rz.Event) {
		event.Uint(key, i)
	}
}

func Uints(key string, i []uint) Field {
	return func(event *rz.Event) {
		event.Uints(key, i)
	}
}

func Uint8(key string, i uint8) Field {
	return func(event *rz.Event) {
		event.Uint8(key, i)
	}
}

func Uints8(key string, i []uint8) Field {
	return func(event *rz.Event) {
		event.Uints8(key, i)
	}
}

func Uint16(key string, i uint16) Field {
	return func(event *rz.Event) {
		event.Uint16(key, i)
	}
}

func Uints16(key string, i []uint16) Field {
	return func(event *rz.Event) {
		event.Uints16(key, i)
	}
}

func Uint32(key string, i uint32) Field {
	return func(event *rz.Event) {
		event.Uint32(key, i)
	}
}

func Uints32(key string, i []uint32) Field {
	return func(event *rz.Event) {
		event.Uints32(key, i)
	}
}

func Uint64(key string, i uint64) Field {
	return func(event *rz.Event) {
		event.Uint64(key, i)
	}
}

func Uints64(key string, i []uint64) Field {
	return func(event *rz.Event) {
		event.Uints64(key, i)
	}
}

func Float32(key string, f float32) Field {
	return func(event *rz.Event) {
		event.Float32(key, f)
	}
}

func Floats32(key string, f []float32) Field {
	return func(event *rz.Event) {
		event.Floats32(key, f)
	}
}

func Float64(key string, f float64) Field {
	return func(event *rz.Event) {
		event.Float64(key, f)
	}
}

func Floats64(key string, f []float64) Field {
	return func(event *rz.Event) {
		event.Floats64(key, f)
	}
}

func Timestamp() Field {
	return func(event *rz.Event) {
		event.Timestamp()
	}
}

func Time(key string, t time.Time) Field {
	return func(event *rz.Event) {
		event.Time(key, t)
	}
}

func Times(key string, t []time.Time) Field {
	return func(event *rz.Event) {
		event.Times(key, t)
	}
}

func Duration(key string, d time.Duration) Field {
	return func(event *rz.Event) {
		event.Duration(key, d)
	}
}

func Durations(key string, d []time.Duration) Field {
	return func(event *rz.Event) {
		event.Durations(key, d)
	}
}

func TimeDiff(key string, t time.Time, start time.Time) Field {
	return func(event *rz.Event) {
		event.TimeDiff(key, t, start)
	}
}

func Interface(key string, i interface{}) Field {
	return func(event *rz.Event) {
		event.Interface(key, i)
	}
}

func Caller() Field {
	return func(event *rz.Event) {
		event.Caller()
	}
}

func IP(key string, ip net.IP) Field {
	return func(event *rz.Event) {
		event.IP(key, ip)
	}
}

func IPNet(key string, pfx net.IPNet) Field {
	return func(event *rz.Event) {
		event.IPNet(key, pfx)
	}
}

func MACAddr(key string, ha net.HardwareAddr) Field {
	return func(event *rz.Event) {
		event.MACAddr(key, ha)
	}
}

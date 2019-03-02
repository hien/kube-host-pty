package rz

import (
	"net"
	"sync"
	"time"
)

var eventPool = &sync.Pool{
	New: func() interface{} {
		return &Event{
			buf: make([]byte, 0, 500),
		}
	},
}

// Event represents a log event. It is instanced by one of the level method.
type Event struct {
	buf                  []byte
	w                    LevelWriter
	level                LogLevel
	done                 func(msg string)
	stack                bool      // enable error stack trace
	caller               bool      // enable caller field
	timestamp            bool      // enable timestamp
	ch                   []LogHook // hooks from context
	timestampFieldName   string
	levelFieldName       string
	messageFieldName     string
	errorFieldName       string
	callerFieldName      string
	errorStackFieldName  string
	timeFieldFormat      string
	callerSkipFrameCount int
	formatter            LogFormatter
	timestampFunc        func() time.Time
}

func putEvent(e *Event) {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16 // 64KiB
	if cap(e.buf) > maxSize {
		return
	}
	eventPool.Put(e)
}

// LogObjectMarshaler provides a strongly-typed and encoding-agnostic interface
// to be implemented by types used with Event/Context's Object methods.
type LogObjectMarshaler interface {
	MarshalRzObject(e *Event)
}

// LogArrayMarshaler provides a strongly-typed and encoding-agnostic interface
// to be implemented by types used with Event/Context's Array methods.
type LogArrayMarshaler interface {
	MarshalRzArray(a *Array)
}

func newEvent(w LevelWriter, level LogLevel) *Event {
	e := eventPool.Get().(*Event)
	e.buf = e.buf[:0]
	e.ch = nil
	e.buf = enc.AppendBeginMarker(e.buf)
	e.w = w
	e.level = level
	return e
}

// Enabled return false if the *Event is going to be filtered out by
// log level or sampling.
func (e *Event) Enabled() bool {
	return e.level != Disabled
}

// Discard disables the event
func (e *Event) Discard() *Event {
	e.level = Disabled
	return e
}

// Fields is a helper function to use a map to set fields using type assertion.
func (e *Event) Fields(fields map[string]interface{}) *Event {
	e.buf = e.appendFields(e.buf, fields)
	return e
}

// Dict adds the field key with a dict to the event context.
// Use rz.Dict() to create the dictionary.
func (e *Event) Dict(key string, dict *Event) *Event {
	dict.buf = enc.AppendEndMarker(dict.buf)
	e.buf = append(enc.AppendKey(e.buf, key), dict.buf...)
	putEvent(dict)
	return e
}

// Dict creates an Event to be used with the *Event.Dict method.
// Call usual field methods like Str, Int etc to add fields to this
// event and give it as argument the *Event.Dict method.
func Dict() *Event {
	return newEvent(nil, 0)
}

// Array adds the field key with an array to the event context.
// Use Event.Arr() to create the array or pass a type that
// implement the LogArrayMarshaler interface.
func (e *Event) Array(key string, arr LogArrayMarshaler) *Event {
	e.buf = enc.AppendKey(e.buf, key)
	var a *Array
	if aa, ok := arr.(*Array); ok {
		a = aa
	} else {
		a = e.Arr()
		arr.MarshalRzArray(a)
	}
	e.buf = a.write(e.buf)
	return e
}

func (e *Event) appendObject(obj LogObjectMarshaler) {
	e.buf = enc.AppendBeginMarker(e.buf)
	obj.MarshalRzObject(e)
	e.buf = enc.AppendEndMarker(e.buf)
}

// Object marshals an object that implement the LogObjectMarshaler interface.
func (e *Event) Object(key string, obj LogObjectMarshaler) *Event {
	e.buf = enc.AppendKey(e.buf, key)
	e.appendObject(obj)
	return e
}

// EmbedObject marshals an object that implement the LogObjectMarshaler interface.
func (e *Event) EmbedObject(obj LogObjectMarshaler) *Event {
	obj.MarshalRzObject(e)
	return e
}

// String adds the field key with val as a string to the *Event context.
func (e *Event) String(key, val string) *Event {
	e.buf = enc.AppendString(enc.AppendKey(e.buf, key), val)
	return e
}

// Strings adds the field key with vals as a []string to the *Event context.
func (e *Event) Strings(key string, vals []string) *Event {
	e.buf = enc.AppendStrings(enc.AppendKey(e.buf, key), vals)
	return e
}

// Bytes adds the field key with val as a string to the *Event context.
//
// Runes outside of normal ASCII ranges will be hex-encoded in the resulting
// JSON.
func (e *Event) Bytes(key string, val []byte) *Event {
	e.buf = enc.AppendBytes(enc.AppendKey(e.buf, key), val)
	return e
}

// Hex adds the field key with val as a hex string to the *Event context.
func (e *Event) Hex(key string, val []byte) *Event {
	e.buf = enc.AppendHex(enc.AppendKey(e.buf, key), val)
	return e
}

// RawJSON adds already encoded JSON to the log line under key.
//
// No sanity check is performed on b; it must not contain carriage returns and
// be valid JSON.
func (e *Event) RawJSON(key string, b []byte) *Event {
	e.buf = appendJSON(enc.AppendKey(e.buf, key), b)
	return e
}

// Error adds the field key with serialized err to the *Event context.
// If err is nil, no field is added.
func (e *Event) Error(key string, err error) *Event {
	switch m := ErrorMarshalFunc(err).(type) {
	case nil:
		return e
	case LogObjectMarshaler:
		return e.Object(key, m)
	case error:
		return e.String(key, m.Error())
	case string:
		return e.String(key, m)
	default:
		return e.Interface(key, m)
	}
}

// Errors adds the field key with errs as an array of serialized errors to the
// *Event context.
func (e *Event) Errors(key string, errs []error) *Event {
	arr := e.Arr()
	for _, err := range errs {
		switch m := ErrorMarshalFunc(err).(type) {
		case LogObjectMarshaler:
			arr = arr.Object(m)
		case error:
			arr = arr.Err(m)
		case string:
			arr = arr.Str(m)
		default:
			arr = arr.Interface(m)
		}
	}

	return e.Array(key, arr)
}

// Err adds the field "error" with serialized err to the *Event context.
// If err is nil, no field is added.
// To customize the key name, uze rz.ErrorFieldName.
////
// If Stack() has been called before and rz.ErrorStackMarshaler is defined,
// the err is passed to ErrorStackMarshaler and the result is appended to the
// rz.ErrorStackFieldName.
func (e *Event) Err(err error) *Event {
	if e.stack && ErrorStackMarshaler != nil {
		switch m := ErrorStackMarshaler(err).(type) {
		case nil:
		case LogObjectMarshaler:
			e.Object(e.errorStackFieldName, m)
		case error:
			e.String(e.errorStackFieldName, m.Error())
		case string:
			e.String(e.errorStackFieldName, m)
		default:
			e.Interface(e.errorStackFieldName, m)
		}
	}
	return e.Error(e.errorFieldName, err)
}

// Stack enables stack trace printing for the error passed to Err().
//
// ErrorStackMarshaler must be set for this method to do something.
func (e *Event) Stack() *Event {
	e.stack = true
	return e
}

// Bool adds the field key with val as a bool to the *Event context.
func (e *Event) Bool(key string, b bool) *Event {
	if e == nil {
		return e
	}
	e.buf = enc.AppendBool(enc.AppendKey(e.buf, key), b)
	return e
}

// Bools adds the field key with val as a []bool to the *Event context.
func (e *Event) Bools(key string, b []bool) *Event {
	e.buf = enc.AppendBools(enc.AppendKey(e.buf, key), b)
	return e
}

// Int adds the field key with i as a int to the *Event context.
func (e *Event) Int(key string, i int) *Event {
	e.buf = enc.AppendInt(enc.AppendKey(e.buf, key), i)
	return e
}

// Ints adds the field key with i as a []int to the *Event context.
func (e *Event) Ints(key string, i []int) *Event {
	e.buf = enc.AppendInts(enc.AppendKey(e.buf, key), i)
	return e
}

// Int8 adds the field key with i as a int8 to the *Event context.
func (e *Event) Int8(key string, i int8) *Event {
	e.buf = enc.AppendInt8(enc.AppendKey(e.buf, key), i)
	return e
}

// Ints8 adds the field key with i as a []int8 to the *Event context.
func (e *Event) Ints8(key string, i []int8) *Event {
	e.buf = enc.AppendInts8(enc.AppendKey(e.buf, key), i)
	return e
}

// Int16 adds the field key with i as a int16 to the *Event context.
func (e *Event) Int16(key string, i int16) *Event {
	e.buf = enc.AppendInt16(enc.AppendKey(e.buf, key), i)
	return e
}

// Ints16 adds the field key with i as a []int16 to the *Event context.
func (e *Event) Ints16(key string, i []int16) *Event {
	e.buf = enc.AppendInts16(enc.AppendKey(e.buf, key), i)
	return e
}

// Int32 adds the field key with i as a int32 to the *Event context.
func (e *Event) Int32(key string, i int32) *Event {
	e.buf = enc.AppendInt32(enc.AppendKey(e.buf, key), i)
	return e
}

// Ints32 adds the field key with i as a []int32 to the *Event context.
func (e *Event) Ints32(key string, i []int32) *Event {
	e.buf = enc.AppendInts32(enc.AppendKey(e.buf, key), i)
	return e
}

// Int64 adds the field key with i as a int64 to the *Event context.
func (e *Event) Int64(key string, i int64) *Event {
	e.buf = enc.AppendInt64(enc.AppendKey(e.buf, key), i)
	return e
}

// Ints64 adds the field key with i as a []int64 to the *Event context.
func (e *Event) Ints64(key string, i []int64) *Event {
	e.buf = enc.AppendInts64(enc.AppendKey(e.buf, key), i)
	return e
}

// Uint adds the field key with i as a uint to the *Event context.
func (e *Event) Uint(key string, i uint) *Event {
	e.buf = enc.AppendUint(enc.AppendKey(e.buf, key), i)
	return e
}

// Uints adds the field key with i as a []int to the *Event context.
func (e *Event) Uints(key string, i []uint) *Event {
	e.buf = enc.AppendUints(enc.AppendKey(e.buf, key), i)
	return e
}

// Uint8 adds the field key with i as a uint8 to the *Event context.
func (e *Event) Uint8(key string, i uint8) *Event {
	e.buf = enc.AppendUint8(enc.AppendKey(e.buf, key), i)
	return e
}

// Uints8 adds the field key with i as a []int8 to the *Event context.
func (e *Event) Uints8(key string, i []uint8) *Event {
	e.buf = enc.AppendUints8(enc.AppendKey(e.buf, key), i)
	return e
}

// Uint16 adds the field key with i as a uint16 to the *Event context.
func (e *Event) Uint16(key string, i uint16) *Event {
	e.buf = enc.AppendUint16(enc.AppendKey(e.buf, key), i)
	return e
}

// Uints16 adds the field key with i as a []int16 to the *Event context.
func (e *Event) Uints16(key string, i []uint16) *Event {
	e.buf = enc.AppendUints16(enc.AppendKey(e.buf, key), i)
	return e
}

// Uint32 adds the field key with i as a uint32 to the *Event context.
func (e *Event) Uint32(key string, i uint32) *Event {
	e.buf = enc.AppendUint32(enc.AppendKey(e.buf, key), i)
	return e
}

// Uints32 adds the field key with i as a []int32 to the *Event context.
func (e *Event) Uints32(key string, i []uint32) *Event {
	e.buf = enc.AppendUints32(enc.AppendKey(e.buf, key), i)
	return e
}

// Uint64 adds the field key with i as a uint64 to the *Event context.
func (e *Event) Uint64(key string, i uint64) *Event {
	e.buf = enc.AppendUint64(enc.AppendKey(e.buf, key), i)
	return e
}

// Uints64 adds the field key with i as a []int64 to the *Event context.
func (e *Event) Uints64(key string, i []uint64) *Event {
	e.buf = enc.AppendUints64(enc.AppendKey(e.buf, key), i)
	return e
}

// Float32 adds the field key with f as a float32 to the *Event context.
func (e *Event) Float32(key string, f float32) *Event {
	e.buf = enc.AppendFloat32(enc.AppendKey(e.buf, key), f)
	return e
}

// Floats32 adds the field key with f as a []float32 to the *Event context.
func (e *Event) Floats32(key string, f []float32) *Event {
	e.buf = enc.AppendFloats32(enc.AppendKey(e.buf, key), f)
	return e
}

// Float64 adds the field key with f as a float64 to the *Event context.
func (e *Event) Float64(key string, f float64) *Event {
	e.buf = enc.AppendFloat64(enc.AppendKey(e.buf, key), f)
	return e
}

// Floats64 adds the field key with f as a []float64 to the *Event context.
func (e *Event) Floats64(key string, f []float64) *Event {
	e.buf = enc.AppendFloats64(enc.AppendKey(e.buf, key), f)
	return e
}

// Timestamp adds the current local time as UNIX timestamp to the *Event context with the
// logger.TimestampFieldName key.
func (e *Event) Timestamp() *Event {
	e.timestamp = false
	e.buf = enc.AppendTime(enc.AppendKey(e.buf, e.timestampFieldName), e.timestampFunc(), e.timeFieldFormat)
	return e
}

// Time adds the field key with t formated as string using rz.TimeFieldFormat.
func (e *Event) Time(key string, t time.Time) *Event {
	e.buf = enc.AppendTime(enc.AppendKey(e.buf, key), t, e.timeFieldFormat)
	return e
}

// Times adds the field key with t formated as string using rz.TimeFieldFormat.
func (e *Event) Times(key string, t []time.Time) *Event {
	e.buf = enc.AppendTimes(enc.AppendKey(e.buf, key), t, e.timeFieldFormat)
	return e
}

// Duration adds the field key with duration d stored as rz.DurationFieldUnit.
// If rz.DurationFieldInteger is true, durations are rendered as integer
// instead of float.
func (e *Event) Duration(key string, d time.Duration) *Event {
	e.buf = enc.AppendDuration(enc.AppendKey(e.buf, key), d, DurationFieldUnit, DurationFieldInteger)
	return e
}

// Durations adds the field key with duration d stored as rz.DurationFieldUnit.
// If rz.DurationFieldInteger is true, durations are rendered as integer
// instead of float.
func (e *Event) Durations(key string, d []time.Duration) *Event {
	e.buf = enc.AppendDurations(enc.AppendKey(e.buf, key), d, DurationFieldUnit, DurationFieldInteger)
	return e
}

// TimeDiff adds the field key with positive duration between time t and start.
// If time t is not greater than start, duration will be 0.
// Duration format follows the same principle as Dur().
func (e *Event) TimeDiff(key string, t time.Time, start time.Time) *Event {
	var d time.Duration
	if t.After(start) {
		d = t.Sub(start)
	}
	e.buf = enc.AppendDuration(enc.AppendKey(e.buf, key), d, DurationFieldUnit, DurationFieldInteger)
	return e
}

// Interface adds the field key with i marshaled using reflection.
func (e *Event) Interface(key string, i interface{}) *Event {
	if obj, ok := i.(LogObjectMarshaler); ok {
		return e.Object(key, obj)
	}
	e.buf = enc.AppendInterface(enc.AppendKey(e.buf, key), i)
	return e
}

// Caller adds the file:line of the caller with the rz.CallerFieldName key.
func (e *Event) Caller() *Event {
	e.caller = true
	return e
}

// IP adds IPv4 or IPv6 Address to the event
func (e *Event) IP(key string, ip net.IP) *Event {
	e.buf = enc.AppendIPAddr(enc.AppendKey(e.buf, key), ip)
	return e
}

// IPNet adds IPv4 or IPv6 Prefix (address and mask) to the event
func (e *Event) IPNet(key string, pfx net.IPNet) *Event {
	e.buf = enc.AppendIPPrefix(enc.AppendKey(e.buf, key), pfx)
	return e
}

// MACAddr adds MAC address to the event
func (e *Event) MACAddr(key string, ha net.HardwareAddr) *Event {
	e.buf = enc.AppendMACAddr(enc.AppendKey(e.buf, key), ha)
	return e
}

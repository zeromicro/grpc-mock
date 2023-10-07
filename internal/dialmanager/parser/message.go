package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
)

type Message struct {
	msg         *desc.MessageDescriptor
	randDefault bool
}

func NewMessage(msg *desc.MessageDescriptor, randDefault bool) *Message {
	return &Message{
		msg:         msg,
		randDefault: randDefault,
	}
}

func (m *Message) MarshalJSONPB(opts *jsonpb.Marshaler) ([]byte, error) {
	var b indentBuffer
	b.indent = opts.Indent
	if len(opts.Indent) == 0 {
		b.indentCount = -1
	}
	b.comma = true
	if err := m.marshalJSON(&b, opts); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (m *Message) marshalJSON(b *indentBuffer, opts *jsonpb.Marshaler) error {
	if m == nil {
		_, err := b.WriteString("null")
		return err
	}

	err := b.WriteByte('{')
	if err != nil {
		return err
	}
	err = b.start()
	if err != nil {
		return err
	}
	first := true

	for _, fd := range m.msg.GetFields() {
		v := m.getValue(fd)

		err := b.maybeNext(&first)
		if err != nil {
			return err
		}

		err = marshalKnownFieldJSON(b, fd, v, opts, m.randDefault)
		if err != nil {
			return err
		}
	}

	err = b.end()
	if err != nil {
		return err
	}

	err = b.WriteByte('}')
	if err != nil {
		return err
	}

	return nil
}

func (m *Message) FindFieldDescriptor(tagNumber int32) *desc.FieldDescriptor {
	return m.msg.FindFieldByNumber(tagNumber)
}

func (m *Message) getValue(fd *desc.FieldDescriptor) interface{} {
	v := fd.GetDefaultValue()
	if !m.randDefault || v == nil {
		return v
	}
	switch v.(type) {
	case int32:
		return int32(time.Now().Nanosecond() % 1000)
	case int64:
		return int64(time.Now().Nanosecond() % 1000)
	case uint32:
		return uint32(time.Now().Nanosecond() % 1000)
	case uint64:
		return uint64(time.Now().Nanosecond() % 1000)
	case float64:
		return float64(time.Now().Nanosecond() % 1000)
	case float32:
		return float32(time.Now().Nanosecond() % 1000)
	case bool:
		return time.Now().Nanosecond()%2 == 0
	case string:
		return RandStringRunes(5)
	}
	return v
}

func marshalKnownFieldJSON(b *indentBuffer, fd *desc.FieldDescriptor, v interface{},
	opts *jsonpb.Marshaler, randDefault bool) error {
	var jsonName string
	if opts.OrigName {
		jsonName = fd.GetName()
	} else {
		jsonName = fd.AsFieldDescriptorProto().GetJsonName()
		if jsonName == "" {
			jsonName = fd.GetName()
		}
	}
	if fd.IsExtension() {
		var scope string
		switch parent := fd.GetParent().(type) {
		case *desc.FileDescriptor:
			scope = parent.GetPackage()
		default:
			scope = parent.GetFullyQualifiedName()
		}
		if scope == "" {
			jsonName = fmt.Sprintf("[%s]", jsonName)
		} else {
			jsonName = fmt.Sprintf("[%s.%s]", scope, jsonName)
		}
	}
	err := writeJsonString(b, jsonName)
	if err != nil {
		return err
	}
	err = b.sep()
	if err != nil {
		return err
	}

	if isNil(v) {
		if fd.GetMessageType() != nil {
			bs, err := NewMessage(fd.GetMessageType(), randDefault).MarshalJSONPB(opts)
			if err != nil {
				_, err := b.WriteString("null")
				return err
			}
			_, err = b.Write(bs)
			return err
		}
		_, err := b.WriteString("null")
		return err
	}

	if fd.IsMap() {
		err = b.WriteByte('{')
		if err != nil {
			return err
		}
		err = b.start()
		if err != nil {
			return err
		}

		md := fd.GetMessageType()
		vfd := md.FindFieldByNumber(2)

		mp := v.(map[interface{}]interface{})
		keys := make([]interface{}, 0, len(mp))
		for k := range mp {
			keys = append(keys, k)
		}
		sort.Sort(sortable(keys))
		first := true
		for _, mk := range keys {
			mv := mp[mk]
			err := b.maybeNext(&first)
			if err != nil {
				return err
			}

			err = marshalKnownFieldMapEntryJSON(b, mk, vfd, mv, opts, randDefault)
			if err != nil {
				return err
			}
		}

		err = b.end()
		if err != nil {
			return err
		}
		return b.WriteByte('}')

	} else if fd.IsRepeated() {
		err = b.WriteByte('[')
		if err != nil {
			return err
		}
		err = b.start()
		if err != nil {
			return err
		}

		sl := v.([]interface{})
		first := true
		for _, slv := range sl {
			err := b.maybeNext(&first)
			if err != nil {
				return err
			}
			err = marshalKnownFieldValueJSON(b, fd, slv, opts, randDefault)
			if err != nil {
				return err
			}
		}

		err = b.end()
		if err != nil {
			return err
		}
		return b.WriteByte(']')

	} else {
		return marshalKnownFieldValueJSON(b, fd, v, opts, randDefault)
	}
}

func marshalKnownFieldValueJSON(b *indentBuffer, fd *desc.FieldDescriptor, v interface{},
	opts *jsonpb.Marshaler, randDefault bool) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int64:
		return writeJsonString(b, strconv.FormatInt(rv.Int(), 10))
	case reflect.Int32:
		ed := fd.GetEnumType()
		if !opts.EnumsAsInts && ed != nil {
			n := int32(rv.Int())
			vd := ed.FindValueByNumber(n)
			if vd == nil {
				_, err := b.WriteString(strconv.FormatInt(rv.Int(), 10))
				return err
			} else {
				return writeJsonString(b, vd.GetName())
			}
		} else {
			_, err := b.WriteString(strconv.FormatInt(rv.Int(), 10))
			return err
		}
	case reflect.Uint64:
		return writeJsonString(b, strconv.FormatUint(rv.Uint(), 10))
	case reflect.Uint32:
		_, err := b.WriteString(strconv.FormatUint(rv.Uint(), 10))
		return err
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		var str string
		if math.IsNaN(f) {
			str = `"NaN"`
		} else if math.IsInf(f, 1) {
			str = `"Infinity"`
		} else if math.IsInf(f, -1) {
			str = `"-Infinity"`
		} else {
			var bits int
			if rv.Kind() == reflect.Float32 {
				bits = 32
			} else {
				bits = 64
			}
			str = strconv.FormatFloat(rv.Float(), 'g', -1, bits)
		}
		_, err := b.WriteString(str)
		return err
	case reflect.Bool:
		_, err := b.WriteString(strconv.FormatBool(rv.Bool()))
		return err
	case reflect.Slice:
		bstr := base64.StdEncoding.EncodeToString(rv.Bytes())
		return writeJsonString(b, bstr)
	case reflect.String:
		return writeJsonString(b, rv.String())
	default:
		if isNil(v) {
			if fd.GetMessageType() != nil {
				bs, err := NewMessage(fd.GetMessageType(), randDefault).MarshalJSONPB(opts)
				if err != nil {
					_, err := b.WriteString("null")
					return err
				}
				_, err = b.Write(bs)
				return err
			}
			_, err := b.WriteString("null")
			return err
		}

		if dm, ok := v.(*Message); ok {
			return dm.marshalJSON(b, opts)
		}

		var err error
		if b.indentCount <= 0 || len(b.indent) == 0 {
			err = opts.Marshal(b, v.(proto.Message))
		} else {
			str, err := opts.MarshalToString(v.(proto.Message))
			if err != nil {
				return err
			}
			indent := strings.Repeat(b.indent, b.indentCount)
			pos := 0
			// add indention prefix to each line
			for pos < len(str) {
				start := pos
				nextPos := strings.Index(str[pos:], "\n")
				if nextPos == -1 {
					nextPos = len(str)
				} else {
					nextPos = pos + nextPos + 1 // include newline
				}
				line := str[start:nextPos]
				if pos > 0 {
					_, err = b.WriteString(indent)
					if err != nil {
						return err
					}
				}
				_, err = b.WriteString(line)
				if err != nil {
					return err
				}
				pos = nextPos
			}
		}
		return err
	}
}

func marshalKnownFieldMapEntryJSON(b *indentBuffer, mk interface{}, vfd *desc.FieldDescriptor, mv interface{},
	opts *jsonpb.Marshaler, randDefault bool) error {
	rk := reflect.ValueOf(mk)
	var strkey string
	switch rk.Kind() {
	case reflect.Bool:
		strkey = strconv.FormatBool(rk.Bool())
	case reflect.Int32, reflect.Int64:
		strkey = strconv.FormatInt(rk.Int(), 10)
	case reflect.Uint32, reflect.Uint64:
		strkey = strconv.FormatUint(rk.Uint(), 10)
	case reflect.String:
		strkey = rk.String()
	default:
		return fmt.Errorf("invalid map key value: %v (%v)", mk, rk.Type())
	}
	err := writeJsonString(b, strkey)
	if err != nil {
		return err
	}
	err = b.sep()
	if err != nil {
		return err
	}
	return marshalKnownFieldValueJSON(b, vfd, mv, opts, randDefault)
}

func writeJsonString(b *indentBuffer, s string) error {
	if sbytes, err := json.Marshal(s); err != nil {
		return err
	} else {
		_, err := b.Write(sbytes)
		return err
	}
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

// +build todo

package ini

import (
	"errors"
	"io"
	"reflect"
)

type Decoder struct {
	p *Parser
}

func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderForParser(NewParser(r))
}

func NewDecoderForParser(p *Parser) *Decoder {
	return &Decoder{
		p: p,
	}
}

func (d *Decoder) Decode(v interface{}) error {
	var val reflect.Value
	if rv, ok := v.(reflect.Value); ok {
		val = rv
	} else {
		val = reflect.ValueOf(v)
	}

	var dec decoderType
	switch t := val.Type(); t.Kind() {
	case reflect.Map:
		dec = &decoderMap{v: val, t: t}
	case reflect.Ptr:
		switch et := t.Elem(); et.Kind() {
		case reflect.Map:
			dec = &decoderMap{v: val.Elem(), t: t}
		//case reflect.Struct:
		//	dec = &decoderStruct{val.Elem(), t}
		default:
			return &DecodeTypeError{t}
		}
	default:
		return &DecodeTypeError{t}
	}

	err := dec.ok()
	if err != nil {
		return err
	}

	section := ""
	for {
		t, err := d.p.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		switch t := t.(type) {
		case *SectionToken:
			section = t.Name
			if !dec.exists(section) {
				return errors.New("Section found but doesn't exist in v: " + section)
			}
		case *SettingToken:
			dec.set(section, t.Left, t.Right)
		}
	}

	return nil
}

type decoderType interface {
	ok() error
	exists(section string) bool
	set(section, key, val string)
}

type decoderMap struct {
	v reflect.Value
	t reflect.Type

	submap bool
}

func (d *decoderMap) ok() error {
	if d.v.Type().Key().Kind() != reflect.String {
		return &DecodeTypeError{d.t}
	}

	switch et := d.v.Type().Elem(); et.Kind() {
	case reflect.String:
	case reflect.Map:
		if et.Key().Kind() != reflect.String {
			return &DecodeTypeError{d.t}
		}

		switch et.Elem().Kind() {
		case reflect.String:
		default:
			return &DecodeTypeError{d.t}
		}

		d.submap = true
	default:
		return &DecodeTypeError{d.t}
	}

	if d.v.IsNil() {
		d.v.Set(reflect.MakeMap(d.v.Type()))
	}

	return nil
}

func (d *decoderMap) exists(section string) bool {
	//v := d.v.MapIndex(reflect.ValueOf(section))

	//return v.IsValid()

	return true
}

func (d *decoderMap) set(section, key, val string) {
	if !d.submap {
		key = section + "/" + key
		if section == "" {
			key = key[1:]
		}

		d.v.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))

		return
	}

	m := d.v.MapIndex(reflect.ValueOf(section))
	if !m.IsValid() {
		d.v.SetMapIndex(reflect.ValueOf(section), reflect.MakeMap(d.v.Type().Elem()))
		m = d.v.MapIndex(reflect.ValueOf(section))
	} else if m.IsNil() {
		m.Set(reflect.MakeMap(m.Type()))
	}

	sub := &decoderMap{v: m}
	sub.set("", key, val)
}

type DecodeTypeError struct {
	Type reflect.Type
}

func (err *DecodeTypeError) Error() string {
	return "Can't decode into type: " + err.Type.String()
}

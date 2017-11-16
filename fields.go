package structFields

import (
	"go/types"
	"reflect"
	"strings"
)

// Field contains the details from a struct field
type Field struct {
	Gen Generator
	Tag string
	V   *types.Var
}

// Name returns the method name
func (f Field) Name() string {
	return f.V.Name()
}

func (f Field) IsExported() bool {
	return f.V.Exported()
}

func (f Field) Field() string {
	if f.Tag == "" {
		return f.Name()
	}
	tag := reflect.StructTag(f.Tag)
	j := tag.Get("json")
	if j == "" {
		return f.Name()
	}
	name := strings.Split(j, ",")[0]
	if name == "" {
		return f.Name()
	}
	if name == "-" {
		name = ""
	}
	return name
}

func (f Field) IsIndex() bool {
	if f.Tag == "" {
		return false
	}
	tag := reflect.StructTag(f.Tag)
	n := tag.Get("neo4j")
	if n == "" {
		return false
	}
	if strings.Split(n, ",")[0] != "index" {
		return false
	}
	return true
}

func (f Field) IsArrayOrSlice() bool {
	if _, ok := f.V.Type().(*types.Slice); ok {
		return true
	}
	_, ok := f.V.Type().(*types.Array)
	return ok
}

func (f Field) IsSlice() bool {
	_, ok := f.V.Type().(*types.Slice)
	return ok
}

func (f Field) IsPointer() bool {
	_, ok := f.V.Type().(*types.Pointer)
	return ok
}

func (f Field) IsStruct() bool {
	_, ok := f.V.Type().(*types.Named)
	return ok
}

func (f Field) Type() string {
	return types.TypeString(f.V.Type(), f.Gen.Qf)
}

func (f Field) UnderlyingTypeName() string {
	return types.TypeString(f.UnderlyingType(), f.Gen.Qf)
}

func (f Field) UnderlyingTarget() fieldser {
	var t types.Type
	switch v := f.V.Type().(type) {
	case elemer:
		t = v.Elem()
	case *types.Named:
		t = v
	}
	if _, ok := t.(underlyinger); !ok {
		return nil
	}
	u := t.(underlyinger).Underlying()
	switch t := u.(type) {
	case *types.Struct:
		return fields{
			g:      f.Gen,
			target: t,
		}
	case *types.Pointer:
		return fields{
			g:      f.Gen,
			target: t.Elem().(*types.Named).Underlying().(*types.Struct),
		}
	}
	return nil
}

type fields struct {
	g      Generator
	target *types.Struct
}

func (f fields) Fields() []Field {
	numFields := f.target.NumFields()
	fields := make([]Field, 0)
	for i := 0; i < numFields; i++ {
		f := Field{f.g, f.target.Tag(i), f.target.Field(i)}
		if f.Field() != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

var fl *types.Basic

func init() {
	for _, t := range types.Typ {
		if t.Kind() == types.Float64 {
			fl = t
			break
		}
	}
}

func (f Field) ConvertibleFromFloat64() bool {
	return types.ConvertibleTo(fl, f.V.Type())
}

func (f Field) UnderlyingConvertibleFromFloat64() bool {
	return types.ConvertibleTo(fl, f.UnderlyingType())
}

func (f Field) IsAnonymous() bool {
	return f.V.Anonymous()
}

type fieldser interface {
	Fields() []Field
}

func (f Field) UnderlyingIsBasic() bool {
	switch t := f.V.Type().(type) {
	case elemer:
		_, basic := t.Elem().(*types.Basic)
		return basic
	}
	_, basic := f.V.Type().(*types.Basic)
	return basic
}

func (f Field) UnderlyingIsPointer() bool {
	switch t := f.V.Type().(type) {
	case elemer:
		_, basic := t.Elem().(*types.Pointer)
		return basic
	}
	_, basic := f.V.Type().(*types.Pointer)
	return basic
}

func (f Field) UnderlyingType() types.Type {
	switch t := f.V.Type().(type) {
	case *types.Array:
		switch t := t.Elem().(type) {
		case *types.Pointer:
			return t.Elem()
		}
		return t.Elem()
	case *types.Slice:
		switch t := t.Elem().(type) {
		case *types.Pointer:
			return t.Elem()
		}
		return t.Elem()
	case elemer:
		return t.Elem()
	}
	return nil
}

type elemer interface {
	Elem() types.Type
}

type underlyinger interface {
	Underlying() types.Type
}

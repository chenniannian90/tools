package envconf

import (
	"go/ast"
	"reflect"

	encodingx "github.com/go-courier/x/encoding"
	reflectx "github.com/go-courier/x/reflect"
)

func DecodeEnvVars(envVars *EnvVars, v interface{}) error {
	walker := NewPathWalker()
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	return scanAndSetValue(envVars, walker, rv)
}

func scanAndSetValue(envVars *EnvVars, walker *PathWalker, rv reflect.Value) error {
	kind := rv.Kind()

	if kind != reflect.Ptr && rv.CanAddr() {
		if defaultsSetter, ok := rv.Addr().Interface().(interface{ SetDefaults() }); ok {
			defaultsSetter.SetDefaults()
		}
	}

	switch kind {
	case reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflectx.New(rv.Type()))
		}
		return scanAndSetValue(envVars, walker, rv.Elem())
	case reflect.Func, reflect.Interface, reflect.Chan, reflect.Map:
		// skip
	default:
		typ := rv.Type()
		if typ.Implements(interfaceTextUnmarshaller) || reflect.PtrTo(typ).Implements(interfaceTextUnmarshaller) {
			v := envVars.Get(walker.String())
			if v != nil {
				if err := encodingx.UnmarshalText(rv, []byte(v.Value)); err != nil {
					return err
				}
			}
			return nil
		}

		switch kind {
		case reflect.Array, reflect.Slice:
			n := envVars.Len(walker.String())

			if kind == reflect.Slice && rv.IsNil() {
				rv.Set(reflect.MakeSlice(rv.Type(), n, n))
			}

			for i := 0; i < rv.Len(); i++ {
				walker.Enter(i)
				if err := scanAndSetValue(envVars, walker, rv.Index(i)); err != nil {
					return err
				}
				walker.Exit()
			}

		case reflect.Struct:
			tpe := rv.Type()
			for i := 0; i < rv.NumField(); i++ {
				field := tpe.Field(i)

				flags := (map[string]bool)(nil)
				name := field.Name

				if !ast.IsExported(name) {
					continue
				}

				if tag, ok := field.Tag.Lookup("env"); ok {
					n, fs := tagValueAndFlags(tag)
					if n == "-" {
						continue
					}
					if n != "" {
						name = n
					}
					flags = fs
				}

				inline := flags == nil && reflectx.Deref(field.Type).Kind() == reflect.Struct && field.Anonymous

				if !inline {
					walker.Enter(name)
				}

				if err := scanAndSetValue(envVars, walker, rv.Field(i)); err != nil {
					return err
				}

				if !inline {
					walker.Exit()
				}
			}
		default:
			v := envVars.Get(walker.String())
			if v != nil {
				if err := encodingx.UnmarshalText(rv, []byte(v.Value)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

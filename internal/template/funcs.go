package template

import (
	"fmt"
	"reflect"
)

func tmpl_int(i interface{}) (int64, error) {
	iv := reflect.ValueOf(i)
	
	switch iv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return iv.Int(), nil
	}

	return 0, fmt.Errorf("unknown type for %q (%T)", iv, i)	
}

func tmpl_float(i interface{}) (float64, error) {
	iv := reflect.ValueOf(i)
	
	switch iv.Kind() {
		case reflect.Float32, reflect.Float64:
			return iv.Float(), nil	
	}

	return 0, fmt.Errorf("unknown type for %q (%T)", iv, i)
}

func tmpl_add(b, a interface{}) (float64, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
		case reflect.Int:
			switch bv.Kind() {
				case reflect.Int:
					return float64(av.Int() + bv.Int()), nil
				case reflect.Float64:
					return float64(av.Int()) + bv.Float(), nil
				default:
					return 0, fmt.Errorf("unknown type for %q (%T)", bv, b)
			}
		case reflect.Float64:
			switch bv.Kind() {
				case reflect.Int:
					return av.Float() + float64(bv.Int()), nil
				case reflect.Float64:
					return av.Float() + bv.Float(), nil
				default:
					return 0, fmt.Errorf("unknown type for %q (%T)", bv, b)
			}
		default:
			return 0, fmt.Errorf("unknown type for %q (%T)", av, a)
	}
}
package main

import (
	"bytes"
	"fmt"
	"gonum.org/v1/plot/plotter"
	"image/color"
)

// Json helper has a list of functions to help us manipulate json's results and perform
// some simple schema checkings
func JsonGetString(v Value) (string, error) {
	if v.Type != kValueTypeString {
		return "", fmt.Errorf("value is not type string but type %s", v.Type.GetName())
	}

	return v.String, nil
}

func JsonGetNumber(v Value) (float64, error) {
	if v.Type != kValueTypeNumber {
		return 0, fmt.Errorf("value is not type number but type %s", v.Type.GetName())
	}
	return v.Number, nil
}

func JsonGetBoolean(v Value) (bool, error) {
	if v.Type != kValueTypeBoolean {
		return false, fmt.Errorf("value is not type boolean but type %s", v.Type.GetName())
	}
	return v.Boolean, nil
}

func JsonGetNull(v Value) error {
	if v.Type != kValueTypeNull {
		return fmt.Errorf("value is not type null but type %s", v.Type.GetName())
	}
	return nil
}

func JsonListGet(v Value, idx int) (Value, error) {
	if v.Type != kValueTypeList {
		return NewNull(), fmt.Errorf("value is not type list but type %s", v.Type.GetName())
	}

	if idx >= len(v.List.Value) {
		return NewNull(), fmt.Errorf("index out of range , index is %d, size is %d", idx, len(v.List.Value))
	}

	return v.List.Value[idx], nil
}

func JsonObjectGet(v Value, key string) (Value, error) {
	if v.Type != kValueTypeObject {
		return NewNull(), fmt.Errorf("value is not type object but type %s", v.Type.GetName())
	}

	if val, err := v.Object.Value[key]; !err {
		return NewNull(), fmt.Errorf("key %s doesn't exist", key)
	} else {
		return val, nil
	}
}

func JsonObjectGetMultipleKey(v Value, keys ...string) (Value, error) {
	if v.Type != kValueTypeObject {
		return NewNull(), fmt.Errorf("value is not type object but type %s", v.Type.GetName())
	}

	for _, k := range keys {
		if val, err := v.Object.Value[k]; err {
			return val, nil
		}
	}

	// generate the keylist
	keyList := bytes.Buffer{}
	for _, k := range keys {
		keyList.WriteString(k)
		keyList.WriteString(",")
	}

	return NewNull(), fmt.Errorf("key list :%s doesn't exist in object", keyList.String())
}

func jsonGetColorComponent(v Value, k1 string, k2 string) (uint8, error) {
	var c float64
	name := k2

	if cval, err := JsonObjectGetMultipleKey(v, k1, k2); err != nil {
		return 0, err
	} else {
		if dc, err := JsonGetNumber(cval); err != nil {
			return 0, fmt.Errorf("component %s failed, %v", name, err)
		} else {
			c = dc
		}
	}

	ic := int(c)
	if ic < 0 || ic > 255 {
		return 0, fmt.Errorf("component %s is not a valid color RGB value, the value is %d", name, ic)
	}

	return uint8(ic), nil
}

// Turn a json object into Color
func JsonObjectToColor(v Value) (color.Color, error) {
	if v.Type != kValueTypeObject {
		return nil, fmt.Errorf("value is not type object but type %s", v.Type.GetName())
	}

	var r, g, b, a uint8
	var err error

	if r, err = jsonGetColorComponent(v, "r", "R"); err != nil {
		return nil, err
	}

	if g, err = jsonGetColorComponent(v, "g", "G"); err != nil {
		return nil, err
	}

	if b, err = jsonGetColorComponent(v, "b", "B"); err != nil {
		return nil, err
	}

	if a, err = jsonGetColorComponent(v, "a", "A"); err != nil {
		return nil, err
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}

// Plotter related Json conversion
func JsonObjectToPoint(v Value) (float64, float64, error) {
	if v.Type != kValueTypeObject {
		return 0, 0, fmt.Errorf("value is not type object but type %s", v.Type.GetName())
	}
	var x float64
	var y float64

	if xval, err := JsonObjectGetMultipleKey(v, "x", "X"); err != nil {
		return 0, 0, err
	} else {
		if dx, err := JsonGetNumber(xval); err != nil {
			return 0, 0, fmt.Errorf("component X failed, %v", err)
		} else {
			x = dx
		}
	}

	if yval, err := JsonObjectGetMultipleKey(v, "y", "Y"); err != nil {
		return 0, 0, err
	} else {
		if dy, err := JsonGetNumber(yval); err != nil {
			return 0, 0, fmt.Errorf("component Y failed, %v", err)
		} else {
			y = dy
		}
	}

	return x, y, nil
}

func JsonListToPointList(v Value) (*plotter.XYs, error) {
	if v.Type != kValueTypeList {
		return nil, fmt.Errorf("value is not type list but type %s", v.Type.GetName())
	}

	if len(v.List.Value) == 0 {
		ret := make(plotter.XYs, 0)
		return &ret, nil
	} else {
		var ret *plotter.XYs
		// decide whether its a list of raw points or a list of objects
		if v.List.Value[0].Type == kValueTypeObject {
			pts := make(plotter.XYs, len(v.List.Value))
			for idx, element := range v.List.Value {
				if x, y, err := JsonObjectToPoint(element); err != nil {
					return nil, fmt.Errorf("index %d failed to parse as point due to reason %v", idx, err)
				} else {
					pts[idx].X = x
					pts[idx].Y = y
				}
			}
			ret = &pts
		} else if v.List.Value[0].Type == kValueTypeNumber {
			sz := len(v.List.Value)
			if sz%2 != 0 {
				sz = sz - 1
			}

			pts := make(plotter.XYs, sz/2)

			idx := 0
			for i := 0; i < sz; i += 2 {
				x := v.List.Value[i]
				y := v.List.Value[i+1]

				if x.Type != kValueTypeNumber || y.Type != kValueTypeNumber {
					return nil, fmt.Errorf("index %d, failed to parse 2 consecutive number", i)
				}

				pts[idx].X = x.Number
				pts[idx].Y = y.Number
				idx++
			}
			ret = &pts
		}
		return ret, nil
	}
}

func JsonListToVector(v Value) (*plotter.Values, error) {
	var ret plotter.Values
	if v.Type != kValueTypeList {
		return nil, fmt.Errorf("value is not type list but type %s", v.Type.GetName())
	}

	for idx, ele := range v.List.Value {
		if val, err := JsonGetNumber(ele); err != nil {
			return nil, fmt.Errorf("index %d failed to parse as number due to reason %v", idx, err)
		} else {
			ret = append(ret, val)
		}
	}

	return &ret, nil
}

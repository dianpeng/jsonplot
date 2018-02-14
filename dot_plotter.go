package main

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type dotPlotter struct{}

func (d *dotPlotter) GetName() string {
	return "dot-plotter"
}

func (d *dotPlotter) Plot(path string, data Value) error {
	title := "dot-plot"
	xlabel := "X"
	ylabel := "Y"
	grids := false
	size := 4.0

	// get the title
	if v, err := JsonObjectGetMultipleKey(data, "Title", "title"); err == nil {
		if val, err := JsonGetString(v); err == nil {
			title = val
		}
	}

	// get the X
	if v, err := JsonObjectGetMultipleKey(data, "X", "x"); err == nil {
		if val, err := JsonGetString(v); err == nil {
			xlabel = val
		}
	}

	// get the Y
	if v, err := JsonObjectGetMultipleKey(data, "Y", "y"); err == nil {
		if val, err := JsonGetString(v); err == nil {
			ylabel = val
		}
	}

	// get the grids
	if v, err := JsonObjectGetMultipleKey(data, "Grids", "grids"); err == nil {
		if val, err := JsonGetBoolean(v); err == nil {
			grids = val
		}
	}

	// get the size
	if v, err := JsonObjectGetMultipleKey(data, "size", "Size"); err == nil {
		if val, err := JsonGetNumber(v); err == nil {
			size = val
		}
	}

	// set up the plotter
	p, err := plot.New()
	if err != nil {
		return fmt.Errorf("cannot create plotter %v", err)
	}

	p.Title.Text = title
	p.X.Label.Text = xlabel
	p.Y.Label.Text = ylabel

	if grids {
		p.Add(plotter.NewGrid())
	}

	// a list of interface used to feed AddLinePoints
	arg := []interface{}{}

	// get the data from it
	if v, err := JsonObjectGetMultipleKey(data, "Data", "data"); err == nil {
		if v.Type != kValueTypeObject {
			return fmt.Errorf("\"data\" field must be an object but got type %s,%s", v.Type.GetName())
		}

		// go through each key value pair in the data list and render them
		for key, val := range v.Object.Value {
			if pts, err := JsonListToPointList(val); err != nil {
				return fmt.Errorf("dot-plotter \"data\" field \"%s\" cannot convert "+
					"to a list of points for reason %v", key, err)
			} else {
				arg = append(arg, key)
				arg = append(arg, *pts)
			}
		}
	}

	plotutil.AddLinePoints(p, arg...)

	sz := vg.Length(size)
	// save it to png file
	if err = p.Save(sz*vg.Inch, sz*vg.Inch, path); err != nil {
		return fmt.Errorf("dot-plotter cannot save file to %s, due to reason %v", path, err)
	}

	return nil
}

func init() {
	PlotterFactory["dot-plotter"] = &dotPlotter{}
}

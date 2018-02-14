package main

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type histPlotter struct{}

func (h *histPlotter) GetName() string { return "hist-plotter" }

func (h *histPlotter) Plot(path string, data Value) error {
	title := "plot"
	xlabel := "X"
	ylabel := "Y"
	size := 4.0
	grids := false
	bins := 8

	if v, err := JsonObjectGetMultipleKey(data, "Title", "title"); err == nil {
		if val, err := JsonGetString(v); err == nil {
			title = val
		}
	}

	if v, err := JsonObjectGetMultipleKey(data, "Y", "y"); err == nil {
		if val, err := JsonGetString(v); err == nil {
			ylabel = val
		}
	}

	if v, err := JsonObjectGetMultipleKey(data, "X", "x"); err == nil {
		if val, err := JsonGetString(v); err == nil {
			xlabel = val
		}
	}

	if v, err := JsonObjectGetMultipleKey(data, "Grids", "grids"); err == nil {
		if val, err := JsonGetBoolean(v); err == nil {
			grids = val
		}
	}

	if v, err := JsonObjectGetMultipleKey(data, "Size", "size"); err == nil {
		if val, err := JsonGetNumber(v); err == nil {
			size = val
		}
	}

	if v, err := JsonObjectGetMultipleKey(data, "Bins", "bins"); err == nil {
		if val, err := JsonGetNumber(v); err == nil {
			bins = int(val)
		}
	}

	p, err := plot.New()
	if err != nil {
		return fmt.Errorf("\"hist-plotter\" cannot create plot due to reason %v", err)
	}
	p.Title.Text = title
	p.X.Label.Text = xlabel
	p.Y.Label.Text = ylabel
	if grids {
		p.Add(plotter.NewGrid())
	}

	var vals *plotter.Values

	if v, err := JsonObjectGetMultipleKey(data, "Data", "data"); err != nil {
		return fmt.Errorf("\"hist-plotter\" cannot get \"Data\" field due to reason %v", err)
	} else {
		if val, err := JsonListToVector(v); err != nil {
			return fmt.Errorf("\"hist-plotter\"'s \"Data\" field must be a list of numbers")
		} else {
			vals = val
		}
	}

	hist, err := plotter.NewHist(*vals, bins)
	if err != nil {
		return fmt.Errorf("\"hist-plotter\" cannot create histgram object due to reason %v", err)
	}

	hist.Normalize(1)
	p.Add(hist)

	sz := vg.Length(size)
	if err := p.Save(sz*vg.Inch, sz*vg.Inch, path); err != nil {
		return fmt.Errorf("\"hist-plotter\" cannot save fiel to path %s due to reason %v", path, err)
	}
	return nil
}

func init() {
	PlotterFactory["hist-plotter"] = &histPlotter{}
}

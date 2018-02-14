package main

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type barPlotter struct{}

func (b *barPlotter) GetName() string { return "bar-plotter" }

func (bar *barPlotter) Plot(path string, data Value) error {
	title := "Plot"
	ylabel := "Heights"
	size := 4.0
	width := 4.0
	grids := false

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

	if v, err := JsonObjectGetMultipleKey(data, "Width", "width"); err == nil {
		if val, err := JsonGetNumber(v); err == nil {
			width = val
		}
	}

	p, err := plot.New()
	if grids {
		p.Add(plotter.NewGrid())
	}
	if err != nil {
		return fmt.Errorf("\"bar-plotter\" cannot create plot due to reason %v", err)
	}
	p.Title.Text = title
	p.Y.Label.Text = ylabel

	// get each groups
	grp, err := JsonObjectGetMultipleKey(data, "Group", "group")

	if err != nil {
		return fmt.Errorf("\"bar-plotter\" data doesn't have \"group\" field")
	}

	if grp.Type != kValueTypeObject {
		return fmt.Errorf("\"bar-plotter\" data field \"group\" must be an object, but got type %s", grp.Type.GetName())
	}

	sz := vg.Length(size)
	xlabel := []string{}
	offset := make([]float64, len(grp.Object.Value))
	bars := make([]plot.Plotter, len(grp.Object.Value))
	nums := []*plotter.Values{}

	// get all the Values from the input data and figure out the maxNum which is how many row will
	// be showned up in the final generated graph/png
	maxNum := 0
	for k, v := range grp.Object.Value {
		if v.Type != kValueTypeObject {
			return fmt.Errorf("\"bar-plotter\" group's entry must be an object, but got type %s", v.Type.GetName())
		}

		xlabel = append(xlabel, k)

		if v, err := JsonObjectGetMultipleKey(v, "Data", "data"); err != nil {
			return fmt.Errorf("\"bar-plotter\" each group must have a \"data\" field")
		} else {
			if val, err := JsonListToVector(v); err != nil {
				return fmt.Errorf("\"bar-plotter\" each group's field \"Data\" must be a list of numbers")
			} else {
				nums = append(nums, val)
				if maxNum < len(*val) {
					maxNum = len(*val)
				}
			}
		}

	}

	// do a simple layout recalculation
	sizeOfOutput := float64(sz * vg.Inch * vg.Length(0.8))
	{

		needSize := float64(len(grp.Object.Value)*maxNum) * width
		realSize := needSize

		if needSize > sizeOfOutput {
			realSize = sizeOfOutput
		}

		// width of each bar
		width := float64(realSize) / float64(maxNum*len(grp.Object.Value))

		// start of the bar
		start := width * -(float64(len(grp.Object.Value)/2 - 1))

		for i := 0; i < len(grp.Object.Value); i++ {
			offset[i] = start
			start += width
		}
	}

	for idx, num := range nums {
		bar, err := plotter.NewBarChart(*num, vg.Length(width))
		if err != nil {
			return fmt.Errorf("\"bar-plotter\" cannot create bar with error %v", err)
		}

		bar.LineStyle.Width = vg.Length(0)
		bar.Color = plotutil.Color(idx)
		bar.Offset = vg.Points(offset[idx])
		bars[idx] = bar
		idx++
	}

	p.Add(bars...)
	{
		idx := 0
		for _, x := range xlabel {
			p.Legend.Add(x, bars[idx].(*plotter.BarChart))
			idx++
		}

		// tihs is the best bet for where the legend should show up
		p.Legend.Top = true
		p.Legend.Left = true
	}
	// generate X label name
	{
		labels := []string{}
		for i := 0; i < maxNum; i++ {
			labels = append(labels, fmt.Sprintf("%d", i))
		}
		p.NominalX(labels...)
	}

	if err := p.Save(vg.Inch*sz, vg.Inch*sz, path); err != nil {
		return fmt.Errorf("\"bar-plotter\" cannot save file to path %s with error %v",
			path, err)
	}

	return nil
}

func init() {
	PlotterFactory["bar-plotter"] = &barPlotter{}
}

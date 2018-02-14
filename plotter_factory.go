package main

var PlotterFactory map[string]Plotter = make(map[string]Plotter)

func NewPlotter(name string) Plotter {
	if v, err := PlotterFactory[name]; !err {
		return nil
	} else {
		return v
	}
}

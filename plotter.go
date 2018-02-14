package main

// Plotter is a interface that describe the type of underlying plot implementation
// It accepts a string represent for saved file path and another [string]Value object
// represents the input data
type Plotter interface {

	// Plot input data described by [string]Value and save the plotted result into
	// a file specified by the 1st argument
	Plot(string, Value) error

	// Get the name of this plotter
	GetName() string
}

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var input = flag.String("input", "-", "the input json for plotting, default to read from stdin")

func getInput() ([]byte, error) {
	if *input == "-" {
		return ioutil.ReadAll(os.Stdin)
	} else {
		return ioutil.ReadFile(*input)
	}
}

func doSinglePlot(index int, jdom Value) error {
	var plotter Plotter
	var path string

	if t, err := JsonObjectGetMultipleKey(jdom, "Type", "type"); err != nil {
		return fmt.Errorf("index %d,%v", index, err)
	} else {
		if name, err := JsonGetString(t); err != nil {
			return fmt.Errorf("index %d,\"Type\" field is not a string", index)
		} else {
			plotter = NewPlotter(name)
			if plotter == nil {
				return fmt.Errorf("index %d,plotter %s doesn't support", index, name)
			}
		}
	}

	if t, err := JsonObjectGetMultipleKey(jdom, "Path", "path"); err != nil {
		return fmt.Errorf("index %d,%v", index, err)
	} else {
		if name, err := JsonGetString(t); err != nil {
			return fmt.Errorf("index %d,\"Path\" field is not a string", index)
		} else {
			path = name
		}
	}

	if d, err := JsonObjectGetMultipleKey(jdom, "Config", "config"); err != nil {
		return fmt.Errorf("index %d,%v", index, err)
	} else {
		return plotter.Plot(path, d)
	}
}

func doPlot(data string) error {
	jdom, err := NewJsonParser(data).Parse()
	if err != nil {
		return err
	}

	if jdom.Type != kValueTypeObject && jdom.Type != kValueTypeList {
		return fmt.Errorf("the root element of input json *MUST* be an object")
	}

	succ := 0
	jobs := []Value{}

	if jdom.Type == kValueTypeList {
		jobs = jdom.List.Value
	} else {
		jobs = append(jobs, jdom)
	}

	for idx, x := range jdom.List.Value {
		if err := doSinglePlot(idx, x); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			succ++
			fmt.Fprintf(os.Stdout, "index %d plot succeeded\n", idx)
		}
	}
	fmt.Fprintf(os.Stdout, "Total Job %d; Successful %d; Failed %d\n", len(jobs), succ, len(jobs)-succ)
	return nil
}

func main() {
	flag.Parse()
	data, err := getInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read input specified as %s with error %v", *input, err)
		os.Exit(1)
	}
	if err := doPlot(string(data)); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

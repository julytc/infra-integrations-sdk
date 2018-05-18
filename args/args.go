package args

import (
	"github.com/namsral/flag"
)

// Default integration arguments. Loaded via cli or environment variables.
var (
	Verbose   bool
	Pretty    bool
	All       bool
	Metrics   bool
	Inventory bool
	Events    bool
)

// LoadDefaultArgs loads default arguments from cli or environment variables.
func LoadDefaultArgs() {
	flag.BoolVar(&Verbose, "verbose", false, "Print more information to logs")
	flag.BoolVar(&Pretty, "pretty", false, "Print pretty formatted JSON")
	flag.BoolVar(&All, "all", false, "Publish all kind of data (metrics, inventory, events)")
	flag.BoolVar(&Metrics, "metrics", false, "Publish metrics data")
	flag.BoolVar(&Inventory, "inventory", false, "Publish inventory data")
	flag.BoolVar(&Events, "events", false, "Publish events data")

	if !Metrics && !Inventory && !Events {
		All = true
	}
}

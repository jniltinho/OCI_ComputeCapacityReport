package config

const Version = "4.0.0"

type Options struct {
	UserAuth       string
	ConfigFilePath string
	ConfigProfile  string
	SuperUser      bool
	Compartment    string
	TargetRegion   string
	Shape          string
	OCPUs          int
	Memory         int
	DRCC           bool
}
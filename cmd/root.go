package cmd

import (
	"os"

	"github.com/Olygo/OCI_ComputeCapacityReport/internal/app"
	"github.com/Olygo/OCI_ComputeCapacityReport/internal/config"
	"github.com/spf13/cobra"
)

var opts config.Options

var rootCmd = &cobra.Command{
	Use:   "oci-compute-capacity-report",
	Short: "Check the availability of any compute shape across OCI regions",
	Long: `OCI_ComputeCapacityReport provides the availability status down to the Fault Domain level
and automatically relaunches after completing the first query or encountering an error.

Output meanings:
  AVAILABLE              => The capacity for the specified shape is currently available.
  HARDWARE_NOT_SUPPORTED => The necessary hardware has not yet been deployed in this region.
  OUT_OF_HOST_CAPACITY   => Additional hardware is currently being deployed in this region.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(opts)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	flags := rootCmd.Flags()

	flags.StringVar(&opts.UserAuth, "auth", "", "Force an authentication method: cs (cloudshell), cf (config file), ip (instance principals)")
	flags.StringVar(&opts.ConfigFilePath, "config-file", "~/.oci/config", "Path to your OCI config file")
	flags.StringVar(&opts.ConfigProfile, "profile", "DEFAULT", "Config file section to use")
	flags.BoolVar(&opts.SuperUser, "su", false, "Notify the tool that you have tenancy-level admin rights to prevent prompting")
	flags.StringVar(&opts.Compartment, "comp", "", "Filter on a compartment when you do not have Admin rights at the tenancy level")
	flags.StringVar(&opts.TargetRegion, "region", "", `Region name to analyze, e.g. "eu-frankfurt-1" or "all_regions" (default: home region)`)
	flags.StringVar(&opts.Shape, "shape", "", "Compute shape name you want to analyze")
	flags.IntVar(&opts.OCPUs, "ocpus", 0, "Specify a particular amount of oCPU")
	flags.IntVar(&opts.Memory, "memory", 0, "Specify a particular amount of memory (GB)")
	flags.BoolVar(&opts.DRCC, "drcc", false, "Display 'available_count' value for DRCC customers and whitelisted tenancies")
}
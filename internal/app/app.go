package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Olygo/OCI_ComputeCapacityReport/internal/auth"
	"github.com/Olygo/OCI_ComputeCapacityReport/internal/capacity"
	"github.com/Olygo/OCI_ComputeCapacityReport/internal/config"
	"github.com/Olygo/OCI_ComputeCapacityReport/internal/identity"
	"github.com/Olygo/OCI_ComputeCapacityReport/internal/output"
	"github.com/oracle/oci-go-sdk/v65/common"
	ociidentity "github.com/oracle/oci-go-sdk/v65/identity"
)

func Run(opts config.Options) error {
	output.Clear()

	configPath := expandPath(opts.ConfigFilePath)

	authResult, err := auth.Init(opts.UserAuth, configPath, opts.ConfigProfile)
	if err != nil {
		return err
	}

	output.Clear()

	homeRegionKey := ""
	if authResult.Tenancy.HomeRegionKey != nil {
		homeRegionKey = *authResult.Tenancy.HomeRegionKey
	}

	fmt.Println(output.Green(fmt.Sprintf("\n%s", strings.Repeat("*", 94))))
	output.PrintInfo(output.Green, "Script", "started", "oci-compute-capacity-report")
	output.PrintInfo(output.Green, "Script", "version", config.Version)
	output.PrintInfo(output.Green, "Login", "success", authResult.AuthName)
	output.PrintInfo(output.Green, "Login", "profile", authResult.Details)
	output.PrintInfo(output.Green, "Tenancy", *authResult.Tenancy.Name, fmt.Sprintf("home region: %s", homeRegionKey))

	identityClient, err := identity.NewClient(authResult.Provider)
	if err != nil {
		return err
	}

	tenancyID, err := authResult.Provider.TenancyOCID()
	if err != nil {
		return err
	}

	regionsToAnalyze, err := identity.GetRegionSubscriptionList(identityClient, tenancyID, opts.TargetRegion)
	if err != nil {
		return err
	}

	regionsValidated, err := identity.ValidateRegionConnectivity(regionsToAnalyze, authResult.Provider, tenancyID)
	if err != nil {
		return err
	}

	homeRegion, err := identity.GetHomeRegion(identityClient, tenancyID)
	if err != nil {
		return err
	}

	if opts.Shape != "" {
		output.PrintInfo(output.Green, "Shape", "analyzed", opts.Shape)
	}
	if opts.OCPUs > 0 {
		output.PrintInfo(output.Green, "oCPUs", "amount", fmt.Sprintf("%d cores", opts.OCPUs))
	}
	if opts.Memory > 0 {
		output.PrintInfo(output.Green, "Memory", "amount", fmt.Sprintf("%d gbs", opts.Memory))
	}

	fmt.Println(output.Green(strings.Repeat("*", 94) + "\n"))

	userCompartment, err := identity.SetUserCompartment(identityClient, opts, tenancyID)
	if err != nil {
		if strings.Contains(err.Error(), "quitting") {
			fmt.Println()
			return nil
		}
		return err
	}

	userShapeName := opts.Shape
	userShapeOCPUs := float64(opts.OCPUs)
	userShapeMemory := float64(opts.Memory)
	firstShapePrompt := true

	homeRegionName := *homeRegion.RegionName

	for {
		err := runAnalysis(
			regionsValidated, authResult.Provider, tenancyID, userCompartment,
			&userShapeName, &userShapeOCPUs, &userShapeMemory,
			homeRegionName, opts, &firstShapePrompt,
		)

		var restart *capacity.RestartFlowError
		if errors.As(err, &restart) {
			continue
		}

		if err != nil {
			if strings.Contains(err.Error(), "quitting") {
				fmt.Println()
				return nil
			}
			if err.Error() == "authorization failed" {
				return err
			}
			return err
		}

		userShapeName = ""
		userShapeOCPUs = 0
		userShapeMemory = 0
	}
}

func runAnalysis(
	regions []ociidentity.RegionSubscription,
	provider common.ConfigurationProvider,
	tenancyID, compartmentID string,
	shapeName *string,
	ocpus, memory *float64,
	homeRegionName string,
	opts config.Options,
	firstShapePrompt *bool,
) error {
	if *shapeName == "" {
		if *firstShapePrompt {
			if _, err := capacity.PrintShapeList(homeRegionName, provider, compartmentID); err != nil {
				return err
			}
		}

		name, err := capacity.SetUserShapeName(homeRegionName, provider, compartmentID, firstShapePrompt)
		if err != nil {
			return err
		}
		*shapeName = name
	}

	if _, ok := capacity.DenseIOFlexShapes[*shapeName]; ok {
		val, err := capacity.SetDenseIOShapeOCPUs(*shapeName)
		if err != nil {
			return err
		}
		*ocpus = float64(val)
		fmt.Println()
	} else if !strings.Contains(*shapeName, ".Flex") || strings.HasPrefix(*shapeName, "BM.") {
		*ocpus = 0
		*memory = 0
	} else {
		if *ocpus == 0 {
			val, err := capacity.SetUserShapeOCPUs(*shapeName)
			if err != nil {
				return err
			}
			*ocpus = float64(val)
		}
		if *memory == 0 {
			val, err := capacity.SetUserShapeMemory(*shapeName)
			if err != nil {
				return err
			}
			*memory = float64(val)
		}
	}

	header := fmt.Sprintf("\n%-20s %-30s %-20s %-25s %-10s %-10s", "REGION", "AVAILABILITY_DOMAIN", "FAULT_DOMAIN", "SHAPE", "OCPU", "MEMORY")
	if opts.DRCC {
		header += fmt.Sprintf(" %-16s", "AVAILABLE_COUNT")
	}
	header += " AVAILABILITY\n"
	fmt.Print(header)

	for _, region := range regions {
		if err := capacity.ProcessRegion(
			*region.RegionName,
			provider,
			tenancyID,
			compartmentID,
			*shapeName,
			*ocpus,
			*memory,
			opts.DRCC,
		); err != nil {
			return err
		}
	}

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
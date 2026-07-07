package capacity

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Olygo/OCI_ComputeCapacityReport/internal/identity"
	"github.com/Olygo/OCI_ComputeCapacityReport/internal/output"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	ociidentity "github.com/oracle/oci-go-sdk/v65/identity"
)

type RestartFlowError struct{}

func (e *RestartFlowError) Error() string { return "restart flow" }

type ShapeConfig struct {
	OCPUs       float64
	Memory      float64
	IsFlex      bool
	ShapeInfo   *core.Shape
}

func PrintShapeList(homeRegion string, provider common.ConfigurationProvider, compartmentID string) ([]string, error) {
	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}
	client.SetRegion(homeRegion)

	shapes, err := client.ListShapes(context.Background(), core.ListShapesRequest{
		CompartmentId: &compartmentID,
	})
	if err != nil {
		output.PrintError([]string{err.Error()}, "ERROR")
		return nil, err
	}

	known := make(map[string]bool)
	allShapes := make([]string, len(KnownShapes))
	copy(allShapes, KnownShapes)
	for _, s := range allShapes {
		known[s] = true
	}

	for _, shape := range shapes.Items {
		if shape.Shape != nil && !known[*shape.Shape] {
			allShapes = append(allShapes, *shape.Shape)
			known[*shape.Shape] = true
		}
	}
	sort.Strings(allShapes)

	fmt.Println(output.Yellow("\nGet all available shapes at: https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm\n"))

	maxLen := 0
	for _, s := range allShapes {
		if l := len(s); l > maxLen {
			maxLen = l
		}
	}
	colWidth := maxLen + 4

	for i := 0; i < len(allShapes); i += 6 {
		end := i + 6
		if end > len(allShapes) {
			end = len(allShapes)
		}
		var parts []string
		for _, shape := range allShapes[i:end] {
			parts = append(parts, fmt.Sprintf("%-*s", colWidth, shape))
		}
		fmt.Println(strings.Join(parts, ""))
	}

	return allShapes, nil
}

func SetUserShapeName(homeRegion string, provider common.ConfigurationProvider, compartmentID string, firstRun *bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	if *firstRun {
		*firstRun = false
		fmt.Print(output.Yellow("\nEnter the name of a shape to discover its regional capacity or [Q]uit: "))
	} else {
		fmt.Print(output.Yellow("\nEnter another shape name, [P]rint shapes list or [Q]uit: "))
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return "", &RestartFlowError{}
	}

	upper := strings.ToUpper(input)
	switch upper {
	case "Q", "QUIT":
		return "", fmt.Errorf("quitting the program as per user request")
	case "P", "PRINT":
		if _, err := PrintShapeList(homeRegion, provider, compartmentID); err != nil {
			return "", err
		}
		fmt.Print(output.Yellow("\nEnter a shape name or [Q]uit: "))
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		upper = strings.ToUpper(input)
		if upper == "Q" || upper == "QUIT" {
			return "", fmt.Errorf("quitting the program as per user request")
		}
		if input == "" {
			return "", &RestartFlowError{}
		}
	}

	return input, nil
}

func SetUserShapeOCPUs(shapeName string) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(output.Yellow(fmt.Sprintf("\nAmount of OCPUs needed for shape %s: ", shapeName)))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		val, err := strconv.Atoi(input)
		if err != nil || val == 0 {
			fmt.Println(output.Red("\nInvalid input. Please enter a valid value (integer)"))
			continue
		}
		return val, nil
	}
}

func SetUserShapeMemory(shapeName string) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(output.Yellow(fmt.Sprintf("\nSpecify the amount of memory (in GB) needed for shape %s: ", shapeName)))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		val, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println(output.Red("\nInvalid input. Please enter a valid value (integer)."))
			continue
		}
		return val, nil
	}
}

func SetDenseIOShapeOCPUs(shapeName string) (float32, error) {
	allowed := DenseIOFlexShapes[shapeName]
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(output.Yellow(fmt.Sprintf("Select the amount of OCPUs %v: ", allowed)))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		for _, v := range allowed {
			if input == v {
				f, _ := strconv.ParseFloat(v, 32)
				return float32(f), nil
			}
		}
		fmt.Println(output.Red(fmt.Sprintf("Invalid input. Please select one of the following values: %v", allowed)))
	}
}

func GetShapeConfig(shapeName string, shapes []core.Shape, ocpus, memory int) ShapeConfig {
	isFlex := strings.Contains(shapeName, ".Flex")
	denseIOFlex := regexp.MustCompile(`^VM\.DenseIO\..{2}\.Flex$`).MatchString(shapeName)

	if denseIOFlex {
		return ShapeConfig{OCPUs: float64(ocpus), Memory: 0, IsFlex: isFlex}
	}

	shapeInfo := findShape(shapes, shapeName)

	if shapeName == "VM.Standard.A2.Flex" || shapeName == "VM.Standard.A4.Flex" {
		if shapeInfo != nil && shapeInfo.OcpuOptions != nil && shapeInfo.OcpuOptions.Max != nil {
			maxOcpus := int(*shapeInfo.OcpuOptions.Max)
			ocpus = clampInt(ocpus, 1, maxOcpus)
			memory = maxInt(1, ocpus*2, memory)
			if shapeInfo.MemoryOptions != nil {
				if shapeInfo.MemoryOptions.MaxPerOcpuInGBs != nil {
					maxPerOcpu := int(*shapeInfo.MemoryOptions.MaxPerOcpuInGBs)
					memory = minInt(memory, ocpus*maxPerOcpu)
				}
				if shapeInfo.MemoryOptions.MaxInGBs != nil {
					memory = minInt(memory, int(*shapeInfo.MemoryOptions.MaxInGBs))
				}
			}
		} else {
			ocpus, memory = 1, 2
		}
	} else if isFlex {
		if shapeInfo != nil && shapeInfo.OcpuOptions != nil && shapeInfo.OcpuOptions.Max != nil {
			ocpus = minInt(ocpus, int(*shapeInfo.OcpuOptions.Max))
			memory = maxInt(1, memory)
			memory = maxInt(ocpus, memory)
			if shapeInfo.MemoryOptions != nil {
				if shapeInfo.MemoryOptions.MaxInGBs != nil {
					memory = minInt(memory, int(*shapeInfo.MemoryOptions.MaxInGBs))
				}
				if shapeInfo.MemoryOptions.MaxPerOcpuInGBs != nil {
					memory = minInt(memory, ocpus*int(*shapeInfo.MemoryOptions.MaxPerOcpuInGBs))
				}
			}
		} else {
			ocpus, memory = 1, 1
		}
	}

	return ShapeConfig{OCPUs: float64(ocpus), Memory: float64(memory), IsFlex: isFlex, ShapeInfo: shapeInfo}
}

func ProcessRegion(
	regionName string,
	provider common.ConfigurationProvider,
	tenancyID string,
	compartmentID string,
	shapeName string,
	ocpus, memory float64,
	drcc bool,
) error {
	identityClient, err := identity.NewClient(provider)
	if err != nil {
		return err
	}
	identityClient.SetRegion(regionName)

	computeClient, err := core.NewComputeClientWithConfigurationProvider(provider)
	if err != nil {
		return err
	}
	computeClient.SetRegion(regionName)

	ads, err := identity.GetAvailabilityDomains(identityClient, tenancyID)
	if err != nil {
		return err
	}

	shapesResp, err := computeClient.ListShapes(context.Background(), core.ListShapesRequest{
		CompartmentId: &tenancyID,
	})
	if err != nil {
		output.PrintError([]string{err.Error()}, "ERROR")
		return err
	}

	cfg := GetShapeConfig(shapeName, shapesResp.Items, int(ocpus), int(memory))

	for _, ad := range ads {
		fds, err := identity.GetFaultDomains(identityClient, tenancyID, ad)
		if err != nil {
			return err
		}

		for _, fd := range fds {
			if err := createAndPrintReport(
				regionName, identityClient, computeClient,
				ad, fd, compartmentID, shapeName, drcc, cfg,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func createAndPrintReport(
	region string,
	identityClient ociidentity.IdentityClient,
	computeClient core.ComputeClient,
	availabilityDomain, faultDomain, compartmentID, shapeName string,
	drcc bool,
	cfg ShapeConfig,
) error {
	shapeOCPUs := cfg.OCPUs
	shapeMemory := cfg.Memory
	var instanceShapeConfig *core.CapacityReportInstanceShapeConfig

	if strings.HasPrefix(shapeName, "BM.") {
		shapeOCPUs, shapeMemory = defaultShapeValues(cfg, shapeOCPUs, shapeMemory)
	} else if cfg.IsFlex {
		if denseCfg, ok := DenseIOShapeConfigs[shapeName][float32(shapeOCPUs)]; ok {
			nvmes := denseCfg.NVMes
			instanceShapeConfig = &core.CapacityReportInstanceShapeConfig{
				Ocpus:       float32Ptr(denseCfg.OCPUs),
				MemoryInGBs: float32Ptr(denseCfg.MemoryInGBs),
				Nvmes:       &nvmes,
			}
			shapeOCPUs = float64(denseCfg.OCPUs)
			shapeMemory = float64(denseCfg.MemoryInGBs)
		} else {
			instanceShapeConfig = &core.CapacityReportInstanceShapeConfig{
				Ocpus:       float32Ptr(float32(shapeOCPUs)),
				MemoryInGBs: float32Ptr(float32(shapeMemory)),
			}
		}
	} else {
		shapeOCPUs, shapeMemory = defaultShapeValues(cfg, shapeOCPUs, shapeMemory)
	}

	reportDetails := core.CreateComputeCapacityReportDetails{
		CompartmentId:       &compartmentID,
		AvailabilityDomain:  &availabilityDomain,
		ShapeAvailabilities: []core.CreateCapacityReportShapeAvailabilityDetails{{
			InstanceShape:       &shapeName,
			FaultDomain:         &faultDomain,
			InstanceShapeConfig: instanceShapeConfig,
		}},
	}

	resp, err := computeClient.CreateComputeCapacityReport(context.Background(), core.CreateComputeCapacityReportRequest{
		CreateComputeCapacityReportDetails: reportDetails,
	})
	if err != nil {
		if svcErr, ok := common.IsServiceError(err); ok {
			if strings.Contains(svcErr.GetMessage(), "Authorization failed") {
				compName, _ := identity.GetCompartmentName(identityClient, compartmentID)
				output.PrintError([]string{
					svcErr.GetMessage(),
					fmt.Sprintf("Please verify that you have the appropriate access to %s", compName),
					"You can restart the script without Admin rights",
					"or, use the '--comp' argument.",
				}, "ERROR")
				return fmt.Errorf("authorization failed")
			}
			output.PrintError([]string{svcErr.GetMessage()}, "ERROR")
			return &RestartFlowError{}
		}
		output.PrintError([]string{err.Error()}, "ERROR")
		return &RestartFlowError{}
	}

	for _, result := range resp.ShapeAvailabilities {
		ocpuStr := formatValue(shapeOCPUs)
		memStr := formatValue(shapeMemory)

		var availCount string
		if drcc {
			if result.AvailableCount != nil {
				availCount = fmt.Sprintf("%16d", *result.AvailableCount)
			} else {
				availCount = fmt.Sprintf("%16s", "-")
			}
			fmt.Printf("%-20s %-30s %-20s %-25s %-10s %-10s %s %s\n",
				region, availabilityDomain, faultDomain, shapeName,
				ocpuStr, memStr, availCount, result.AvailabilityStatus)
		} else {
			fmt.Printf("%-20s %-30s %-20s %-25s %-10s %-10s %s\n",
				region, availabilityDomain, faultDomain, shapeName,
				ocpuStr, memStr, result.AvailabilityStatus)
		}
	}

	return nil
}

func defaultShapeValues(cfg ShapeConfig, ocpus, memory float64) (float64, float64) {
	if ocpus != 0 {
		return ocpus, memory
	}
	if cfg.ShapeInfo != nil {
		if cfg.ShapeInfo.Ocpus != nil {
			ocpus = float64(*cfg.ShapeInfo.Ocpus)
		}
		if cfg.ShapeInfo.MemoryInGBs != nil {
			memory = float64(*cfg.ShapeInfo.MemoryInGBs)
		}
	}
	if ocpus == 0 {
		ocpus = -1
	}
	if memory == 0 {
		memory = -1
	}
	return ocpus, memory
}

func formatValue(v float64) string {
	if v < 0 {
		return "-"
	}
	if v == float64(int(v)) {
		return fmt.Sprintf("%d", int(v))
	}
	return fmt.Sprintf("%.1f", v)
}

func findShape(shapes []core.Shape, name string) *core.Shape {
	for i := range shapes {
		if shapes[i].Shape != nil && *shapes[i].Shape == name {
			return &shapes[i]
		}
	}
	return nil
}

func float32Ptr(f float32) *float32 { return &f }

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func minInt(vals ...int) int {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func maxInt(vals ...int) int {
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
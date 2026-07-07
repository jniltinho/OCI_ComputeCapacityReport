package identity

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Olygo/OCI_ComputeCapacityReport/internal/config"
	"github.com/Olygo/OCI_ComputeCapacityReport/internal/output"
	"github.com/oracle/oci-go-sdk/v65/common"
	ociidentity "github.com/oracle/oci-go-sdk/v65/identity"
)

func NewClient(provider common.ConfigurationProvider) (ociidentity.IdentityClient, error) {
	return ociidentity.NewIdentityClientWithConfigurationProvider(provider)
}

func GetHomeRegion(client ociidentity.IdentityClient, tenancyID string) (*ociidentity.RegionSubscription, error) {
	fmt.Print(output.Yellow("\r => Fetching home region..."))
	fmt.Print(strings.Repeat(" ", 50) + "\r")

	regions, err := client.ListRegionSubscriptions(context.Background(), ociidentity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyID,
	})
	if err != nil {
		return nil, err
	}

	for _, region := range regions.Items {
		if region.IsHomeRegion != nil && *region.IsHomeRegion {
			return &region, nil
		}
	}

	return nil, fmt.Errorf("home region not found")
}

func GetRegionSubscriptionList(client ociidentity.IdentityClient, tenancyID, targetRegion string) ([]ociidentity.RegionSubscription, error) {
	fmt.Print(output.Yellow("\r => Loading regions..."))
	fmt.Print(strings.Repeat(" ", 50) + "\r")

	resp, err := client.ListRegionSubscriptions(context.Background(), ociidentity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyID,
	})
	if err != nil {
		output.PrintError([]string{"Region error:", targetRegion, err.Error()}, "ERROR")
		return nil, err
	}

	subscribed := resp.Items

	if targetRegion == "" {
		for _, region := range subscribed {
			if region.IsHomeRegion != nil && *region.IsHomeRegion {
				output.PrintInfo(output.Green, "Region", "analyzed", *region.RegionName)
				return []ociidentity.RegionSubscription{region}, nil
			}
		}
		return nil, fmt.Errorf("home region not found")
	}

	if strings.EqualFold(targetRegion, "all_regions") {
		output.PrintInfo(output.Green, "Region", "analyzed", "all subscribed regions")
		return subscribed, nil
	}

	regionMap := make(map[string]ociidentity.RegionSubscription)
	for _, region := range subscribed {
		if region.RegionName != nil {
			regionMap[strings.ToLower(*region.RegionName)] = region
		}
	}

	if region, ok := regionMap[strings.ToLower(targetRegion)]; ok {
		output.PrintInfo(output.Green, "Region", "analyzed", targetRegion)
		return []ociidentity.RegionSubscription{region}, nil
	}

	allRegions, err := client.ListRegions(context.Background())
	if err != nil {
		output.PrintError([]string{"Region error:", targetRegion, err.Error()}, "ERROR")
		return nil, err
	}

	ociRegionNames := make(map[string]bool)
	for _, r := range allRegions.Items {
		if r.Name != nil {
			ociRegionNames[strings.ToLower(*r.Name)] = true
		}
	}

	if ociRegionNames[strings.ToLower(targetRegion)] {
		output.PrintError([]string{"Region error:", targetRegion + " is not subscribed"}, "ERROR")
	} else {
		output.PrintError([]string{"Region error:", targetRegion + " does not exist"}, "ERROR")
	}

	return nil, fmt.Errorf("region error")
}

func ValidateRegionConnectivity(regions []ociidentity.RegionSubscription, provider common.ConfigurationProvider, tenancyID string) ([]ociidentity.RegionSubscription, error) {
	var (
		mu        sync.Mutex
		validated []ociidentity.RegionSubscription
		wg        sync.WaitGroup
	)

	for _, region := range regions {
		wg.Add(1)
		go func(r ociidentity.RegionSubscription) {
			defer wg.Done()

			regionName := *r.RegionName
			fmt.Print(output.Yellow(fmt.Sprintf("\r => Checking connectivity to region %s...", regionName)))
			fmt.Print(strings.Repeat(" ", 50) + "\r")

			client, err := ociidentity.NewIdentityClientWithConfigurationProvider(provider)
			if err != nil {
				return
			}
			client.SetRegion(regionName)

			_, err = client.GetTenancy(context.Background(), ociidentity.GetTenancyRequest{TenancyId: &tenancyID})
			if err != nil {
				output.PrintInfo(output.Red, "Region", "error", regionName)
				output.PrintInfo(output.Red, "Region", "status", string(r.Status))
				output.PrintInfo(output.Red, "Region", "ignored", "check domain replication")
				return
			}

			mu.Lock()
			validated = append(validated, r)
			mu.Unlock()
		}(region)
	}

	wg.Wait()

	if len(validated) > 0 {
		return validated, nil
	}

	last := regions[len(regions)-1]
	msg := fmt.Sprintf("%s - %s - %s", *last.RegionName, *last.RegionKey, last.Status)
	output.PrintError([]string{"No available region found", msg, "check domain replication"}, "ERROR")
	return nil, fmt.Errorf("no available region found")
}

func SetUserCompartment(client ociidentity.IdentityClient, opts config.Options, tenancyID string) (string, error) {
	if opts.SuperUser {
		return tenancyID, nil
	}

	if opts.Compartment != "" {
		if id := validateCompartment(client, opts.Compartment); id != "" {
			return id, nil
		}
	}

	reader := bufio.NewReader(os.Stdin)
	validInputs := map[string]bool{
		"Y": true, "YES": true, "N": true, "NO": true, "Q": true, "QUIT": true,
	}

	for {
		fmt.Print(output.Yellow("Do you have Administrator rights at the tenancy level? [Y]es, [N]o, [Q]uit: "))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToUpper(input))

		if !validInputs[input] {
			fmt.Println(output.Red("\nInvalid input. Please enter 'Y', 'Yes', 'N', 'No', 'Q', or 'Quit'\n"))
			continue
		}

		switch input {
		case "Y", "YES":
			fmt.Println()
			return tenancyID, nil

		case "N", "NO":
			for {
				fmt.Print(output.Yellow("\nEnter a compartment OCID to which you have access or [Q]uit: "))
				compInput, _ := reader.ReadString('\n')
				compInput = strings.TrimSpace(strings.ToLower(compInput))

				if compInput == "q" || compInput == "quit" {
					return "", fmt.Errorf("quitting the program as per user request")
				}

				if id := validateCompartment(client, compInput); id != "" {
					fmt.Println()
					return id, nil
				}
			}

		case "Q", "QUIT":
			return "", fmt.Errorf("quitting the program as per user request")
		}
	}
}

func validateCompartment(client ociidentity.IdentityClient, compartmentID string) string {
	resp, err := client.GetCompartment(context.Background(), ociidentity.GetCompartmentRequest{
		CompartmentId: &compartmentID,
	})
	if err != nil {
		if svcErr, ok := common.IsServiceError(err); ok {
			fmt.Println(output.Red(fmt.Sprintf("\nCompartment error: %s => %s - %s", compartmentID, svcErr.GetCode(), svcErr.GetMessage())))
		} else {
			fmt.Println(output.Red(fmt.Sprintf("\nCompartment error: %s => %s", compartmentID, err.Error())))
		}
		return ""
	}

	compartment := resp.Compartment
	if compartment.LifecycleState == ociidentity.CompartmentLifecycleStateActive {
		output.PrintInfo(output.Green, "Compartment", "analyzed", *compartment.Name)
		output.PrintInfo(output.Green, "Compartment", "state", string(compartment.LifecycleState))
		return compartmentID
	}

	fmt.Println(output.Red(fmt.Sprintf("\nCompartment state error: %s is %s", *compartment.Name, compartment.LifecycleState)))
	return ""
}

func GetCompartmentName(client ociidentity.IdentityClient, compartmentID string) (string, error) {
	resp, err := client.GetCompartment(context.Background(), ociidentity.GetCompartmentRequest{
		CompartmentId: &compartmentID,
	})
	if err != nil {
		if svcErr, ok := common.IsServiceError(err); ok {
			output.PrintError([]string{"Compartment_id error:", compartmentID, svcErr.GetCode(), svcErr.GetMessage()}, "ERROR")
		}
		return "", err
	}
	return *resp.Compartment.Name, nil
}

func GetAvailabilityDomains(client ociidentity.IdentityClient, tenancyID string) ([]string, error) {
	var names []string

	resp, err := client.ListAvailabilityDomains(context.Background(), ociidentity.ListAvailabilityDomainsRequest{
		CompartmentId: &tenancyID,
	})
	if err != nil {
		output.PrintError([]string{"Error in get_availability_domains:", err.Error()}, "ERROR")
		return nil, err
	}

	for _, ad := range resp.Items {
		if ad.Name != nil {
			names = append(names, *ad.Name)
		}
	}

	return names, nil
}

func GetFaultDomains(client ociidentity.IdentityClient, tenancyID, availabilityDomain string) ([]string, error) {
	var names []string

	resp, err := client.ListFaultDomains(context.Background(), ociidentity.ListFaultDomainsRequest{
		CompartmentId:      &tenancyID,
		AvailabilityDomain: &availabilityDomain,
	})
	if err != nil {
		output.PrintError([]string{"Error in get_fault_domains:", err.Error()}, "ERROR")
		return nil, err
	}

	for _, fd := range resp.Items {
		if fd.Name != nil {
			names = append(names, *fd.Name)
		}
	}

	return names, nil
}
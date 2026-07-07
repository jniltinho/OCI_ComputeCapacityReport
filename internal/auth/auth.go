package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Olygo/OCI_ComputeCapacityReport/internal/output"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type Result struct {
	Provider  common.ConfigurationProvider
	Tenancy   identity.Tenancy
	AuthName  string
	Details   string
}

func Init(userAuth, configFilePath, configProfile string) (Result, error) {
	errors := make(map[string]string)

	methods := []struct {
		key string
		fn  func(map[string]string) (common.ConfigurationProvider, string, string, error)
	}{
		{"cs", func(e map[string]string) (common.ConfigurationProvider, string, string, error) {
			return authenticateCloudShell(e)
		}},
		{"cf", func(e map[string]string) (common.ConfigurationProvider, string, string, error) {
			return authenticateConfigFile(e, configFilePath, configProfile)
		}},
		{"ip", func(e map[string]string) (common.ConfigurationProvider, string, string, error) {
			return authenticateInstancePrincipals(e)
		}},
	}

	var toTry []struct {
		key string
		fn  func(map[string]string) (common.ConfigurationProvider, string, string, error)
	}

	if userAuth != "" {
		for _, m := range methods {
			if m.key == userAuth {
				toTry = append(toTry, m)
				break
			}
		}
	} else {
		toTry = methods
	}

	for _, method := range toTry {
		provider, authName, details, err := method.fn(errors)
		if err == nil && provider != nil {
			tenancy, err := validateTenancy(provider)
			if err == nil {
				return Result{Provider: provider, Tenancy: tenancy, AuthName: authName, Details: details}, nil
			}
			errors[method.key+"_authentication"] = err.Error()
		}
	}

	output.Clear()
	for name, msg := range errors {
		output.PrintError([]string{name, msg}, "ERROR")
		fmt.Println()
	}

	return retryAuth()
}

func authenticateCloudShell(errors map[string]string) (common.ConfigurationProvider, string, string, error) {
	fmt.Print(output.Yellow("\r => Trying CloudShell authentication..."))
	fmt.Print(strings.Repeat(" ", 50) + "\r")

	configFile := os.Getenv("OCI_CONFIG_FILE")
	configProfile := os.Getenv("OCI_CONFIG_PROFILE")
	if configFile == "" || configProfile == "" {
		errors["CloudShell_authentication"] = fmt.Sprintf(
			"Not a CloudShell session: $OCI_CONFIG_FILE=%s, $OCI_CONFIG_PROFILE=%s",
			configFile, configProfile,
		)
		return nil, "", "", fmt.Errorf("cloudshell not available")
	}

	token, err := readDelegationToken(configFile, configProfile)
	if err != nil {
		errors["CloudShell_authentication"] = err.Error()
		return nil, "", "", err
	}

	provider, err := auth.InstancePrincipalDelegationTokenConfigurationProvider(&token)
	if err != nil {
		errors["CloudShell_authentication"] = err.Error()
		return nil, "", "", err
	}

	return provider, "delegation_token", configFile, nil
}

func authenticateConfigFile(errors map[string]string, configFilePath, configProfile string) (common.ConfigurationProvider, string, string, error) {
	fmt.Print(output.Yellow("\r => Trying Config File authentication..."))
	fmt.Print(strings.Repeat(" ", 50) + "\r")

	expanded := expandPath(configFilePath)
	provider, err := common.ConfigurationProviderFromFileWithProfile(expanded, configProfile, "")
	if err != nil {
		errors["Config_File_authentication"] = err.Error()
		return nil, "", "", err
	}

	if ok, err := common.IsConfigurationProviderValid(provider); !ok {
		errors["Config_File_authentication"] = err.Error()
		return nil, "", "", err
	}

	return provider, "config_file", configProfile, nil
}

func authenticateInstancePrincipals(errors map[string]string) (common.ConfigurationProvider, string, string, error) {
	fmt.Print(output.Yellow("\r => Trying Instance Principals authentication..."))
	fmt.Print(strings.Repeat(" ", 50) + "\r")

	provider, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		errors["Instance_Principals_authentication"] = err.Error()
		return nil, "", "", err
	}

	return provider, "instance_principals", "", nil
}

func validateTenancy(provider common.ConfigurationProvider) (identity.Tenancy, error) {
	tenancyID, err := provider.TenancyOCID()
	if err != nil {
		return identity.Tenancy{}, err
	}

	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	if err != nil {
		return identity.Tenancy{}, err
	}

	resp, err := client.GetTenancy(context.Background(), identity.GetTenancyRequest{TenancyId: &tenancyID})
	if err != nil {
		return identity.Tenancy{}, err
	}

	return resp.Tenancy, nil
}

func retryAuth() (Result, error) {
	fmt.Println(output.Yellow("\n-- All authentication methods have failed -- \n"))
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(output.Yellow("Do you want to specify another config file path ? [Y/N]"))
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToUpper(answer))
	fmt.Println()

	if answer != "Y" && answer != "YES" {
		return Result{}, fmt.Errorf("all authentication methods have failed")
	}

	fmt.Print(output.Yellow("What is the path of your OCI config file: "))
	configPath, _ := reader.ReadString('\n')
	configPath = strings.TrimSpace(configPath)
	fmt.Println()

	fmt.Print(output.Yellow("What is the profile section name to use in this config file: "))
	profile, _ := reader.ReadString('\n')
	profile = strings.TrimSpace(profile)

	errors := make(map[string]string)
	provider, authName, details, err := authenticateConfigFile(errors, configPath, profile)
	if err != nil || provider == nil {
		return Result{}, fmt.Errorf("retrying config_file authentication failed. Please check your file and relaunch the process")
	}

	tenancy, err := validateTenancy(provider)
	if err != nil {
		return Result{}, err
	}

	return Result{Provider: provider, Tenancy: tenancy, AuthName: authName, Details: details}, nil
}

func readDelegationToken(configFile, profile string) (string, error) {
	expanded := expandPath(configFile)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return "", err
	}

	inProfile := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section := strings.Trim(trimmed, "[]")
			inProfile = section == profile
			continue
		}
		if inProfile && strings.HasPrefix(strings.ToLower(trimmed), "delegation_token_file") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				tokenPath := expandPath(strings.TrimSpace(parts[1]))
				content, err := os.ReadFile(tokenPath)
				if err != nil {
					return "", err
				}
				return strings.TrimSpace(string(content)), nil
			}
		}
	}

	return "", fmt.Errorf("delegation_token_file not found in profile %s", profile)
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
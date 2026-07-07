# OCI_ComputeCapacityReport

**Check the Availability of Any Compute Shape Across OCI Regions !**

> **v4.0.0** — This project is now a Go binary. The previous Python implementation (v3.x) is archived in [`legacy/python/`](legacy/python/).

Easily find out which regions offer the latest compute shapes, such as VM.Standard.E6.Flex.

OCI_ComputeCapacityReport provides the availability status down to the Fault Domain level and automatically relaunches after completing the first query or encountering an error.

Output meanings are:

- **AVAILABLE** => The capacity for the specified shape is currently available.
- **HARDWARE_NOT_SUPPORTED** => The necessary hardware has not yet been deployed in this region.
- **OUT_OF_HOST_CAPACITY** => Additional hardware is currently being deployed in this region.

DRCC customers and whitelisted tenancies can display the value of **available_count** by using the **--drcc** flag.

This tool automatically discovers new shapes and does not require updates to maintain a complete list of existing shapes in the code.

If you want to check the availability of a shape that is offered only in a specific region (and not in your home region), you must connect directly to the target region.

```bash
./bin/oci-compute-capacity-report --region eu-paris-1
```

## Quick Start

```bash
git clone https://github.com/Olygo/OCI_ComputeCapacityReport
cd OCI_ComputeCapacityReport
make build
./bin/oci-compute-capacity-report
```

### Requirements

- Go 1.26 or later
- [UPX](https://upx.github.io/) (used by `make build` to compress the binary)
- OCI config file (`~/.oci/config`) or CloudShell / Instance Principal environment

## Build

```bash
# Build and compress with UPX (default)
make build

# Build without UPX
make build-noupx

# Install to /usr/local/bin
make install

# Clean build artifacts
make clean
```

The compressed binary is written to `bin/oci-compute-capacity-report`.

## Table of Contents

- [How to Use OCI_ComputeCapacityReport](#how-to-use-oci_computecapacityreport)
- [CLI options](#cli-options)
- [Optional parameters for execution](#optional-parameters-for-execution)
- [Examples of Usage](#examples-of-usage)
- [Setup](#setup)
- [Screenshots](#screenshots)
- [Compute Shapes Tested and Validated](#compute-shapes-tested-and-validated-as-of-september-1-2024)
- [Legacy Python version](#legacy-python-version)
- [Questions and Feedback](#questions-and-feedback)
- [Disclaimer](#disclaimer)

## How to Use OCI_ComputeCapacityReport

```bash
./bin/oci-compute-capacity-report
```

When no flags are provided, OCI_ComputeCapacityReport **automatically**:

- Attempts to authenticate using all available authentication methods:

    1. CloudShell authentication
    2. Config file authentication
    3. Instance Principal authentication
    4. If all authentication methods fail, prompts the user to provide a config file custom path and a config profile section.

- Selects the tenancy's Home Region
- Asks if the user is a tenancy Admin
    - If user is not a tenancy Admin, asks for a compartment OCID
    - Using `--su` will bypass this question
- Displays available [compute shape names](https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm)
- Asks user to enter a compute shape name

## CLI options

OCI_ComputeCapacityReport can be fully automated using the following flags:

- Enforce an authentication method: **--auth cs** | **cf** | **ip**
- Tenant administrators can bypass the 'Admin question' using **--su**
- Non-admin users can specify their own compartment OCID: **--comp ocid1.xxxx**
- Select a specific subscribed region instead of the home region: **--region eu-frankfurt-1**
- Target all subscribed regions: **--region all_regions**
- Specify a shape name: **--shape VM.Standard.E5.Flex**
- Optionally define oCPUs and memory amount: **--ocpus 10** **--memory 30**

**OCI_ComputeCapacityReport** ensures the correct oCPU-to-memory ratio for each shape type.
If an invalid configuration is entered, such as requesting 10 oCPUs with only 2 GB of memory,
it automatically adjusts the values to meet the required specifications.
Similarly, it enforces both the minimum and maximum limits for oCPUs and memory based on the shape type.

## Optional parameters for execution

| Flag            | Parameter            | Description                                                                                        |
| --------------- | -------------------- | -------------------------------------------------------------------------------------------------- |
| --auth          | auth_method          | Force an authentication method: 'cs' (cloudshell), 'cf' (config file), 'ip' (instance principals) |
| --config-file   | config_file_path     | Path to your OCI config file, default: '~/.oci/config'                                              |
| --profile       | config_profile       | Config file section to use, default: 'DEFAULT'                                                     |
| --su            |                      | Notify the tool that you have tenancy-level admin rights to prevent prompting                      |
| --comp          | compartment_ocid     | Filter on a compartment when you do not have Admin rights at the tenancy level                     |
| --region        | region_name          | Region name to analyze, e.g. "eu-frankfurt-1" or "all_regions", default: home region               |
| --shape         | shape_name           | Compute shape name you want to analyze                                                             |
| --ocpus         | integer              | Specify a particular amount of oCPU                                                                |
| --memory        | integer              | Specify a particular amount of memory (GB)                                                         |
| --drcc          |                      | Display 'available_count' value for DRCC customers and whitelisted tenancies                       |

## Examples of Usage

### Default (interactive)

```bash
./bin/oci-compute-capacity-report
```

Tries all authentication methods, checks capacity in the Home Region only and prompts for a compute shape name.

### Config file authentication

Uses `~/.oci/config` and the `DEFAULT` profile:

```bash
./bin/oci-compute-capacity-report --auth cf
```

Custom config file path and profile:

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --config-file ~/Documents/config/my_config \
  --profile OCI_FBO
```

### Home region — fully automated

Checks capacity in the tenancy home region without prompts (requires tenancy admin rights):

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --su \
  --shape VM.Standard.E5.Flex \
  --ocpus 1 \
  --memory 1
```

Example output:

```
*****      Script               started                      oci-compute-capacity-report *****
*****      Script               version                                            4.0.0 *****
*****      Login                success                                      config_file *****
*****      Login                profile                                          DEFAULT *****
*****      Tenancy              my-tenancy                              home region: IAD *****
*****      Region               analyzed                                    us-ashburn-1 *****
*****      Shape                analyzed                             VM.Standard.E5.Flex *****
*****      oCPUs                amount                                           1 cores *****
*****      Memory               amount                                             1 gbs *****

REGION               AVAILABILITY_DOMAIN            FAULT_DOMAIN         SHAPE                     OCPU       MEMORY     AVAILABILITY
us-ashburn-1         VXXt:US-ASHBURN-AD-1           FAULT-DOMAIN-1       VM.Standard.E5.Flex       1          1          AVAILABLE
us-ashburn-1         VXXt:US-ASHBURN-AD-1           FAULT-DOMAIN-2       VM.Standard.E5.Flex       1          1          AVAILABLE
us-ashburn-1         VXXt:US-ASHBURN-AD-1           FAULT-DOMAIN-3       VM.Standard.E5.Flex       1          1          AVAILABLE
...
```

After the report, the tool relaunches automatically and prompts for another shape name.

### Specific region

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --su \
  --region eu-paris-1 \
  --shape VM.Standard.E6.Flex \
  --ocpus 4 \
  --memory 16
```

### All subscribed regions

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --su \
  --region all_regions \
  --shape VM.Standard.E5.Flex \
  --ocpus 10 \
  --memory 20
```

### Full automated run

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --config-file ~/Documents/config/my_config \
  --profile OCI_FBO \
  --shape VM.Standard.E5.Flex \
  --ocpus 10 \
  --memory 20 \
  --region eu-paris-1 \
  --su
```

### DRCC — display available_count

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --su \
  --shape VM.Standard.E5.Flex \
  --ocpus 4 \
  --memory 16 \
  --drcc
```

Adds an `AVAILABLE_COUNT` column to the report output.

### Instance Principal (OCI compute instance)

```bash
./bin/oci-compute-capacity-report \
  --auth ip \
  --su \
  --shape VM.Standard.E5.Flex \
  --ocpus 2 \
  --memory 8
```

### Non-admin user with compartment

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --comp ocid1.compartment.oc1..aaaaaaaaxxxxx \
  --shape VM.Standard.E5.Flex \
  --ocpus 2 \
  --memory 8
```

### CloudShell

```bash
./bin/oci-compute-capacity-report --auth cs --su --shape VM.Standard.E5.Flex
```

### Run tests

```bash
# Unit tests
go test ./... -v

# Build and smoke test
make build
./bin/oci-compute-capacity-report --auth cf --su --shape VM.Standard.E5.Flex --ocpus 1 --memory 1
```

## Setup

##### Download and build locally

```bash
git clone https://github.com/Olygo/OCI_ComputeCapacityReport
cd OCI_ComputeCapacityReport
make build
```

If you run this tool from an OCI compute instance you should use [Instance Principal authentication](https://docs.public.oneportal.content.oci.oraclecloud.com/en-us/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm).

When using Instance Principal authentication, you need to create the following resources:

##### Dynamic Group

Create a Dynamic Group called OCI_Scripting and add the OCID of your instance to the group:

```
ANY {instance.id = 'OCID_of_your_Compute_Instance'}
```

##### Policy

Create a policy in the root compartment, giving your dynamic group the permissions to read resources in tenancy:

```
allow dynamic-group 'Your_Identity_Domain_Name'/'OCI_Scripting' to read all_resources in tenancy
```

## Screenshots

##### Default run, prompt user for a compute shape name

![00](./.images/00.png)

##### Support of Flexible DenseIO compute shapes

![01](./.images/01.png)

##### Run with parameters

```bash
./bin/oci-compute-capacity-report \
  --auth cf \
  --config-file ~/Documents/config/my_config \
  --profile OCI_FBO \
  --shape VM.Standard.E5.Flex \
  --ocpus 10 \
  --memory 20 \
  --region eu-paris-1 \
  --su
```

![02](./.images/02.png)

##### Run without tenancy admin rights

![03](./.images/03.png)

## Compute Shapes Tested and Validated as of September 1, 2024

This list includes all the compute shapes that have been tested with this tool.
However, it does not represent the only shapes you can use.
If a new compute shape is released and isn't included in this list, it should still work correctly.
If you encounter any issues, please let me know.

|                     |                      |                        |                       |
| ------------------- | -------------------- | ---------------------- | --------------------- |
| BM.DenseIO1.36      | BM.DenseIO2.52       | BM.DenseIO.E4.128      | BM.DenseIO.E5.128     |
| BM.GPU2.2           | BM.GPU3.8            | BM.GPU4.8              | BM.GPU.A10.4          |
| BM.GPU.A100-v2.8.   | BM.GPU.H100.8        | BM.GPU.L40S.4          | BM.HPC2.36            |
| BM.HPC.E5.144       | BM.Optimized3.36     | BM.Standard1.36        | BM.Standard2.52       |
| BM.Standard3.64     | BM.Standard.A1.160   | BM.Standard.B1.44      | BM.Standard.E2.64     |
| BM.Standard.E3.128  | BM.Standard.E4.128   | BM.Standard.E5.192     | VM.DenseIO1.16        |
| VM.DenseIO1.4       | VM.DenseIO1.8        | VM.DenseIO2.16         | VM.DenseIO2.24.       |
| VM.DenseIO2.8       | VM.DenseIO.E4.Flex   | VM.DenseIO.E5.Flex     | VM.GPU2.1             |
| VM.GPU3.1           | VM.GPU3.2            | VM.GPU3.4              | VM.GPU.A10.1          |
| VM.GPU.A10.2        | VM.Optimized3.Flex.  | VM.Standard1.1         | VM.Standard1.16       |
| VM.Standard1.2      | VM.Standard1.4       | VM.Standard1.8         | VM.Standard2.1.       |
| VM.Standard2.16     | VM.Standard2.2       | VM.Standard2.24        | VM.Standard2.4        |
| VM.Standard2.8      | VM.Standard3.Flex    | VM.Standard.A1.Flex    | VM.Standard.A2.Flex   |
| VM.Standard.B1.1    | VM.Standard.B1.16    | VM.Standard.B1.2.      | VM.Standard.B1.4      |
| VM.Standard.B1.8    | VM.Standard.E2.1     | VM.Standard.E2.1.Micro | VM.Standard.E2.2      |
| VM.Standard.E2.4    | VM.Standard.E2.8     | VM.Standard.E3.Flex    | VM.Standard.E4.Flex   |
| VM.Standard.E5.Flex |                      |                        |                       |

## Legacy Python version

The Python script (v3.0.4) has been archived and is no longer maintained. It remains available under [`legacy/python/`](legacy/python/) for reference.

```bash
python3 -m pip install oci -U --user
cd legacy/python
python3 ./OCI_ComputeCapacityReport.py
```

See [`legacy/python/README.md`](legacy/python/README.md) for legacy CLI flag mapping and usage notes.

## Questions and Feedback

**_olygo.git@gmail.com_**

## Disclaimer

**Always ensure thorough testing of any tool on test resources prior to deployment in a production environment to avoid potential outages or unexpected costs. The OCI_ComputeCapacityReport tool does not interact with or create any resources in your existing environment.**

**This tool is an independent project developed by Florian Bonneville and is not affiliated with or supported by Oracle.
It is provided as-is and without any warranty or official endorsement from Oracle**
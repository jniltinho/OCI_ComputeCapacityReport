# Changelog

## Version 4.0.0

- Rewrite the tool in Go (replaces the Python implementation)
- Single binary build with `make build` (UPX-compressed by default)
- CLI flags now use double-dash syntax (`--auth`, `--shape`, `--region`, etc.)
- Python v3.x source archived under `legacy/python/`

## Version 3.0.4

- Update default shape list
- Add support for VM.Standard.A4.Flex CPU/memory ratio

## Version 3.0.3

- Add `available_count` value for DRCC customers and whitelisted tenancies

## Version 3.0.2

- Handle specific use-case exceptions with flexible shapes
- Increase region connectivity timeout and improve exception handling

## Version 3.0.1

- Minor fixes

## Version 3.0.0

### New Features

- Add interactive support for admin and non-admin users; admins can bypass the prompt using `-su`
- Add oCPU and memory filters and display
- Add a loop so the script automatically relaunches after completing the first query or encountering an error
- Add re-authentication: prompt the user for a config file if all prior authentication methods have failed
- Improve code structure and error management

## Version 2.0.1

### New Features

- Add support for non-admin users:
  - Add the `set_user_compartment` function in `legacy/python/modules/identity.py`
  - Include the `-su` argument to let administrators bypass the script prompt
  - Include the `-comp` argument to let non-admin users specify their compartment and bypass the script prompt
- Implement a compartment state check
- Improve the error log message in the `create_and_print_report` function

## Version 2.0.0

### New Features

- Automated authentication testing: automatically tests all available authentication methods by default, removing the need to specify command-line arguments for authentication
- Manual authentication selection: force a specific authentication method using the `-auth` argument with options `cs`, `cf`, or `ip`
- Dynamic region handling: check connectivity to a region before executing any requests against it
- Improved region management: analyze the home region by default; use `-region` for a specific target region or `-region all_regions` for all subscribed regions
- Shape verification and input handling: verify that a given `shape_name` exists before running the capacity report; prompt the user and display available shapes when none is provided
- Support for DenseIO flexible VM shapes with customized resource configurations
- Custom oCPUs and memory: force specific oCPU and memory values for requested shapes
- Automatic shape list update: keep the list of available shapes current without manual code changes
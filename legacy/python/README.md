# Legacy Python implementation (v3.x)

This directory contains the archived Python implementation of OCI_ComputeCapacityReport (last release: **3.0.4**).

The project has been rewritten in Go as of **v4.0.0**. Use the Go binary for new deployments:

```bash
make build
./bin/oci-compute-capacity-report
```

See the [main README](../../README.md) for current documentation.

## Running the legacy script

If you still need the Python version:

```bash
python3 -m pip install oci -U --user
cd legacy/python
python3 ./OCI_ComputeCapacityReport.py
```

### Legacy CLI flags

The Python script uses single-dash flags (e.g. `-auth`, `-shape`, `-region`). The Go binary uses double-dash flags (`--auth`, `--shape`, `--region`). Behavior is otherwise equivalent.

| Python (legacy) | Go (current) |
| --------------- | ------------ |
| `-auth`         | `--auth`     |
| `-config_file`  | `--config-file` |
| `-profile`      | `--profile`  |
| `-su`           | `--su`       |
| `-comp`         | `--comp`     |
| `-region`       | `--region`   |
| `-shape`        | `--shape`    |
| `-ocpus`        | `--ocpus`    |
| `-memory`       | `--memory`   |
| `-drcc`         | `--drcc`     |

## Status

This code is **frozen** and will not receive new features. It is kept for reference and backward compatibility only.
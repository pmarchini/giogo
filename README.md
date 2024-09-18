# Giogo

Giogo is a command-line tool that allows you to run processes with specified resource limitations using Linux cgroups.  
It provides an easy-to-use interface to limit CPU and memory resources, etc. for a process and its children. 

**Note: Root privileges are required, and cgroups v1 is currently not supported.**

> Giogo means "yoke" in Italian

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Available Flags](#available-flags)
  - [CPU Limitations](#cpu-limitations)
  - [Memory Limitations](#memory-limitations)
- [Examples](#examples)

## Features

- **CPU Limiting**: Restrict CPU usage as a fraction of total CPU time.
- **Memory Limiting**: Set maximum memory usage.
- **Cgroups Support**: Works with cgroups v2 only (cgroups v1 is not supported at this time).
- **Process Isolation**: Limits apply to the process and all its child processes.

## Installation

### Prerequisites

- **Linux** operating system with cgroups v2 enabled.
- **Root privileges**: Required for setting cgroup limitations.

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/giogo.git

# Change to the giogo directory
cd giogo

# Build the executable
go build -o giogo main.go
```

### Install Binary (Optional)

You can move the `giogo` binary to a directory in your `PATH` for easier access:

```bash
sudo mv giogo /usr/local/bin/
```

## Usage

```bash
sudo giogo [flags] -- command [arguments]
```

- **`[flags]`**: Resource limitation flags (e.g., `--cpu`, `--ram`).
- **`--`**: Separator between giogo flags and the command to execute.
- **`command [arguments]`**: The command you wish to run with resource limitations.

**Note:** Root privileges are required, so use `sudo` when running `giogo`.

## Available Flags

Giogo supports various flags to set cgroup resource limitations:

### CPU Limitations

- **`--cpu=VALUE`**

  Limit the CPU usage of the process.

  - **`VALUE`**: A decimal between `0` and `1`, representing the fraction of a single CPU core.
  - **Example**: `--cpu=0.5` limits CPU usage to 50% of one core.

### Memory Limitations

- **`--ram=VALUE`**

  Limit the memory usage of the process.

  - **`VALUE`**: Memory limit with units (`k`, `m`, `g`). Defaults to bytes if no unit is specified.
  - **Units**:
    - `k` or `K`: Kilobytes
    - `m` or `M`: Megabytes
    - `g` or `G`: Gigabytes
  - **Example**: `--ram=256m` limits RAM usage to 256 Megabytes.

## Examples

### Limit CPU and Memory

```bash
sudo giogo --cpu=0.2 --ram=128m -- your_command --option1 --option2
```

- **Description**: Runs `your_command` with CPU usage limited to 20% of a single core and maximum RAM usage of 128 MB.

### Full Resource Limitation

```bash
sudo giogo --cpu=0.5 --ram=1g -- python3 heavy_script.py
```

- **Description**: Runs `heavy_script.py` with CPU and RAM limitations.

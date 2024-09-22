# Giogo

Giogo is a command-line tool that allows you to run processes with specified resource limitations using Linux cgroups.  
It provides an easy-to-use interface to limit CPU, memory, and IO resources for a process and its children.

**Note: Root privileges are required, and cgroups v1 is currently not supported.**

> Giogo means "yoke" in Italian

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Available Flags](#available-flags)
  - [CPU Limitations](#cpu-limitations)
  - [Memory Limitations](#memory-limitations)
  - [IO Limitations](#io-limitations)
- [Examples](#examples)

## Features

- **CPU Limiting**: Restrict CPU usage as a fraction of total CPU time.
- **Memory Limiting**: Set maximum memory usage.
- **IO Limiting**: Control IO read and write bandwidth.
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

- **`[flags]`**: Resource limitation flags (e.g., `--cpu`, `--ram`, `--io-read-max`, `--io-write-max`).
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

### IO Limitations

- **`--io-read-max=VALUE`**

  Set a bandwidth throttle on read operations for every block device's IO.

  - **`VALUE`**: Maximum read bandwidth using the same notation as memory (`k`, `m`, `g`).
  - **Units**:
    - `k` or `K`: Kilobytes per second
    - `m` or `M`: Megabytes per second
    - `g` or `G`: Gigabytes per second
  - **Example**: `--io-read-max=1m` limits IO read to 1 MB/s.

- **`--io-write-max=VALUE`**

  Set a bandwidth throttle on write operations for every block device's IO.

  - **`VALUE`**: Maximum write bandwidth using the same notation as memory (`k`, `m`, `g`).
  - **Units**:
    - `k` or `K`: Kilobytes per second
    - `m` or `M`: Megabytes per second
    - `g` or `G`: Gigabytes per second
  - **Example**: `--io-write-max=512k` limits IO write to 512 KB/s.

**Note:**  
By default, Giogo sets a bandwidth throttle on every block device's IO. The Linux kernel uses caching by default, which means that `io-write-max`, with fallback on `io-read-max`, is also set as a RAM limit unless another RAM limit is explicitly declared. If you need to bypass this behavior, set a high value for the RAM limit using the `--ram` flag.

**Additional Note:**  
If your operations utilize the `O_DIRECT` flag, the RAM limit is not required, as `O_DIRECT` bypasses the kernel's caching mechanism.

## Examples

### Limit CPU and Memory

```bash
sudo giogo --cpu=0.2 --ram=128m -- your_command --option1 --option2
```

- **Description**: Runs `your_command` with CPU usage limited to 20% of a single core and maximum RAM usage of 128 MB.

### Full Resource Limitation

```bash
sudo giogo --cpu=0.5 --ram=1g --io-read-max=1m --io-write-max=512k -- python3 heavy_script.py
```

- **Description**: Runs `heavy_script.py` with CPU usage limited to 50% of one core, RAM usage limited to 1 GB, IO read limited to 1 MB/s, and IO write limited to 512 KB/s.

### IO-Only Limitation with High RAM Limit

```bash
sudo giogo --io-read-max=2m --io-write-max=1m --ram=2g -- your_io_intensive_command
```

- **Description**: Runs `your_io_intensive_command` with IO read limited to 2 MB/s and IO write limited to 1 MB/s, while setting a high RAM limit of 2 GB to bypass the default association between `io-write-max` and RAM usage.

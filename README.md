# Giogo

Giogo is a command-line tool that allows you to run processes with specified resource limitations using Linux cgroups.  
It provides an easy-to-use interface to limit CPU, memory, IO, and network resources for a process and its children.

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
  - [Network Limitations](#network-limitations)
- [Examples](#examples)

## Features

- **CPU Limiting**: Restrict CPU usage as a fraction of total CPU time.
- **Memory Limiting**: Set maximum memory usage.
- **IO Limiting**: Control IO read and write bandwidth.
- **Network Limiting**: Set network class identifiers and priorities for network traffic.
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

### Network Limitations

- **`--network-class-id=VALUE`**

  Set a class identifier for the container's network packets.

  - **`VALUE`**: A numeric identifier (uint32) used for packet classification and traffic control.
  - **Example**: `--network-class-id=100` sets the network class identifier to 100.
  - **Use Case**: This can be used in conjunction with Linux traffic control (tc) for advanced network QoS configuration.

- **`--network-priority=VALUE`**

  Set the priority of network traffic for the container.

  - **`VALUE`**: A numeric priority value (uint32), where higher values typically indicate higher priority.
  - **Example**: `--network-priority=50` sets the network traffic priority to 50.
  - **Use Case**: Helps prioritize network traffic when multiple processes compete for bandwidth.

- **`--network-max-bandwidth=VALUE`**

  Set a maximum network bandwidth limit for the container. Requires `--network-class-id` to be set.

  - **`VALUE`**: Maximum bandwidth using the same notation as memory (`k`, `m`, `g`).
  - **Units**:
    - `k` or `K`: Kilobytes per second
    - `m` or `M`: Megabytes per second
    - `g` or `G`: Gigabytes per second
  - **Example**: `--network-max-bandwidth=1m` limits network bandwidth to 1 MB/s.
  - **Use Case**: Enforces hard bandwidth limits on network traffic using Linux traffic control (tc) with HTB qdisc.

**Note:**  
Network limitations work with cgroups v2's network controller to provide packet classification and prioritization. The priority setting applies to all network interfaces in the container.

When `--network-max-bandwidth` is specified with `--network-class-id`, giogo automatically configures Linux traffic control (tc) with HTB (Hierarchical Token Bucket) to enforce the bandwidth limit. The tc rules are automatically cleaned up when the process exits.

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

### Network Traffic Control

```bash
sudo giogo --network-class-id=100 --network-priority=50 -- your_network_intensive_app
```

- **Description**: Runs `your_network_intensive_app` with network class identifier set to 100 and network priority set to 50, allowing for packet classification and traffic prioritization.

### Network Bandwidth Limiting

```bash
sudo giogo --network-class-id=100 --network-max-bandwidth=1m -- your_app
```

- **Description**: Runs `your_app` with network bandwidth limited to 1 MB/s. This automatically configures traffic control (tc) with HTB qdisc to enforce the limit.

### Combined Resource Limitation

```bash
sudo giogo --cpu=0.5 --ram=512m --network-class-id=200 --network-max-bandwidth=500k -- your_app
```

- **Description**: Runs `your_app` with CPU limited to 50% of one core, RAM limited to 512 MB, network class identifier set to 200, and network bandwidth limited to 500 KB/s.

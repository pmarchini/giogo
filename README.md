# Giogo

Giogo is a command-line tool that allows you to run processes with specified resource limitations using Linux cgroups. It provides an easy-to-use interface to limit CPU, memory, and IO resources for a process and its children.

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
- **IO Limiting**: Control maximum read and write rates.
- **Cgroups Support**: Works with both cgroups v1 and v2.
- **Process Isolation**: Limits apply to the process and all its child processes.

## Installation

### Prerequisites

- **Linux** operating system with cgroups enabled.

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
giogo [flags] -- command [arguments]
```

- **`[flags]`**: Resource limitation flags (e.g., `--cpu`, `--ram`).
- **`--`**: Separator between giogo flags and the command to execute.
- **`command [arguments]`**: The command you wish to run with resource limitations.

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

  Limit the maximum read rate.

  - **`VALUE`**: Maximum read bandwidth in kilobytes per second (`k` or `K`).
  - **Example**: `--io-read-max=128k` limits read operations to 128 KB/s.

- **`--io-write-max=VALUE`**

  Limit the maximum write rate.

  - **`VALUE`**: Maximum write bandwidth in kilobytes per second (`k` or `K`).
  - **Example**: `--io-write-max=256k` limits write operations to 256 KB/s.

## Examples

### Limit CPU and Memory

```bash
giogo --cpu=0.2 --ram=128m -- your_command --option1 --option2
```

- **Description**: Runs `your_command` with CPU usage limited to 20% of a single core and maximum RAM usage of 128 MB.

### Limit IO Bandwidth

```bash
giogo --io-read-max=512k --io-write-max=256k -- dd if=/dev/zero of=/tmp/testfile bs=1M count=1024
```

- **Description**: Runs `dd` to write a 1 GB file, limiting read to 512 KB/s and write to 256 KB/s.

### Full Resource Limitation

```bash
giogo --cpu=0.5 --ram=1g --io-read-max=1m --io-write-max=1m -- python3 heavy_script.py
```

- **Description**: Runs `heavy_script.py` with CPU, RAM, and IO limitations.

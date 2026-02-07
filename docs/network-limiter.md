# NetworkLimiter Documentation

## Overview

The `NetworkLimiter` provides comprehensive network bandwidth control for processes running under giogo. It uses Linux Traffic Control (tc) with cgroups v2 integration to enforce bandwidth limits on both egress (outgoing) and ingress (incoming) traffic.

## Architecture

The NetworkLimiter implements two main patterns for bandwidth control:

### 1. Egress (Outgoing) Traffic Control

**Pattern: Direct HTB on main interface**

```
Process → cgroup (classID) → tc HTB qdisc → Bandwidth limit enforced
```

**How it works:**

1. **Cgroup Classification**: Process packets are marked with a cgroup classID via the cgroups v2 network controller
2. **HTB Qdisc Setup**: An HTB (Hierarchical Token Bucket) qdisc is created on the main network interface
3. **Traffic Class**: A traffic class is created within the HTB qdisc with the specified rate limit
4. **Cgroup Filter**: A cgroup-based filter matches packets from the process to the traffic class
5. **Rate Enforcement**: HTB enforces the bandwidth limit (rate = ceil for strict limiting)

**tc commands executed:**
```bash
# Create root HTB qdisc
tc qdisc add dev <interface> root handle 1: htb default 30

# Add bandwidth-limited class
tc class add dev <interface> parent 1: classid 1:<classID> htb rate <bandwidth>bit ceil <bandwidth>bit

# Add cgroup filter to match packets
tc filter add dev <interface> parent 1: protocol ip prio 1 handle <classID> cgroup
```

### 2. Ingress (Incoming) Traffic Control

**Pattern: Redirect ingress → IFB → tc rules**

```
Incoming packets → Ingress qdisc → Redirect to IFB → HTB on IFB → Bandwidth limit enforced
```

**Why IFB?**

Linux tc cannot directly apply bandwidth limiting on ingress traffic because ingress happens before the kernel can apply shaping. The solution is to use an IFB (Intermediate Functional Block) device:
- IFB is a virtual network device that acts as a traffic redirection target
- Incoming traffic is redirected to the IFB device
- On the IFB device, the traffic is treated as egress, allowing HTB shaping to work

**How it works:**

1. **Load IFB Module**: The `ifb` kernel module is loaded
2. **Create IFB Device**: A virtual IFB device is created (named `ifb-<interface>`)
3. **Bring Up IFB**: The IFB device is activated
4. **Setup Ingress Redirection**: 
   - An ingress qdisc is added to the main interface
   - A u32 filter redirects all ingress traffic to the IFB device using the `mirred` action
5. **Apply HTB on IFB**: Standard HTB qdisc with bandwidth limiting is set up on the IFB device
6. **Cgroup Filter on IFB**: A cgroup filter matches packets on the IFB device
7. **Rate Enforcement**: HTB on IFB enforces the ingress bandwidth limit

**tc commands executed:**
```bash
# Load IFB module
modprobe ifb

# Create IFB device
ip link add name ifb-<interface> type ifb
ip link set dev ifb-<interface> up

# Setup ingress redirection
tc qdisc add dev <interface> ingress
tc filter add dev <interface> parent ffff: protocol ip u32 match u32 0 0 flowid 1:1 \
    action mirred egress redirect dev ifb-<interface>

# Setup HTB on IFB device
tc qdisc add dev ifb-<interface> root handle 1: htb default 30
tc class add dev ifb-<interface> parent 1: classid 1:<classID> htb rate <bandwidth>bit ceil <bandwidth>bit
tc filter add dev ifb-<interface> parent 1: protocol ip prio 1 handle <classID> cgroup
```

## Usage

### CLI Flags

- `--network-class-id=VALUE`: Network class identifier for packet classification (required for bandwidth limiting)
- `--network-priority=VALUE`: Network priority for traffic prioritization
- `--network-max-bandwidth=VALUE`: Maximum egress (outgoing) bandwidth limit
- `--network-max-bandwidth-ingress=VALUE`: Maximum ingress (incoming) bandwidth limit

### Examples

#### Egress-Only Bandwidth Limiting

```bash
sudo giogo --network-class-id=100 --network-max-bandwidth=1m -- your_app
```

Limits outgoing bandwidth to 1 MB/s (megabytes per second).

#### Ingress-Only Bandwidth Limiting

```bash
sudo giogo --network-class-id=100 --network-max-bandwidth-ingress=500k -- your_app
```

Limits incoming bandwidth to 500 KB/s (kilobytes per second).

#### Bidirectional Bandwidth Limiting

```bash
sudo giogo --network-class-id=100 --network-max-bandwidth=1m --network-max-bandwidth-ingress=500k -- your_app
```

Limits outgoing bandwidth to 1 MB/s and incoming bandwidth to 500 KB/s independently.

#### Combined with Other Limiters

```bash
sudo giogo --cpu=0.5 --ram=512m --network-class-id=100 --network-max-bandwidth=1m -- your_app
```

Combines CPU, memory, and network bandwidth limits.

## Bandwidth Units

The bandwidth values support the following units:
- `k` or `K`: Kilobytes per second (1024 bytes/sec)
- `m` or `M`: Megabytes per second (1024 KB/sec)
- `g` or `G`: Gigabytes per second (1024 MB/sec)
- No unit: Bytes per second

Examples:
- `500k` = 500 KB/s = 512,000 bytes/sec
- `1m` = 1 MB/s = 1,048,576 bytes/sec
- `2g` = 2 GB/s = 2,147,483,648 bytes/sec

## Implementation Details

### Code Structure

- **`internal/limiter/network.go`**: Core NetworkLimiter implementation
  - `NetworkLimiter` struct with classID, priority, and bandwidth fields
  - Implements `ResourceLimiter` interface (Apply method for cgroup config)
  - Implements `LifecycleLimiter` interface (Setup/Cleanup methods for tc)

- **`internal/limiter/network_tc.go`**: Traffic control operations
  - `setupHTB()`: Egress bandwidth setup
  - `setupIngressHTB()`: Ingress bandwidth setup with IFB
  - `cleanupHTB()`: Egress cleanup
  - `cleanupIngressHTB()`: Ingress cleanup with IFB removal

### Lifecycle

1. **Initialization** (`NewNetworkLimiter`):
   - Parse bandwidth values
   - Detect default network interface
   - Create IFB device name

2. **Setup** (`Setup` method):
   - Called before process starts
   - Sets up egress tc rules if `MaxBandwidth > 0`
   - Sets up ingress tc rules (with IFB) if `MaxBandwidthIngress > 0`
   - Rolls back on error

3. **Cleanup** (`Cleanup` method):
   - Called when process exits
   - Removes egress tc rules
   - Removes ingress tc rules and deletes IFB device
   - Non-fatal errors are logged but don't fail the cleanup

### Error Handling

- **Setup errors**: If tc setup fails, the NetworkLimiter performs cleanup of any successfully configured components before returning the error
- **Cleanup errors**: Cleanup operations continue even if individual steps fail, with the first error being returned (but process exit is not blocked)
- **Interface validation**: The interface name is validated before attempting tc setup

## Requirements

### System Requirements

- Linux kernel with cgroups v2 support
- `tc` (traffic control) utility from iproute2 package
- `ifb` kernel module (for ingress bandwidth limiting)
- Root privileges (required for tc operations)

### Kernel Modules

The IFB module is automatically loaded when ingress bandwidth limiting is used:
```bash
modprobe ifb
```

### Verification

You can verify the tc rules are active:

```bash
# Check egress rules on main interface
tc qdisc show dev eth0
tc class show dev eth0
tc filter show dev eth0

# Check ingress rules (if ingress bandwidth is set)
tc qdisc show dev eth0 ingress
tc filter show dev eth0 ingress

# Check IFB device and rules
ip link show ifb-eth0
tc qdisc show dev ifb-eth0
tc class show dev ifb-eth0
tc filter show dev ifb-eth0
```

## Limitations and Considerations

1. **Interface Detection**: Currently uses the default route interface. For multi-interface systems, this may need customization.

2. **IFB Device Naming**: IFB devices are named `ifb-{interface}`. If multiple giogo instances run simultaneously on the same interface with ingress limiting, they will conflict.

3. **HTB Qdisc Sharing**: If an HTB qdisc already exists on the interface, it's reused. This allows multiple giogo instances to coexist, but class IDs must be unique.

4. **Cleanup on Crash**: If giogo crashes or is killed with SIGKILL, tc rules and IFB devices may persist. Use `tc qdisc del dev <interface> root` and `ip link del ifb-<interface>` to clean up manually.

5. **Kernel Buffering**: Network stack buffering can affect the perceived bandwidth limit, especially for short-lived connections or bursty traffic.

6. **Protocol Support**: The current implementation uses `protocol ip` filters, which only matches IPv4. IPv6 traffic is not limited.

## Advanced Topics

### Integration with External tc Configuration

The NetworkLimiter can coexist with external tc configurations:
- Uses handle `1:` for HTB qdisc (standard)
- Class IDs are based on the cgroup classID
- Filters use cgroup matching, which is independent of other filter types

### Debugging

Enable verbose output for tc commands by checking the output in case of errors. The error messages include the full tc command output.

### Performance Impact

- **Egress**: Minimal overhead, standard HTB performance
- **Ingress**: Additional packet processing through IFB device, slight overhead for redirection
- HTB is a well-optimized queuing discipline suitable for production use

## References

- [Linux Traffic Control HOWTO](https://tldp.org/HOWTO/Traffic-Control-HOWTO/)
- [HTB Linux Queuing Discipline Manual](http://luxik.cdi.cz/~devik/qos/htb/)
- [IFB Device Documentation](https://www.kernel.org/doc/Documentation/networking/ifb.txt)
- [Cgroups v2 Network Controller](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#network)

package limiter

import (
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pmarchini/giogo/internal/utils"
)

// IOLimiter custom error
type IOLimiterError struct {
	Message string
	Cause   error
}

func (e *IOLimiterError) Error() string {
	return e.Message
}

func (e *IOLimiterError) Is(target error) bool {
	return e.Cause == target
}

// chain the errors
func (e *IOLimiterError) Unwrap() error {
	return e.Cause
}

// Parse error
var IOErrUnparsableValue = &IOLimiterError{Message: "unparsable value", Cause: nil}

type IOLimiter struct {
	// Limit is the maximum number of bytes that can be read or written
	ReadThrottle, WriteThrottle uint64
	systemBlockDir              string
	BlockDevices                []BlockDevice
}

// BlockDevice struct holds the information about a block device
type BlockDevice struct {
	Name  string
	Major int64
	Minor int64
}

// GetBlockDevices retrieves all block devices along with their major and minor numbers
func GetBlockDevices(blockDir string) ([]BlockDevice, error) {
	var devices []BlockDevice

	// Read the block directory to get the list of block devices
	entries, err := ioutil.ReadDir(blockDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Construct the path to the dev file, which holds the major:minor numbers
		devFilePath := filepath.Join(blockDir, entry.Name(), "dev")

		// Read the content of the dev file (it contains major:minor numbers)
		devFileContent, err := ioutil.ReadFile(devFilePath)
		if err != nil {
			return nil, err
		}

		// Split the major and minor numbers from the content
		devParts := strings.Split(strings.TrimSpace(string(devFileContent)), ":")
		if len(devParts) != 2 {
			return nil, fmt.Errorf("unexpected format in %s", devFilePath)
		}

		major, err := strconv.ParseInt(devParts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		minor, err := strconv.ParseInt(devParts[1], 10, 64)
		if err != nil {
			return nil, err
		}

		// Create a BlockDevice and append to the list
		devices = append(devices, BlockDevice{
			Name:  entry.Name(),
			Major: major,
			Minor: minor,
		})
	}

	return devices, nil
}

func (i *IOLimiter) Apply(resources *specs.LinuxResources) {
	// Set the throttle values for read and write operations
	var linuxThrottleDevices []specs.LinuxThrottleDevice
	for _, device := range i.BlockDevices {
		linuxThrottleDevices = append(
			linuxThrottleDevices,
			specs.LinuxThrottleDevice{
				LinuxBlockIODevice: specs.LinuxBlockIODevice{
					Major: device.Major,
					Minor: device.Minor,
				},
				Rate: i.ReadThrottle,
			},
		)
	}
	resources.BlockIO = &specs.LinuxBlockIO{}
	if i.ReadThrottle != math.MaxUint64 {
		resources.BlockIO.ThrottleReadBpsDevice = linuxThrottleDevices
	}
	if i.WriteThrottle != math.MaxUint64 {
		resources.BlockIO.ThrottleWriteBpsDevice = linuxThrottleDevices
	}
}

type IOLimiterInitializer struct {
	ReadThrottle, WriteThrottle string
	OverrideSystemBlockDir      string
}

func NewIOLimiter(init *IOLimiterInitializer) (*IOLimiter, error) {
	var systemBlockDir string
	var readThrottle, writeThrottle uint64 = math.MaxUint64, math.MaxUint64
	if init.OverrideSystemBlockDir != "" {
		systemBlockDir = init.OverrideSystemBlockDir
	} else {
		systemBlockDir = "/sys/block"
	}
	// Get the list of block devices
	blockDevices, err := GetBlockDevices(systemBlockDir)
	if err != nil {
		return nil, &IOLimiterError{Message: "error retrieving block devices", Cause: err}
	}
	if init.ReadThrottle != "-1" {
		readThrottle, err = utils.BytesStringToBytes(init.ReadThrottle)
		if err != nil {
			return nil, &IOLimiterError{Message: "unparsable ReadThrottle value", Cause: err}
		}
	}
	if init.WriteThrottle != "-1" {
		writeThrottle, err = utils.BytesStringToBytes(init.WriteThrottle)
		if err != nil {
			return nil, &IOLimiterError{Message: "unparsable WriteThrottle value", Cause: err}
		}
	}
	return &IOLimiter{
		ReadThrottle:   readThrottle,
		WriteThrottle:  writeThrottle,
		systemBlockDir: systemBlockDir,
		BlockDevices:   blockDevices,
	}, nil
}

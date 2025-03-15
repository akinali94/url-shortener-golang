package idgenerator

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	epoch uint64 = 1577836800000

	datacenterBits uint8 = 5
	machineBits    uint8 = 5
	sequenceBits   uint8 = 12

	maxDatacenterID uint64 = -1 ^ (-1 << datacenterBits)
	maxMachineID    uint64 = -1 ^ (-1 << machineBits)
	maxSequence     uint64 = -1 ^ (-1 << sequenceBits)

	datacenterShift uint8 = sequenceBits + machineBits
	machineShift    uint8 = sequenceBits
	timestampShift  uint8 = datacenterBits + machineBits + sequenceBits
)

type Snowflake struct {
	mu            sync.Mutex
	lastTimestamp uint64
	sequence      uint64
	datacenterID  uint64
	machineID     uint64
}

func NewSnowflake(datacenterID, machineID uint64) (*Snowflake, error) {

	if datacenterID < 0 || datacenterID > maxDatacenterID {
		return nil, errors.New("datacenter ID out of range")
	}
	if machineID < 0 || machineID > maxMachineID {
		return nil, errors.New("machine ID out of range")
	}

	return &Snowflake{
		datacenterID:  datacenterID,
		machineID:     machineID,
		lastTimestamp: 0, //-1 or 0 TODO: because of i turned to int64 to uint64 it can not be 1
		sequence:      0,
	}, nil
}

func (s *Snowflake) NextID() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := uint64(time.Now().UnixMilli()) - epoch

	if timestamp < s.lastTimestamp {
		return 0, errors.New("clock moved backwards")
	}

	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// Wait for next millisecond
			for timestamp <= s.lastTimestamp {
				timestamp = uint64(time.Now().UnixMilli()) - epoch
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = timestamp

	id := (timestamp << timestampShift) |
		(s.datacenterID << datacenterShift) |
		(s.machineID << machineShift) |
		s.sequence

	return id, nil
}

func GetDatacenterID() (uint64, error) {
	id, err := strconv.ParseUint(os.Getenv("DATACENTER_ID"), 10, 64)
	if err != nil {
		return 1, errors.New("error on getting Machine ID")
	}
	return id & maxDatacenterID, nil
}

func GetMachineID() (uint64, error) {
	id, err := strconv.ParseUint(os.Getenv("MACHINE_ID"), 10, 64)
	if err != nil {
		return 1, errors.New("error on getting Machine ID")
	}
	return id & maxMachineID, nil
}

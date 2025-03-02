package main

import (
	"sync"
	"errors"
	"time"
	"os"
	"strconv"
)


const (
	epoch int64 = 1577836800000

	datacenterBits uint8 = 5
	machineBits uint8 = 5
	sequenceBits uint8 = 12

	maxDatacenterID int64 = -1 ^ (-1 << datacenterBits)
	maxMachineID int64 = -1 ^ (-1 << machineBits)
	maxSequence int64 = -1 ^ (-1 << sequenceBits)

	datacenterShift uint8 = sequenceBits + machineBits
	machineShift    uint8 = sequenceBits
	timestampShift  uint8 = datacenterBits + machineBits + sequenceBits

)

type Snowflake struct {
	mu            sync.Mutex
	lastTimestamp int64
	sequence      int64
	datacenterID  int64
	machineID     int64
}

func NewSnowflake(datacenterID, machineID int64) (*Snowflake, error) {

	if datacenterID < 0 || datacenterID > maxDatacenterID {
		return nil, errors.New("datacenter ID out of range")
	}
	if machineID < 0 || machineID > maxMachineID {
		return nil, errors.New("machine ID out of range")
	}

	return &Snowflake{
		datacenterID: datacenterID,
		machineID:    machineID,
		lastTimestamp: -1,
		sequence:      0,
	}, nil
}

func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().UnixMilli() - epoch

	if timestamp < s.lastTimestamp {
		return 0, errors.New("clock moved backwards")
	}

	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// Wait for next millisecond
			for timestamp <= s.lastTimestamp {
				timestamp = time.Now().UnixMilli() - epoch
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


func GetDatacenterID() (int64, error) {
    id, err := strconv.ParseInt(os.Getenv("DATACENTER_ID"), 10, 64)
    if err != nil {
        return 1, errors.New("error on getting Machine ID")
    }
    return id & maxDatacenterID, nil
}

func GetMachineID() (int64, error) {
    id, err := strconv.ParseInt(os.Getenv("MACHINE_ID"), 10, 64)
    if err != nil {
        return 1, errors.New("error on getting Machine ID")
    }
    return id & maxMachineID, nil
}
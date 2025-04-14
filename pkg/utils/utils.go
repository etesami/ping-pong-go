package utils

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"time"
)

// calculateRtt calculates the round-trip time (RTT) based on the current time and the ack time
func CalculateRtt(msgSentTime string, msgRecTime string, ackSentTime string, ackRecTime time.Time) (float64, error) {
	msgSentTime1, err1 := StrUnixToTime(msgSentTime)
	msgRecTime1, err2 := StrUnixToTime(msgRecTime)
	ackSentTime1, err3 := StrUnixToTime(ackSentTime)
	if err1 != nil || err2 != nil || err3 != nil {
		return -1, fmt.Errorf("error parsing timestamps: (%v, %v, %v)", err1, err2, err3)
	}
	t1 := msgRecTime1.Sub(msgSentTime1)
	t2 := ackRecTime.Sub(ackSentTime1)
	rtt := float64(t1+t2) / 1000.0
	return rtt, nil
}

func StrUnixToTime(unixStr string) (time.Time, error) {
	unixInt, err := strconv.ParseInt(unixStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse unix time: %v", err)
	}
	return UnixMilliToTime(unixInt), nil
}

// UnixMilliToTime converts a Unix timestamp in milliseconds to a time.Time object
func UnixMilliToTime(unixMilli int64) time.Time {
	return time.Unix(unixMilli/1000, (unixMilli%1000)*int64(time.Millisecond))
}

func GenerateRandomBytes(sizeMB float64) ([]byte, error) {
	size := int(sizeMB * 1024 * 1024)
	randomBytes := make([]byte, size)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

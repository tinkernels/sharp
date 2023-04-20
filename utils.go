package main

import "os"

func PathExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// IterWithControlBit return iterator with control bit
// ctlBit is the init bit to turn back iterator sequence.
func IterWithControlBit[T any](slice []T, ctlBit bool) func() T {
	len_ := len(slice)
	if len_ == 0 {
		return func() (ret T) {
			return
		}
	} else if len_ == 1 {
		return func() T {
			return slice[0]
		}
	}
	ctlBit_ := ctlBit
	valueIdx_ := 0
	if !ctlBit_ {
		valueIdx_ = len_ - 1
	}
	return func() (ret T) {
		ret = slice[valueIdx_]
		if ctlBit_ {
			if valueIdx_ == 0 || valueIdx_+1 < len_ {
				valueIdx_++
			} else if valueIdx_ == len_-1 {
				ctlBit_ = false
			}
		} else {
			if valueIdx_ == len_-1 || valueIdx_-1 >= 0 {
				valueIdx_--
			} else if valueIdx_ == 0 {
				ctlBit_ = true
			}
		}
		return
	}
}

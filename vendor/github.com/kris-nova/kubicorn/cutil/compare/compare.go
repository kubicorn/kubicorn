package compare

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/kris-nova/kubicorn/logger"
)

func IsEqual(actual, expected interface{}) (bool, error) {
	abytes, err := json.Marshal(actual)
	if err != nil {
		return false, fmt.Errorf("Recoverable error comparing JSON: %v", err)
	}
	ebytes, err := json.Marshal(expected)
	if err != nil {
		return false, fmt.Errorf("Recoverable error comparing JSON: %v", err)
	}
	ahash := md5.Sum(abytes)
	ehash := md5.Sum(ebytes)
	logger.Debug("Actual   : %x", ahash)
	logger.Debug("Expected : %x", ehash)
	alen := len(abytes)
	blen := len(ebytes)
	if alen != blen {
		return false, nil
	}
	for i := 0; i < alen; i++ {
		if abytes[i] != ebytes[i] {
			return false, nil
		}
	}
	return true, nil
}

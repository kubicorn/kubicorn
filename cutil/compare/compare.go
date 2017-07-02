package compare

import (
	"crypto/md5"
	"encoding/json"
	"github.com/kris-nova/kubicorn/logger"
)

func Compare(a, b interface{}) bool {

	abytes, err := json.Marshal(a)
	if err != nil {
		logger.Warning("Recoverable error comparing JSON: %v", err)
		return false
	}
	bbytes, err := json.Marshal(b)
	if err != nil {
		logger.Warning("Recoverable error comparing JSON: %v", err)
		return false
	}

	ahash := md5.Sum(abytes)
	bhash := md5.Sum(bbytes)
	logger.Debug("A: %x", ahash)
	logger.Debug("B: %x", bhash)

	alen := len(abytes)
	blen := len(bbytes)
	if alen != blen {
		return false
	}
	for i := 0; i < alen; i++ {
		if abytes[i] != bbytes[i] {
			return false
		}
	}
	return true
}

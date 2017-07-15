package bootstrap

import "strings"

func Inject(data []byte, values map[string]string) ([]byte, error) {
	strData := string(data)
	for k, v := range values {
		strData = strings.Replace(strData, k, v, 1)
	}
	return []byte(strData), nil
}


package profiles

import (
	"encoding/json"
	"fmt"
	"github.com/kris-nova/kubicorn/apis/cluster"
)

func SerializeProviderConfig(config interface{}) (string, error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("unable to marshal provider config: %v", err)
	}
	return string(bytes), nil
}

func DeserializeProviderConfig(config string) (*cluster.Cluster, error) {
	cluster := &cluster.Cluster{}
	err := json.Unmarshal([]byte(config), cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal provider config: %v", err)
	}
	return cluster, nil
}
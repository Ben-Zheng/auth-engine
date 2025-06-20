package json

import "k8s.io/apimachinery/pkg/util/json"

func UnmarshalJSON[T any](data []byte) (T, error) {
	var obj T
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

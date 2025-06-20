package dao

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"golang.org/x/xerrors"
)

// JSON defined JSON data type, need to implements driver.Valuer, sql.Scanner interface.
type JSON json.RawMessage

// Value return json value, implement driver.Valuer interface.
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	bytes, err := json.RawMessage(j).MarshalJSON()
	return string(bytes), err
}

// Scan scan value into Jsonb, implements sql.Scanner interface.
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return xerrors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}
	return nil
}

// MarshalJSON to output non base64 encoded []byte.
func (j JSON) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON to deserialize []byte.
func (j *JSON) UnmarshalJSON(b []byte) error {
	result := json.RawMessage{}
	err := result.UnmarshalJSON(b)
	*j = JSON(result)
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}
	return nil
}

func (j JSON) String() string {
	return string(j)
}

// GormDataType gorm common data type.
func (JSON) GormDataType() string {
	return "json"
}

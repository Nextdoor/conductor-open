package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

// Custom type for JSON formatting.
// Implements beego 'fielder' interface.
type Time struct {
	Value time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var timeStr string
	err := json.Unmarshal(data, &timeStr)
	if err != nil {
		return err
	}
	if timeStr == "" {
		t.Value = time.Time{}
	} else {
		value, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return err
		}
		t.Value = value
	}
	return nil
}

func (t Time) HasValue() bool {
	zeroTime := time.Time{}
	return !t.Value.Equal(zeroTime)
}

func (t Time) Get() *time.Time {
	if t.HasValue() {
		return &t.Value
	}
	return nil
}

func (t Time) String() string {
	if t.HasValue() {
		return fmt.Sprintf(`"%s"`, t.Value.Format(time.RFC3339))
	}
	return "null"
}

func (t Time) FieldType() int {
	return orm.TypeDateTimeField
}
func (t *Time) SetRaw(value interface{}) error {
	if value == nil {
		t.Value = time.Time{}
	} else {
		t.Value = value.(time.Time)
	}
	return nil
}

func (t Time) RawValue() interface{} {
	if t.HasValue() {
		return t.Value
	} else {
		return nil
	}
}

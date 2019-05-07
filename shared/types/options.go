package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/xeipuuv/gojsonschema"

	"github.com/Nextdoor/conductor/shared/logger"
)

type Options struct {
	// CloseTime is when trains should be automatically closed.
	// This is defined as an array of TimeIntervals.
	// Example: M-F 9-5.
	//  []TimeInterval{
	//      TimeInterval{
	//          Every: []time.Weekday{
	//              time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
	//          StartTime: Clock{Hour: 9, Minute: 0},
	//          EndTime: Clock{Hour: 17, Minute: 0},
	//      },
	//  }
	CloseTime            RepeatingTimeIntervals  `json:"close_time"`
	ValidationError      error                   `orm:"-" json:"-"`
	InvalidOptionsString string                  `orm:"-" json:"-"`
}

// Implement beego Fielder interface to handle serialization and deserialization.
func (o Options) String() string {
	b, err := json.Marshal(o)
	if err != nil {
		logger.Error("Error marshalling options: %v", err.Error())
		return ""
	}
	return string(b)
}

func (o Options) FieldType() int {
	return orm.TypeTextField
}

func (o *Options) SetRaw(raw interface{}) error {
	optionsString := raw.(string)

	// The orm panics if this returns an error. Invalid options are demarcated using the ValidationError field.
	err := o.FromString(optionsString)
	if err != nil {
		o.CloseTime = nil
		o.ValidationError = err
		o.InvalidOptionsString = optionsString
	}

	// Never return the error to prevent the orm from panicing.
	return nil
}

func (o Options) RawValue() interface{} {
	return o.String()
}

func (o *Options) FromString(optionsString string) error {
	err := validateOptionsString(optionsString)
	if err != nil {
		return fmt.Errorf("Options validation error: %v", err.Error())
	}

	err = json.Unmarshal([]byte(optionsString), o)
	if err != nil {
		return err
	}

	return nil
}

func (o Options) InCloseTime() bool {
	now := time.Now()
	for _, interval := range o.CloseTime {
		if interval.Includes(now) {
			return true
		}
	}
	return false
}

func (o Options) CloseTimeOverlap(start time.Time, end time.Time) time.Duration {
	return o.CloseTime.TotalOverlap(start, end)
}

// Default is M-F 9-5. Hours are in UTC.
var defaultCloseTime = RepeatingTimeIntervals{
	{
		Every: []time.Weekday{
			time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime: Clock{9, 0},
		EndTime:   Clock{17, 0},
	},
}

var DefaultOptions = Options{CloseTime: defaultCloseTime}

// Used for JSON Schema validation.
// Update this when the structure of Options changes.
// Note: On redeploy, options will be replaced by defaults!
const optionsSchema = `
{
	"type": "object",
	"properties": {
		"close_time": {
			"type": "array",
			"minItems": 1,
			"items": {
				"type": "object",
				"properties": {
					"every": {
						"type": "array",
						"minItems": 1,
						"items": { "type": "number" }
					},
					"start_time": {
						"type": "object",
						"properties": {
							"hour": { "type": "number" },
							"minute": { "type": "number" }
						},
						"required": ["hour", "minute"]
					},
					"end_time": {
						"type": "object",
						"properties": {
							"hour": { "type": "number" },
							"minute": { "type": "number" }
						},
						"required": ["hour", "minute"]
					}
				},
				"required": ["every", "start_time", "end_time"]
			}
		}
	},
	"required": ["close_time"]
}
`

func validateOptionsString(optionsString string) error {
	schemaLoader := gojsonschema.NewStringLoader(optionsSchema)
	optionsLoader := gojsonschema.NewStringLoader(optionsString)
	result, err := gojsonschema.Validate(schemaLoader, optionsLoader)
	if err != nil {
		return err
	}
	return getValidationError(result)
}

func getValidationError(result *gojsonschema.Result) error {
	if result.Valid() {
		return nil
	}
	errString := ""
	validationErrors := result.Errors()
	for i, err := range validationErrors {
		errString += err.String()
		if i < len(validationErrors)-1 {
			errString += "; "
		}
	}
	return fmt.Errorf("%s", errString)
}

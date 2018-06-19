package pipeline

type DataPoint struct {
	TenantID string
	ShapeID  string
	Action   DataPointAction        `json:"action"`         // The action for the dataPoint
	Meta     map[string]string      `json:"meta,omitempty"` // An optional map of strings for sending metadata
	Data     map[string]interface{} `json:"data"`           // The data being sent through the pipe
}

type DataPointAction string

const (
	// DataPointUpsert respresnts an upsert action
	DataPointUpsert DataPointAction = "upsert"
	// DataPointDelete represents a delete action
	DataPointDelete DataPointAction = "delete"

	// DataPointStartPublish represents the start of a publish operation.
	// The Data field will be a map of PropertyDefinition.ID to PropertyDefinition.Name
	// The subscriber should drop and recreate the destination table when this
	// action is received
	DataPointStartPublish DataPointAction = "start-publish"

	// DataPointEndPublish represents the normal end of a publish operation.
	DataPointEndPublish DataPointAction = "end-publish"

	// DataPointAbendPublish represents an unexpected end of a publish operation.
	DataPointAbendPublish DataPointAction = "abend"

	// DataPointMalformed represents a data point which should be logged but which
	// the subscriber should not attempt to insert into the target database.
	DataPointMalformed DataPointAction = "malformed"

	// DataPointSample represents a data point which was created as part
	// of a discovery or editing process and should not be inserted into
	// the destination table.
	DataPointSample DataPointAction = "sample"
)

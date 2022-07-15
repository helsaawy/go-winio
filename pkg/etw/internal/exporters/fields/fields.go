package fields

/* ETW field names
copied from
https://github.com/open-telemetry/opentelemetry-cpp/blob/c22463384db4149c6e0cfae3a9b7336ebf2b3019/exporters/etw/include/opentelemetry/exporters/etw/etw_fields.h#L1

Configurable field names:
 Version                      - Schema version (optional for ETW exporter).
 _name                        - Built-in ETW name at envelope level (dedicated ETW field).
 _time                        - Built-in ETW time at envelope level (dedicated ETW field).
 SpanId                       - OT SpanId
 TraceId                      - OT TraceId
 StartTime                    - OT Span start time
 Kind                         - OT Span kind
 Name                         - OT Span name in ETW 'Payload["Name"]'
 ParentId                     - OT Span parentId
 Links                        - OT Span links array

Reserved names:
 Level                        - a 1-byte integer that enables filtering based on the severity or verbosity of events
 ProviderGuid                 - ETW Provider Guid
 ProviderName                 - ETW Provider Name
 OpcodeName                   - Name of Opcode (e.g. Start, Stop)
 KeywordName                  - Name of Keyword
 TaskName                     - TaskName, could be handled as an alias to Payload['name']
 ChannelName                  - ETW Channel Name
 EventMessage                 - ETW Event Message string for unstructured events
 ActivityId                   - ActivityId for EventSource parenting (current event)
 RelatedActivityId            - RelatedActivityId for EventSource parenting (parent event)
 Pid                          - Process Id
 Tid                          - Thread Id
*/

const (
	KeywordReserved uint64 = 0xFFFF000000000000
	KeywordGeneral  uint64 = 0x0000FFFFFFFFFFFF

	Version = "Version" // Event version
	Type    = "Type"    // Event type
	Name    = "_name"   // Event name
	Time    = "_time"   // Event time
	OpCode  = "OpCode"  // OpCode for TraceLogging

	TraceID      = "TraceId"  // Trace Id
	SpanID       = "SpanId"   // Span Id
	SpanParentID = "ParentId" // Span ParentId
	SpanKind     = "Kind"     // Span Kind
	SpanLinks    = "Links"    // Span Links array

	PayloadName = "Name" // ETW Payload["Name"]

	// Span option
	StartTime     = "StartTime"     // Operation start time
	EndTime       = "EndTime"       // Operation start time
	Duration      = "Duration"      // Operation duration
	StatusCode    = "StatusCode"    // Span status code
	StatusMessage = "StatusMessage" // Span status message
	Success       = "Success"       // Span success
	Timestamp     = "Timestamp"     // Log timestamp

	// Span value
	SpanEventName = "Span"      // ETW event name for Span
	SpanStart     = "SpanStart" // ETW for Span Start
	SpanEnd       = "SpanEnd"   // ETW for Span Start

	// Log value
	LogEventName    = "Log"            // ETW event name for Log
	LogBody         = "body"           // Log body
	LogSeverityText = "severityText"   // Sev text
	LogSeverityNum  = "severityNumber" // Sev num
)

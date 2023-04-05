package geneva

// copied from
// https://github.com/microsoft/opentelemetry-cpp-exporters/blob/main/exporters/etw/include/opentelemetry/exporters/geneva/geneva_fields.h

/*
  List of configurable Field Name constants:

  - env_ver                      - Schema version (optional for ETW exporter).
  - env_name                     - Built-in ETW name at envelope level (dedicated ETW field).
  - env_time                     - Built-in ETW time at envelope level.
  - env_dt_spanId                - OT SpanId
  - env_dt_traceId               - OT TraceId
  - startTime                    - OT Span start time
  - kind                         - OT Span kind
  - name                         - OT Span name in ETW 'Payload["name"]'
  - parentId                     - OT Span parentId
  - links                        - OT Span links array

  Other standard fields (reserved names) that may be appended by ETW channel:

  - Level
  - ProviderGuid
  - ProviderName
  - OpcodeName
  - KeywordName
  - TaskName
  - ChannelName
  - EventMessage
  - ActivityId
  - Pid
  - Tid
*/

// todo: make this a struct that can be modified so that etw and geneva exporters can share the same code

//nolint:unused
const (
	fieldVersion = "env_ver"  // Event version
	fieldType    = "env_type" // Event type
	fieldName    = "env_name" // Event name

	fieldTime = "env_time" // Event time at envelope

	fieldOpCode = "env_opcode" // OpCode for TraceLogging

	fieldTraceID      = "env_dt_traceId" // Trace Id
	fieldSpanID       = "env_dt_spanId"  // Span Id
	fieldSpanParentID = "parentId"       // Span ParentId
	fieldSpanKind     = "kind"           // Span Kind
	fieldSpanLinks    = "links"          // Span Links array

	fieldPayloadName = "name" // ETW Payload["name"]

	// Span option constants

	fieldStartTime     = "startTime"     // Operation start time
	fieldEndTime       = "env_time"      // Operation end time
	fieldDuration      = "duration"      // Operation duration
	fieldStatuscode    = "statusCode"    // OT Span status code
	fieldStatusmessage = "statusMessage" // OT Span status message
	fieldSuccess       = "success"       // OT Span success
	fieldTimestamp     = "Timestamp"     // Log timestamp

	fieldClientReqID = "clientRequestId"
	fieldCorrelReqID = "correlationRequestId"

	// Value constants

	valueSpan = "Span" // ETW event name for Span
	valueLog  = "Log"  // ETW event name for Log

	valueSpanStart = "SpanStart" // ETW for Span Start
	valueSpanEnd   = "SpanEnd"   // ETW for Span Start

	fieldLogBody         = "body"           // Log body
	fieldLogSeverityText = "severityText"   // Sev text
	fieldLogSeverityNum  = "severityNumber" // Sev num
)

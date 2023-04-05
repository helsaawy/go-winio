package etw

// copied from
// https://github.dev/open-telemetry/opentelemetry-cpp/blob/main/exporters/etw/include/opentelemetry/exporters/etw/etw_fields.h

/*
  List of configurable Field Name constants:

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

  Other standard fields (reserved names) that may be appended by ETW channel:

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

  Example "Span" as shown in Visual Studio "Diagnostic Events" view. EventName="Span":

    {
      "Timestamp": "2021-04-01T00:33:25.5876605-07:00",
      "ProviderName": "OpenTelemetry-ETW-TLD",
      "Id": 20,
      "Message": null,
      "ProcessId": 10424,
      "Level": "Always",
      "Keywords": "0x0000000000000000",
      "EventName": "Span",
      "ActivityID": "56f2366b-5475-496f-0000-000000000000",
      "RelatedActivityID": null,
      "Payload": {
        "Duration": 0,
        "Name": "B.max",
        "ParentId": "8ad900d0587fad4a",
        "SpanId": "6b36f25675546f49",
        "StartTime": "2021-04-01T07:33:25.587Z",
        "TraceId": "8f8ac710c37c5a419f0fe574f335e986"
      }
    }

  Example named Event on Span. Note that EventName="MyEvent2" in this case:

    {
      "Timestamp": "2021-04-01T00:33:22.5848789-07:00",
      "ProviderName": "OpenTelemetry-ETW-TLD",
      "Id": 15,
      "Message": null,
      "ProcessId": 10424,
      "Level": "Always",
      "Keywords": "0x0000000000000000",
      "EventName": "MyEvent2",
      "ActivityID": null,
      "RelatedActivityID": null,
      "Payload": {
        "SpanId": "0da9f6bf7524a449",
        "TraceId": "7715c9d490f54f44a5d0c6b62570f1b2",
        "strKey": "anotherValue",
        "uint32Key": 9876,
        "uint64Key": 987654321
      }
    }

*/

//nolint:unused
const (
	fieldVersion = "Version" // Event version
	fieldType    = "Type"    // Event type
	fieldName    = "_name"   // Event name
	fieldTime    = "_time"   // Event time
	fieldOpCode  = "OpCode"  // OpCode for TraceLogging

	fieldTraceID      = "TraceId"  // Trace Id
	fieldSpanID       = "SpanId"   // Span Id
	fieldSpanParentID = "ParentId" // Span ParentId
	fieldSpanKind     = "Kind"     // Span Kind
	fieldSpanLinks    = "Links"    // Span Links array

	fieldPayloadName = "Name" // ETW Payload["Name"]

	// Span option constants

	fieldStartTime     = "StartTime"     // Operation start time
	fieldDuration      = "Duration"      // Operation duration
	fieldStatusCode    = "StatusCode"    // Span status code
	fieldStatusMessage = "StatusMessage" // Span status message
	fieldSuccess       = "Success"       // Span success
	fieldTimeStamp     = "Timestamp"     // Log timestamp

	// Value constants

	valueSpan = "Span" // ETW event name for Span
	valueLog  = "Log"  // ETW event name for Log

	valueSpanStart = "SpanStart" // ETW for Span Start
	valueSpanEnd   = "SpanEnd"   // ETW for Span Start

	fieldEnvProperties = "env_properties" // ETW event_properties with JSON string

	// Log specific

	fieldLogBody         = "body"           // Log body
	fieldLogSeverityText = "severityText"   // Sev text
	fieldLogSeverityNum  = "severityNumber" // Sev num
)

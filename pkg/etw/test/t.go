package main

import (
	"fmt"
	"log"

	"github.com/Microsoft/go-winio/pkg/etw/exporters/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	s := "6b36f25675546f49"
	id, err := trace.SpanIDFromHex(s)
	if err != nil {
		log.Fatalf("span id: %v", err)
	}
	fmt.Printf("%q\n%q \n", s, id.String())
	g := otel.SpanIDToGUID(id)
	fmt.Println(g.String())
	return
}

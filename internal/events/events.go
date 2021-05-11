/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/onflow/cadence"
	jsonCadence "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "events",
	Short:            "Utilities to read events",
	TraverseChildren: true,
}

func init() {
	GetCommand.AddToParent(Cmd)
}

// EventResult result structure
type EventResult struct {
	BlockEvents []client.BlockEvents
	Events      []flow.Event
}

// JSON convert result to JSON
func (e *EventResult) JSON() interface{} {
	result := make([]interface{}, 0)

	for _, blockEvent := range e.BlockEvents {
		if len(blockEvent.Events) > 0 {
			for _, event := range blockEvent.Events {
				result = append(result, map[string]interface{}{
					"blockID":       blockEvent.Height,
					"index":         event.EventIndex,
					"type":          event.Type,
					"transactionId": event.TransactionID.String(),
					"values": json.RawMessage(
						jsonCadence.MustEncode(event.Value),
					),
				})
			}
		}
	}

	return result
}

// String convert result to string
func (e *EventResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	for _, blockEvent := range e.BlockEvents {
		if len(blockEvent.Events) > 0 {
			_, _ = fmt.Fprintf(writer, "Events Block #%v:", blockEvent.Height)
			eventsString(writer, blockEvent.Events)
			_, _ = fmt.Fprintf(writer, "\n")
		}
	}

	// if we have events passed directly and not in relation to block
	eventsString(writer, e.Events)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (e *EventResult) Oneliner() string {
	result := ""
	for _, blockEvent := range e.BlockEvents {
		if len(blockEvent.Events) > 0 {
			result += fmt.Sprintf("Events Block #%v: [", blockEvent.Height)
			for _, event := range blockEvent.Events {
				result += fmt.Sprintf(
					"Index: %v, Type: %v, TxID: %s, Value: %v",
					event.EventIndex, event.Type, event.TransactionID, event.Value,
				)
			}
			result += "] "
		}
	}

	return result
}

func eventsString(writer io.Writer, events []flow.Event) {
	for _, event := range events {
		eventString(writer, event)
	}
}

func eventString(writer io.Writer, event flow.Event) {
	_, _ = fmt.Fprintf(writer, "\n    Index\t%d\n", event.EventIndex)
	_, _ = fmt.Fprintf(writer, "    Type\t%s\n", event.Type)
	_, _ = fmt.Fprintf(writer, "    Tx ID\t%s\n", event.TransactionID)
	_, _ = fmt.Fprintf(writer, "    Values\n")

	for i, field := range event.Value.EventType.Fields {
		value := event.Value.Fields[i]
		printField(writer, field, value)
	}
}

func printField(writer io.Writer, field cadence.Field, value cadence.Value) {
	v := value.ToGoValue()
	typeInfo := "Unknown"

	if field.Type != nil {
		typeInfo = field.Type.ID()
	} else if _, isAddress := v.([8]byte); isAddress {
		typeInfo = "Address"
	}

	fmt.Fprintf(writer, "\t\t-")
	fmt.Fprintf(writer, " %s (%s):\t", field.Identifier, typeInfo)
	// Try the two most obvious cases
	if address, ok := v.([8]byte); ok {
		fmt.Fprintf(writer, "%x", address)
	} else if util.IsByteSlice(v) || field.Identifier == "publicKey" {
		// make exception for public key, since it get's interpreted as []*big.Int
		for _, b := range v.([]interface{}) {
			fmt.Fprintf(writer, "%x", b)
		}
	} else if uintVal, ok := v.(uint64); typeInfo == "UFix64" && ok {
		fmt.Fprintf(writer, "%v", cadence.UFix64(uintVal))
	} else {
		fmt.Fprintf(writer, "%v", v)
	}
	fmt.Fprintf(writer, "\n")
}

package mdata_payload

import (
	"fmt"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"reflect"
	"strconv"
	"strings"
)

type MdPayload struct {
	Action     string
	Gtin       string
	Attributes []string
	State      string
}

func (p MdPayload) invaildChar() (bool, string) {
	//Returns true if the payload contains "|", indicating an invalid payload
	values := reflect.ValueOf(p)
	num := values.NumField()

	for i := 0; i < num; i++ {
		v := values.Field(i)
		if v.Kind() == reflect.String {
			strv := v.String()
			if strings.Contains(strv, "|") {
				return true, strv
			}
		}
		if v.Kind() == reflect.Slice {
			for _, value := range v.Interface().([]string) {
				if strings.Contains(value, "|") {
					return true, value
				}
			}
		}
	}

	return false, ""
}

func (p *MdPayload) invalidAttributes() bool {
	//Return false for empty attributes
	if len(p.Attributes) == 1 && p.Attributes[0] == "" {
		return false
	}

	// Verify they are key=value pairs in the slice of string
	for _, pair := range p.Attributes {
		if strings.Count(pair, "=") != 1 {
			return true
		}
	}

	return false
}

func (p *MdPayload) invalidGtin() bool {
	// Verify the length of GTIN is 14 integers (no symbols, no letters)
	_, err := strconv.Atoi(p.Gtin)
	if err != nil {
		// Error converting string to int; invalid
		return true
	}

	if len(p.Gtin) != 14 {
		return true
	}

	return false
}

func (p *MdPayload) invalidState() bool {
	// Verify the state setting is valid: one of ACTIVE, INACTIVE, DISCONTINUED
	validStates := []string{"ACTIVE", "INACTIVE", "DISCONTINUED"}
	var foundMatch bool
	for _, state := range validStates {
		if p.State == state {
			foundMatch = true
		}
		if foundMatch {
			return false
		}
	}
	return true
}

func FromBytes(payloadData []byte) (*MdPayload, error) {
	if payloadData == nil {
		return nil, &processor.InvalidTransactionError{Msg: "Must contain payload"}
	}
	/*
		Sample Payload
		action,gtin,key=value,key=value,


	*/
	parts := strings.Split(string(payloadData), ",")
	if len(parts) < 4 { //Client will always send a payload with at least 4 parts: "%v,%v,%v,%v
		return nil, &processor.InvalidTransactionError{Msg: "Payload is malformed"}
	}

	payload := MdPayload{}
	payload.Action = parts[0]
	payload.Gtin = parts[1]
	payload.Attributes = parts[2 : len(parts)-1]
	payload.State = parts[len(parts)-1]

	if len(payload.Action) < 1 {
		return nil, &processor.InvalidTransactionError{Msg: "Action is required"}
	}

	if payload.invalidGtin() {
		return nil, &processor.InvalidTransactionError{Msg: "Gtin-14 is required"}
	}

	if payload.Action == "update" {
		if len(payload.Attributes) < 1 || payload.Attributes[0] == "" {
			return nil, &processor.InvalidTransactionError{Msg: "Attributes are required for update"}
		}
	}

	if payload.invalidAttributes() {
		return nil, &processor.InvalidTransactionError{
			Msg: fmt.Sprintf("Invalid attributes (attributes must be in key=value pairs): %v", payload.Attributes)}
	}

	if payload.Action == "set" {

		if len(payload.State) < 1 {
			return nil, &processor.InvalidTransactionError{Msg: "State is required to set"}
		}

		if payload.invalidState() {
			return nil, &processor.InvalidTransactionError{
				Msg: fmt.Sprintf("Invalid state (state must be one of ACTIVE, INACTIVE, DISCONTINUED), GOT: %v", payload.State)}
		}
	}

	isInvalid, invalidString := payload.invaildChar()
	if isInvalid {
		return nil, &processor.InvalidTransactionError{
			Msg: fmt.Sprintf("Invalid Name (char '|' not allowed): '%v'", invalidString)}
	}

	return &payload, nil
}

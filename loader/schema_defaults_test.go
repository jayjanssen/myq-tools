package loader

import (
	"testing"
)

func TestSchemaParse(t *testing.T) {
	err := LoadDefaultSchemas()
	if err != nil {
		t.Error(err)
	}
	if len(Schemas) < 1 {
		t.Error("No schemas parsed!")
	}

	// Check the status schema
	status := Schemas[0]
	if status.Name != "status" {
		t.Error("First view is not named `status`")
	}
	if uptimeType, ok := status.Keys[`Uptime`]; ok {
		if uptimeType != INT {
			t.Error("`Uptime` type is not INT")
		}
	} else {
		t.Error("`Uptime` not found")
	}
}

package telemetry

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	_, err := Init(context.Background(), "localhost:4317", "name")
	if err != nil {
		if strings.Contains(err.Error(), "cannot merge resource due to conflicting Schema URL") {
			assert.Fail(t, "mismatch semconv version. https://github.com/open-telemetry/opentelemetry-specification/blob/v1.20.0/specification/resource/sdk.md#merge")
			return
		}
	}
	assert.NoError(t, err)
}

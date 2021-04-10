package prepare_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.xsfx.dev/schnutibox/pkg/prepare"
)

func TestBoxService(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tmpDir := t.TempDir()
	serviceFile := tmpDir + "schnutibox.service"
	err := prepare.BoxService(serviceFile, false)
	assert.NoError(err)
	assert.FileExists(serviceFile)
}

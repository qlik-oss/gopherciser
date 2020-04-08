package scenario

import (
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	someSpace = elasticstructs.Space{
		ID:   "a",
		Name: "MySpace",
	}
)

func TestDestinationSpace_GetSpaceByName(t *testing.T) {
	am := session.NewAppMap()
	spaces := make([]elasticstructs.Space, 1)
	spaces = append(spaces, someSpace)
	am.FillSpaces(spaces)
	sessionState := &session.State{ArtifactMap: am}
	nameRequest := DestinationSpace{DestinationSpaceName: "MySpace"}
	space, err := nameRequest.ResolveDestinationSpace(sessionState)
	assert.NoError(t, err)
	assert.Equal(t, "a", space.ID)
}

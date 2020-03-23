package scenario

import (
	"testing"

	"github.com/qlik-oss/gopherciser/session"
)

func TestElasticExploreSettings_Validate(t *testing.T) {
	explore := ElasticExploreSettings{}

	if err := explore.Validate(); err != nil {
		t.Errorf("Validate explore action (Default) failed: %v", err)
	}

	explore.CollectionNames = append(explore.CollectionNames, "tag1", "tag2")
	if err := explore.Validate(); err != nil {
		t.Errorf("Validate explore action (CollectionNames) failed: %v", err)
	}

	explore.CollectionIds = append(explore.CollectionIds, "tagID1", "tagID2")
	if err := explore.Validate(); err != nil {
		t.Errorf("Validate explore action (CollectionNames+CollectionIds) failed: %v", err)
	}

	explore.CollectionNames = nil
	if err := explore.Validate(); err != nil {
		t.Errorf("Validate explore action (CollectionIds) failed: %v", err)
	}

	explore.SpaceId = "spaceID"
	if err := explore.Validate(); err != nil {
		t.Errorf("Validate explore action (SpaceID) failed: %v", err)
	}

	spaceName, errTemplate := session.NewSyncedTemplate("spaceName")
	if errTemplate != nil {
		t.Fatal(errTemplate)
	}
	explore.SpaceName = *spaceName

	if err := explore.Validate(); err == nil {
		t.Errorf("Validate explore action (SpaceID+SpaceName) failed: %v", err)
	}

	explore.SpaceId = ""
	if err := explore.Validate(); err != nil {
		t.Errorf("Validate explore action (SpaceName) failed: %v", err)
	}
}

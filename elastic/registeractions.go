package elastic

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/qlik-oss/gopherciser/scenario"
)

const (
	ActionElasticOpenHub          = "elasticopenhub"
	ActionElasticReload           = "elasticreload"
	ActionElasticUploadApp        = "elasticuploadapp"
	ActionElasticCreateCollection = "elasticcreatecollection"
	ActionElasticDeleteCollection = "elasticdeletecollection"
	ActionElasticHubSearch        = "elastichubsearch"
	ActionElasticDeleteApp        = "elasticdeleteapp"
	ActionElasticCreateApp        = "elasticcreateapp"
	ActionElasticExportApp        = "elasticexportapp"
	ActionElasticGenerateOdag     = "elasticgenerateodag"
	ActionElasticDeleteOdag       = "elasticdeleteodag"
	ActionElasticDuplicateApp     = "elasticduplicateapp"
	ActionElasticExplore          = "elasticexplore"
	ActionElasticMoveApp          = "elasticmoveapp"
	ActionElasticPublishApp       = "elasticpublishapp"
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	err := scenario.RegisterActions(
		map[string]scenario.ActionSettings{
			ActionElasticOpenHub:          ElasticOpenHubSettings{},
			ActionElasticReload:           ElasticReloadSettings{},
			ActionElasticUploadApp:        ElasticUploadAppSettings{},
			ActionElasticCreateCollection: ElasticCreateCollectionSettings{},
			ActionElasticDeleteCollection: ElasticDeleteCollectionSettings{},
			ActionElasticHubSearch:        ElasticHubSearchSettings{},
			ActionElasticDeleteApp:        ElasticDeleteAppSettings{},
			ActionElasticCreateApp:        ElasticCreateAppSettings{},
			ActionElasticExportApp:        ElasticExportAppSettings{},
			ActionElasticGenerateOdag:     ElasticGenerateOdagSettings{},
			ActionElasticDeleteOdag:       ElasticDeleteOdagSettings{},
			ActionElasticDuplicateApp:     ElasticDuplicateAppSettings{},
			ActionElasticExplore:          ElasticExploreSettings{},
			ActionElasticMoveApp:          ElasticMoveAppSettings{},
			ActionElasticPublishApp:       ElasticPublishAppSettings{},
		})
	if err != nil {
		panic(fmt.Sprintf("failed to register actions:\n %+v", err))
	}

	if err := scenario.RegisterActionsOverride(map[string]scenario.ActionSettings{}); err != nil {
		panic(fmt.Sprintf("failed to register action overide:\n %+v", err))
	}
}

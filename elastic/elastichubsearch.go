package elastic

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/session"
)

const searchResultLimit = "100" //Default from client
const searchCollectionsEndpoint = "api/v1/collections"
const searchAppsEndpoint = "api/v1/items"

type (
	// SearchModeEnum defines what to search for
	SearchModeEnum int

	// QuerySourceEnum defines source of query
	QuerySourceEnum int

	// ElasticHubSearchSettingsCore specify app to reload
	ElasticHubSearchSettingsCore struct {
		SearchMode  SearchModeEnum  `json:"searchfor" displayname:"Search for" doc-key:"elastichubsearch.searchfor"`
		QuerySource QuerySourceEnum `json:"querysource" displayname:"Query source" doc-key:"elastichubsearch.querysource"`
		Query       string          `json:"query" displayname:"Query" doc-key:"elastichubsearch.query"`
		Filename    string          `json:"queryfile" displayname:"Query file" displayelement:"file" doc-key:"elastichubsearch.queryfile"`
	}

	// ElasticHubSearchSettings settings for search
	ElasticHubSearchSettings struct {
		ElasticHubSearchSettingsCore
		queries []string
	}
)

const (
	// Collections search collections
	Collections SearchModeEnum = iota
	// Apps search apps
	Apps
	// Both search both collections and apps
	Both
)

const (
	// QueryString queries from string or file
	QueryString QuerySourceEnum = iota
	// FromFile queries read from file
	FromFile
)

func (value SearchModeEnum) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"collections": int(Collections),
		"apps":        int(Apps),
		"both":        int(Both),
	})
	return enumMap
}

func (value QuerySourceEnum) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"string":   int(QueryString),
		"fromfile": int(FromFile),
	})
	return enumMap
}

// UnmarshalJSON unmarshal ElasticHubSearchSettings
// settings
func (settings *ElasticHubSearchSettings) UnmarshalJSON(arg []byte) error {
	core := ElasticHubSearchSettingsCore{}
	err := jsonit.Unmarshal(arg, &core)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal ElasticHubSearchSettingsCore")
	}
	settings.ElasticHubSearchSettingsCore = core

	if core.QuerySource == FromFile && runtime.GOOS != "js" {
		file, err := os.Open(core.Filename)
		if err != nil {
			return errors.Wrap(err, "Failed to open query file")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			settings.queries = append(settings.queries, scanner.Text())
		}
		if scanner.Err() != nil {
			return errors.Wrap(err, "Failed reading query file")
		}
	}
	return nil
}

// UnmarshalJSON unmarshal SearchModeEnum
func (value *SearchModeEnum) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal SearchModeEnum")
	}

	*value = SearchModeEnum(i)
	return nil
}

// MarshalJSON marshal SearchModeEnum type
func (value SearchModeEnum) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown SearchModeEnum<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// UnmarshalJSON unmarshal QuerySourceEnum
func (value *QuerySourceEnum) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal QuerySourceEnum")
	}

	*value = QuerySourceEnum(i)
	return nil
}

// MarshalJSON marshal SearchModeEnum type
func (value QuerySourceEnum) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown QuerySourceEnum<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate EfeHubSearch action (Implements ActionSettings interface)
func (settings ElasticHubSearchSettings) Validate() error {
	if settings.QuerySource == FromFile {
		file, err := os.Open(settings.Filename)
		if err != nil {
			return errors.Wrap(err, "Failed to open query term file")
		}
		defer file.Close()
	}
	return nil
}

// Execute EfeHubSearch action (Implements ActionSettings interface)
func (settings ElasticHubSearchSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	// Todo log selected query in details
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	var query string

	switch settings.QuerySource {
	case QueryString:
		query = settings.Query
	case FromFile:
		n := len(settings.queries)
		if n < 1 {
			actionState.AddErrors(errors.Errorf("No queries specified <%v>", n))
			return
		}
		i := sessionState.Randomizer().Rand(n)
		query = settings.queries[i]
	default:
		actionState.AddErrors(errors.Errorf("Unknown query source <%d>", settings.QuerySource))
		return
	}

	switch settings.SearchMode {
	case Collections:
		searchCollections := searchQuery(host, query, searchCollectionsEndpoint, sessionState, actionState)
		if searchCollections.ResponseStatusCode != http.StatusOK {
			actionState.AddErrors(errors.Errorf("Failed to perform search: %v", searchCollections.ResponseBody))
		}
	case Apps:
		searchApps := searchQuery(host, query, searchAppsEndpoint, sessionState, actionState)
		if searchApps.ResponseStatusCode != http.StatusOK {
			actionState.AddErrors(errors.Errorf("Failed to perform search: %v", searchApps.ResponseBody))
		}
	case Both:
		searchCollections := searchQuery(host, query, searchCollectionsEndpoint, sessionState, actionState)
		if searchCollections.ResponseStatusCode != http.StatusOK {
			actionState.AddErrors(errors.Errorf("Failed to perform search: %v", searchCollections.ResponseBody))
		}
		searchApps := searchQuery(host, query, searchAppsEndpoint, sessionState, actionState)
		if searchApps.ResponseStatusCode != http.StatusOK {
			actionState.AddErrors(errors.Errorf("Failed to perform search: %v", searchApps.ResponseBody))
		}
	default:
		actionState.AddErrors(errors.Errorf("Unknown search mode <%d>", settings.SearchMode))
		return
	}
}

func searchQuery(host string, query string, endpoint string, sessionState *session.State, actionState *action.State) session.RestRequest {
	searchRequest := session.RestRequest{
		Method:      session.GET,
		Destination: fmt.Sprintf("%v/%v?limit=%v&sort=-createdAt&query=%v", host, endpoint, searchResultLimit, query),
	}
	sessionState.Rest.QueueRequest(actionState, true, &searchRequest, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return searchRequest // we had an error
	}
	var searchResult *elasticstructs.SearchResult
	if err := jsonit.Unmarshal(searchRequest.ResponseBody, &searchResult); err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed to unmarshal search result"))
		return searchRequest
	}
	sessionState.LogEntry.LogInfo("hubsearch", fmt.Sprintf("query:<%v> numresults:<%v>", query, len(searchResult.Data)))
	return searchRequest
}

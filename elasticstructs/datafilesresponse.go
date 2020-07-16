package elasticstructs

type DataFilesResp struct {
	Data  []DataFilesRespData `json:"data"`
	Links DataFilesRespLinks  `json:"links"`
}

type DataFilesRespQCredentials struct {
	Ciphertext string `json:"ciphertext"`
}

type DataFilesRespData struct {
	Created              string                    `json:"created"`
	DatasourceID         string                    `json:"datasourceID"`
	ID                   string                    `json:"id"`
	Links                DataFilesRespLinks        `json:"links"`
	Privileges           []string                  `json:"privileges"`
	QArchitecture        int64                     `json:"qArchitecture"`
	QConnectStatement    string                    `json:"qConnectStatement"`
	QCredentials         DataFilesRespQCredentials `json:"qCredentials"`
	QCredentialsID       string                    `json:"qCredentialsID"`
	QEngineObjectID      string                    `json:"qEngineObjectID"`
	QID                  string                    `json:"qID"`
	QLogOn               int64                     `json:"qLogOn"`
	QName                string                    `json:"qName"`
	QSeparateCredentials bool                      `json:"qSeparateCredentials"`
	QType                string                    `json:"qType"`
	Space                string                    `json:"space"`
	Tenant               string                    `json:"tenant"`
	Updated              string                    `json:"updated"`
	User                 string                    `json:"user"`
	Version              string                    `json:"version"`
}

type DataFilesRespLinksHref struct {
	Href string `json:"href"`
}

type DataFilesRespLinks struct {
	Self DataFilesRespLinksHref `json:"self"`
}

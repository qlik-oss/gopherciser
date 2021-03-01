package tempcontent

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/eventials/go-tus"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
)

// Current client default: 300 mb
const (
	defaultChunkSize         int64 = 300 * 1024 * 1024
	tempContentFilesEndpoint       = "api/v1/temp-contents/files"
)

type (
	RemoteFile struct {
		ID  string
		URL string
	}

	tusClient struct {
		tusClient  *tus.Client
		maxRetries int
	}
)

// UploadTempContentFromFile uploads a tempfile with smart selection of
// chunksize. If chunksize is not set (<=0) and file size is smaller than
// default chunksize, the chunksize it will be set to the file size roofed to
// closest 1024 bytes.
func UploadTempContentFromFile(ctx context.Context, sessionState *session.State, connection *connection.ConnectionSettings,
	file *os.File, chunkSize int64, maxRetries int) (*RemoteFile, error) {

	fileStat, err := file.Stat()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	fileSize := fileStat.Size()
	if chunkSize <= 0 && fileSize < defaultChunkSize {
		chunkSize = (fileSize/1024 + 1) * 1024
	}
	tempFileClient, err := newTUSClient(sessionState, connection, chunkSize, maxRetries)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tempFile, err := tempFileClient.uploadFromFile(ctx, file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to upload temp content from file")
	}
	return tempFile, nil
}

func newTUSClient(sessionState *session.State, connection *connection.ConnectionSettings, chunkSize int64, maxRetries int) (*tusClient, error) {
	if maxRetries < 0 {
		maxRetries = 0
	}
	restURL, err := connection.GetRestUrl()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	host, err := connection.GetHost()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// upload file using tus chunked uploads protocol
	tusConfig := tus.DefaultConfig()
	tusConfig.ChunkSize = defaultChunkSize
	if chunkSize > 0 {
		tusConfig.ChunkSize = chunkSize
	}
	tusConfig.Header = sessionState.HeaderJar.GetHeader(host)
	tusConfig.HttpClient = sessionState.Rest.Client

	// upload to temporary storage
	client, err := tus.NewClient(fmt.Sprintf("%s/%s", restURL, tempContentFilesEndpoint), tusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tus client")
	}
	return &tusClient{
		tusClient:  client,
		maxRetries: maxRetries,
	}, nil
}

func (client tusClient) uploadFromFile(ctx context.Context, file *os.File) (*RemoteFile, error) {
	upload, err := tus.NewUploadFromFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tus upload from file")
	}
	uploader, err := client.tusClient.CreateUpload(upload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tus uploader")
	}

	retries := 0
	retryWithBackoff := func() bool {
		if retries < client.maxRetries {
			helpers.WaitFor(ctx, time.Second*time.Duration(retries))
			retries++
			return true
		}
		return false
	}

	for err == nil || err != io.EOF && retryWithBackoff() {
		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "tus upload aborted")
		default:
			err = uploader.UploadChunck()
		}
	}

	if err != io.EOF {
		return nil, errors.Wrap(err, "failed tus upload")
	}

	tempFile := &RemoteFile{
		URL: uploader.Url(),
	}
	tempLocationURL, err := url.Parse(tempFile.URL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse temp content location")
	}
	tempFile.ID = path.Base(tempLocationURL.Path)
	return tempFile, nil
}

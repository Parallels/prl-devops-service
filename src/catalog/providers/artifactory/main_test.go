package artifactory

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/jobs/tracker"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestArtifactoryPullFile_ReportsDownloaderProgressToWorkflow(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	progressService := tracker.NewProgressService(ctx)
	defer progressService.Stop()

	jobID := "job-artifactory-download"
	updates := make(chan []data_models.JobStep, 10)
	progressService.OnUpdateJobProgressAndSteps = func(currentJobID string, percent int, state string, steps []data_models.JobStep) {
		if currentJobID != jobID {
			return
		}
		select {
		case updates <- steps:
		default:
		}
	}
	progressService.RegisterJobWorkflow(jobID, []tracker.JobStep{
		{Name: constants.ActionDownloader, Weight: 100, HasPercentage: true, DisplayName: "Downloading Pack File"},
	})

	content := []byte("test-artifactory-pack-file")
	oldClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case http.MethodHead:
				return &http.Response{
					StatusCode:    http.StatusOK,
					Status:        "200 OK",
					ContentLength: int64(len(content)),
					Header: http.Header{
						"Content-Length": []string{strconv.Itoa(len(content))},
					},
					Body: io.NopCloser(bytes.NewReader(nil)),
				}, nil
			case http.MethodGet:
				return &http.Response{
					StatusCode:    http.StatusOK,
					Status:        "200 OK",
					ContentLength: int64(len(content)),
					Header: http.Header{
						"Content-Length": []string{strconv.Itoa(len(content))},
					},
					Body: io.NopCloser(bytes.NewReader(content)),
				}, nil
			default:
				return &http.Response{
					StatusCode: http.StatusMethodNotAllowed,
					Status:     "405 Method Not Allowed",
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil
			}
		}),
	}
	defer func() {
		http.DefaultClient = oldClient
	}()

	provider := NewArtifactoryProvider()
	provider.Repo = ArtifactoryRepo{
		Host:     "http://example.com",
		RepoName: "repo",
		ApiKey:   "secret",
	}
	provider.SetJobId(jobID)

	tempDir := t.TempDir()
	err := provider.PullFile(ctx, "catalog/path", "sample.pdpack", tempDir)
	require.NoError(t, err)

	downloadedFile := filepath.Join(tempDir, "sample.pdpack")
	data, err := os.ReadFile(downloadedFile)
	require.NoError(t, err)
	require.Equal(t, content, data)

	require.Eventually(t, func() bool {
		for {
			select {
			case steps := <-updates:
				if len(steps) != 1 {
					continue
				}
				step := steps[0]
				if step.Name != constants.ActionDownloader {
					continue
				}
				return step.CurrentPercentage == 100 &&
					step.Filename == "sample.pdpack" &&
					step.Total == int64(len(content)) &&
					step.Value == int64(len(content))
			default:
				return false
			}
		}
	}, 4*time.Second, 100*time.Millisecond)
}

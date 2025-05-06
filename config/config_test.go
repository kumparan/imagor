package config

import (
	"github.com/kumparan/imagor"
	"github.com/kumparan/imagor/imagorpath"
	"github.com/kumparan/imagor/loader/httploader"
	"github.com/kumparan/imagor/metrics/prometheusmetrics"
	"github.com/kumparan/imagor/storage/filestorage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	srv := CreateServer(nil)
	assert.Equal(t, ":8000", srv.Addr)
	app := srv.App.(*imagor.Imagor)

	assert.False(t, app.Debug)
	assert.False(t, app.Unsafe)
	assert.Equal(t, time.Second*30, app.RequestTimeout)
	assert.Equal(t, time.Second*20, app.LoadTimeout)
	assert.Equal(t, time.Second*20, app.SaveTimeout)
	assert.Equal(t, time.Second*20, app.ProcessTimeout)
	assert.Empty(t, app.BasePathRedirect)
	assert.Empty(t, app.ProcessConcurrency)
	assert.Empty(t, app.BaseParams)
	assert.False(t, app.ModifiedTimeCheck)
	assert.False(t, app.AutoWebP)
	assert.False(t, app.AutoAVIF)
	assert.False(t, app.DisableErrorBody)
	assert.False(t, app.DisableParamsEndpoint)
	assert.Equal(t, time.Hour*24*7, app.CacheHeaderTTL)
	assert.Equal(t, time.Hour*24, app.CacheHeaderSWR)
	assert.Empty(t, app.ResultStorages)
	assert.Empty(t, app.Storages)
	loader := app.Loaders[0].(*httploader.HTTPLoader)
	assert.Empty(t, loader.BaseURL)
	assert.Equal(t, "https", loader.DefaultScheme)
}

func TestBasic(t *testing.T) {
	placeholderData := `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`

	srv := CreateServer([]string{
		"-debug",
		"-port", "2345",
		"-imagor-secret", "foo",
		"-imagor-unsafe",
		"-imagor-auto-webp",
		"-imagor-auto-avif",
		"-imagor-disable-error-body",
		"-imagor-disable-params-endpoint",
		"-imagor-request-timeout", "16s",
		"-imagor-load-timeout", "7s",
		"-imagor-process-timeout", "19s",
		"-imagor-process-concurrency", "199",
		"-imagor-process-queue-size", "1999",
		"-imagor-base-path-redirect", "https://www.google.com",
		"-imagor-base-params", "filters:watermark(example.jpg)",
		"-imagor-cache-header-ttl", "169h",
		"-imagor-cache-header-swr", "167h",
		"-imagor-image-error-fallback", placeholderData,
		"-http-loader-insecure-skip-verify-transport",
		"-http-loader-override-response-headers", "cache-control,content-type",
		"-http-loader-base-url", "https://www.example.com/foo.org",
	})
	app := srv.App.(*imagor.Imagor)

	assert.Equal(t, 2345, srv.Port)
	assert.Equal(t, ":2345", srv.Addr)
	assert.True(t, app.Debug)
	assert.True(t, app.Unsafe)
	assert.True(t, app.AutoWebP)
	assert.True(t, app.DisableErrorBody)
	assert.True(t, app.DisableParamsEndpoint)
	assert.Equal(t, "RrTsWGEXFU2s1J1mTl1j_ciO-1E=", app.Signer.Sign("bar"))
	assert.Equal(t, time.Second*16, app.RequestTimeout)
	assert.Equal(t, time.Second*7, app.LoadTimeout)
	assert.Equal(t, time.Second*19, app.ProcessTimeout)
	assert.Equal(t, int64(199), app.ProcessConcurrency)
	assert.Equal(t, int64(1999), app.ProcessQueueSize)
	assert.Equal(t, "https://www.google.com", app.BasePathRedirect)
	assert.Equal(t, "filters:watermark(example.jpg)/", app.BaseParams)
	assert.Equal(t, time.Hour*169, app.CacheHeaderTTL)
	assert.Equal(t, time.Hour*167, app.CacheHeaderSWR)
	assert.Equal(t, placeholderData, app.ImageErrorFallback)

	httpLoader := app.Loaders[0].(*httploader.HTTPLoader)
	assert.True(t, httpLoader.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
	assert.Equal(t, "https://www.example.com/foo.org", httpLoader.BaseURL.String())
	assert.Equal(t, []string{"cache-control", "content-type"}, httpLoader.OverrideResponseHeaders)
}

func TestVersion(t *testing.T) {
	assert.Empty(t, CreateServer([]string{"-version"}))
}

func TestBind(t *testing.T) {
	srv := CreateServer([]string{
		"-debug",
		"-port", "2345",
		"-bind", ":4567",
	})
	assert.Equal(t, ":4567", srv.Addr)
}

func TestSentry(t *testing.T) {
	srv := CreateServer([]string{
		"-sentry-dsn", "https://12345@sentry.com/123",
	})
	assert.Equal(t, "https://12345@sentry.com/123", srv.SentryDsn)
}

func TestSignerAlgorithm(t *testing.T) {
	srv := CreateServer([]string{
		"-imagor-signer-type", "sha256",
	})
	app := srv.App.(*imagor.Imagor)
	assert.Equal(t, "WN6mgyl8pD4KTy5IDSBs0GcFPaV7-R970JLsd01pqAU=", app.Signer.Sign("bar"))

	srv = CreateServer([]string{
		"-imagor-signer-type", "sha512",
		"-imagor-signer-truncate", "32",
	})
	app = srv.App.(*imagor.Imagor)
	assert.Equal(t, "Kmml5ejnmsn7M7TszYkeM2j5G3bpI7mp", app.Signer.Sign("bar"))
}

func TestCacheHeaderNoCache(t *testing.T) {
	srv := CreateServer([]string{"-imagor-cache-header-no-cache"})
	app := srv.App.(*imagor.Imagor)
	assert.Empty(t, app.CacheHeaderTTL)
}

func TestDisableHTTPLoader(t *testing.T) {
	srv := CreateServer([]string{"-http-loader-disable"})
	app := srv.App.(*imagor.Imagor)
	assert.Empty(t, app.Loaders)
}

func TestFileLoader(t *testing.T) {
	srv := CreateServer([]string{
		"-file-safe-chars", "!",

		"-file-loader-base-dir", "./foo",
		"-file-loader-path-prefix", "abcd",
	})
	app := srv.App.(*imagor.Imagor)
	fileLoader := app.Loaders[0].(*filestorage.FileStorage)
	assert.Equal(t, "./foo", fileLoader.BaseDir)
	assert.Equal(t, "/abcd/", fileLoader.PathPrefix)
	assert.Equal(t, "!", fileLoader.SafeChars)
}

func TestFileStorage(t *testing.T) {
	srv := CreateServer([]string{
		"-file-safe-chars", "!",

		"-file-storage-base-dir", "./foo",
		"-file-storage-path-prefix", "abcd",

		"-file-result-storage-base-dir", "./bar",
		"-file-result-storage-path-prefix", "bcda",
	})
	app := srv.App.(*imagor.Imagor)
	assert.Equal(t, 1, len(app.Loaders))
	storage := app.Storages[0].(*filestorage.FileStorage)
	assert.Equal(t, "./foo", storage.BaseDir)
	assert.Equal(t, "/abcd/", storage.PathPrefix)
	assert.Equal(t, "!", storage.SafeChars)

	resultStorage := app.ResultStorages[0].(*filestorage.FileStorage)
	assert.Equal(t, "./bar", resultStorage.BaseDir)
	assert.Equal(t, "/bcda/", resultStorage.PathPrefix)
	assert.Equal(t, "!", resultStorage.SafeChars)
}

func TestPathStyle(t *testing.T) {
	srv := CreateServer([]string{
		"-imagor-storage-path-style", "digest",
		"-imagor-result-storage-path-style", "digest",
	})
	app := srv.App.(*imagor.Imagor)
	assert.Equal(t, "a9/99/3e364706816aba3e25717850c26c9cd0d89d", app.StoragePathStyle.Hash("abc"))
	assert.Equal(t, "30/fd/be2aa5086e0f0c50ea72dd3859a10d8071ad", app.ResultStoragePathStyle.HashResult(imagorpath.Parse("200x200/abc")))

	srv = CreateServer([]string{
		"-imagor-result-storage-path-style", "suffix",
	})
	app = srv.App.(*imagor.Imagor)
	assert.Equal(t, "abc.30fdbe2aa5086e0f0c50", app.ResultStoragePathStyle.HashResult(imagorpath.Parse("200x200/abc")))

	srv = CreateServer([]string{
		"-imagor-result-storage-path-style", "size",
	})
	app = srv.App.(*imagor.Imagor)
	assert.Equal(t, "abc.30fdbe2aa5086e0f0c50_200x200", app.ResultStoragePathStyle.HashResult(imagorpath.Parse("200x200/abc")))
}

func TestPrometheusBind(t *testing.T) {
	srv := CreateServer([]string{
		"-bind", ":2345",
		"-prometheus-bind", ":6789",
		"-prometheus-path", "/myprom",
	})
	assert.Equal(t, ":2345", srv.Addr)
	pm := srv.Metrics.(*prometheusmetrics.PrometheusMetrics)
	assert.Equal(t, pm.Path, "/myprom")
	assert.Equal(t, pm.Addr, ":6789")
}

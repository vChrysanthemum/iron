package iron

import "testing"

func TestSanitizeOptions(t *testing.T) {
	var (
		err error
	)

	var options = Options{
		RunMode:         "dev",
		ServeType:       "server",
		ListenStr:       "127.0.0.1:7812",
		LogPath:         "./test.log",
		AccessWhiteList: []string{},

		SiteViewDir:              "./",
		SiteStaticBasePath:       "./",
		SiteStaticUploadBasePath: "./",
	}

	var server Server
	err = server.Init(options)
	AssertErrIsNilForTest(t, err)
}

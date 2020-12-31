package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDownloadCommand(t *testing.T) {
	doTest := func(downloadPath, componentPackage string, expectedCommand []string) {
		actualResult := getDownloadCommand(downloadPath, componentPackage)
		assert.Equal(t, expectedCommand, actualResult)
	}

	testData := []struct {
		downloadPath     string
		componentPackage string
		expectedCommand  []string
	}{
		// test get the download command with package name
		{"function://public/default/test@v1", "function-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"packages", "download", "function://public/default/test@v1", "--path", PulsarDownloadRootDir + "/function-package.jar",
			},
		},
		{"sink://public/default/test@v1", "sink-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"packages", "download", "sink://public/default/test@v1", "--path", PulsarDownloadRootDir + "/sink-package.jar",
			},
		},
		{"source://public/default/test@v1", "source-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"packages", "download", "source://public/default/test@v1", "--path", PulsarDownloadRootDir + "/source-package.jar",
			},
		},
		// test get the download command with normal name
		{"/test", "test.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"functions", "download", "--path", "/test", "--destination-file", PulsarDownloadRootDir + "/test.jar",
			},
		},
		// test get the download command with a wrong package name
		{"source/public/default/test@v1", "source-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"functions", "download", "--path", "source/public/default/test@v1", "--destination-file", PulsarDownloadRootDir + "/source-package.jar",
			},
		},
		{"source:/public/default/test@v1", "source-package.jar",
			[]string{
				PulsarAdminExecutableFile,
				"--admin-url", "$webServiceURL",
				"functions", "download", "--path", "source:/public/default/test@v1", "--destination-file", PulsarDownloadRootDir + "/source-package.jar",
			},
		},
	}

	for _, v := range testData {
		doTest(v.downloadPath, v.componentPackage, v.expectedCommand)
	}
}

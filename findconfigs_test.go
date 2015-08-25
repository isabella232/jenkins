package jenkins

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindConfigFiles(t *testing.T) {
	root, err := extractTestConfigs()
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer func() {
		os.RemoveAll(root)
	}()

	/*
		Test root directory looks like this:

		$ unzip -l test-config-files.zip
		  Length     Date   Time    Name
		 --------    ----   ----    ----
		        0  08-25-15 11:32   a/
		        0  08-25-15 08:49   a/b/
		        0  08-25-15 08:49   a/b/c/
		        0  08-25-15 09:28   a/b/c/config.xml/
		        0  08-25-15 09:28   a/b/c/config.xml/other.txt   <<< is a directory
		        0  08-25-15 08:49   a/b/config.xml               <<< too deep
		      755  08-25-15 11:32   a/config.xml                 <<< want this one
		        0  08-25-15 08:50   config.xml
		        0  08-25-15 11:32   x/
		     1159  08-25-15 11:32   x/config.xml                 <<< and want this one
		 --------                   -------
		     1914                   10 files
	*/

	configs, err := findJobsInFilesystem(root)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}
	if len(configs) != 2 {
		t.Fatalf("want 2 but got %d\n", len(configs))
	}

	if configs[0] != root+"/"+"a/config.xml" && configs[0] != root+"/"+"x/config.xml" {
		t.Fatalf("want <root>/[a|x]/config.xml but got %s\n", configs[0])
	}

	if configs[1] != root+"/"+"a/config.xml" && configs[1] != root+"/"+"x/config.xml" {
		t.Fatalf("want <root>/[a|x]/config.xml but got %s\n", configs[1])
	}
}

func extractTestConfigs() (string, error) {
	r, err := zip.OpenReader("test-config-files.zip")
	if err != nil {
		return "", err
	}
	defer r.Close()

	name, err := ioutil.TempDir("", "configxml-")
	if err != nil {
		return "", err
	}

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}

		destinationFileName := name + "/" + f.Name
		parentDir := filepath.Dir(destinationFileName)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return "", err
		}

		dst, err := os.Create(destinationFileName)
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		_, err = io.Copy(dst, rc)
		if err != nil {
			return "", err
		}
		rc.Close()
		dst.Close()
	}
	return name, nil
}

func TestValidJobNameFromConfigFileName(t *testing.T) {
	for _, v := range []string{"path1/path2/pathN/jobname/config.xml", "./jobname/config.xml"} {
		jobName, err := jobNameFromConfigFileName(v)
		if err != nil {
			t.Fatalf("Unexpected error: %s\n", err)
		}
		if jobName != "jobname" {
			t.Fatalf("Want jobname but got %s\n", jobName)
		}
	}
}

func TestInValidJobNameFromConfigFileName(t *testing.T) {
	var err error
	_, err = jobNameFromConfigFileName("path1/path2/pathN/jobname/notaconfig.xml")
	if err == nil {
		t.Fatalf("Expected error.  Last path token is not config.xml.")
	}

	_, err = jobNameFromConfigFileName("config.xml")
	if err == nil {
		t.Fatalf("Expected error.  Want at least one / in job config file name.")
	}
}

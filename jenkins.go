package jenkins

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ae6rt/retry"
)

func NewClient(baseURL *url.URL, username, password string) Jenkins {
	return Client{baseURL: baseURL, userName: username, password: password}
}

func (client Client) GetJobSummariesFromFilesystem(root string) ([]JobSummary, error) {
	log.Printf("jenkins.GetJobSummariesFromFilesystem from %s...\n", root)

	if exists, err := dirExists(root); err != nil || !exists {
		if err != nil {
			return nil, err
		} else {
			return nil, fmt.Errorf("jenkins.GetJobSummariesFromFilesystem: root directory %s does not exist.\n", root)
		}
	}

	if jobDescriptors, err := client.GetJobs(); err != nil {
		return nil, err
	} else {
		summaries := make([]JobSummary, 0)
		for _, jobDescriptor := range jobDescriptors {
			if jobSummary, err := client.getJobSummary(jobDescriptor); err != nil {
				continue
			} else {
				summaries = append(summaries, jobSummary)
			}
		}
		return summaries, nil
	}
}

func (client Client) GetJobSummaries() ([]JobSummary, error) {
	log.Printf("jenkins.GetJobSummaries...\n")
	if jobDescriptors, err := client.GetJobs(); err != nil {
		return nil, err
	} else {
		summaries := make([]JobSummary, 0)
		for _, jobDescriptor := range jobDescriptors {
			if jobSummary, err := client.getJobSummary(jobDescriptor); err != nil {
				continue
			} else {
				summaries = append(summaries, jobSummary)
			}
		}
		return summaries, nil
	}
}

func (client Client) getJobSummary(jobDescriptor JobDescriptor) (JobSummary, error) {
	retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)

	var data []byte
	work := func() error {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/job/%s/config.xml", client.baseURL.String(), jobDescriptor.Name), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/xml")
		req.SetBasicAuth(client.userName, client.password)

		var responseCode int
		responseCode, data, err = consumeResponse(req)
		if err != nil {
			return err
		}

		if responseCode != http.StatusOK {
			log.Printf("%s", string(data))
			return fmt.Errorf("%s", string(data))
		}
		return nil
	}

	if err := retry.Try(work); err != nil {
		return JobSummary{}, err
	}

	jobType, err := getJobType(data)
	if err != nil {
		return JobSummary{}, err
	}

	reader := bytes.NewBuffer(data)

	switch jobType {
	case Maven:
		var maven JobConfig
		err = xml.NewDecoder(reader).Decode(&maven)
		if err != nil {
			return JobSummary{}, err
		}
		if !buildsSingleBranch(maven.SCM) {
			return JobSummary{}, fmt.Errorf("Maven-type job %#v contains more than one branch to build.  This is not supported.", jobDescriptor)
		}
		if !referencesSingleGitRepo(maven.SCM) {
			return JobSummary{}, fmt.Errorf("Maven-type job %#v contains more than one Git repository URL.  This is not supported.", jobDescriptor)
		}

		gitURL := maven.SCM.UserRemoteConfigs.UserRemoteConfig[0].URL
		if !strings.HasPrefix(gitURL, "ssh://") {
			return JobSummary{}, fmt.Errorf("Only ssh:// Git URLs are supported.", jobDescriptor)
		}

		return JobSummary{
			JobType:       Maven,
			JobDescriptor: jobDescriptor,
			GitURL:        gitURL,
			Branch:        maven.SCM.Branches.Branch[0].Name,
		}, nil
	case Freestyle:
		var freestyle FreeStyleJobConfig
		err = xml.NewDecoder(reader).Decode(&freestyle)
		if err != nil {
			return JobSummary{}, err
		}
		if !buildsSingleBranch(freestyle.SCM) {
			return JobSummary{}, fmt.Errorf("Freestyle-type job %s contains more than one branch to build.  This is not supported.", jobDescriptor)
		}
		if !referencesSingleGitRepo(freestyle.SCM) {
			return JobSummary{}, fmt.Errorf("Freestyle-type job %s contains more than one Git repository URL.  This is not supported.", jobDescriptor)
		}

		gitURL := freestyle.SCM.UserRemoteConfigs.UserRemoteConfig[0].URL
		if !strings.HasPrefix(gitURL, "ssh://") {
			return JobSummary{}, fmt.Errorf("Only ssh:// Git URLs are supported.", jobDescriptor)
		}
		return JobSummary{
			JobType:       Freestyle,
			JobDescriptor: jobDescriptor,
			GitURL:        gitURL,
			Branch:        freestyle.SCM.Branches.Branch[0].Name,
		}, nil
	}
	return JobSummary{}, fmt.Errorf("Unhandled job type for job name: %s\n", jobDescriptor.Name)
}

func buildsSingleBranch(scmInfo Scm) bool {
	return len(scmInfo.Branches.Branch) == 1
}

func referencesSingleGitRepo(scmInfo Scm) bool {
	return len(scmInfo.UserRemoteConfigs.UserRemoteConfig) == 1
}

// GetJobs retrieves the set of Jenkins jobs as a map indexed by job name.
func (client Client) GetJobs() (map[string]JobDescriptor, error) {
	retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)

	var data []byte
	work := func() error {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/json/jobs", client.baseURL.String()), nil)
		log.Printf("jenkins.GetJobs URL: %s\n", req.URL)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(client.userName, client.password)

		var responseCode int
		responseCode, data, err = consumeResponse(req)
		if err != nil {
			return err
		}

		if responseCode != http.StatusOK {
			log.Printf("%s", string(data))
			return fmt.Errorf("%s", string(data))
		}

		return nil
	}

	if err := retry.Try(work); err != nil {
		return nil, err
	}

	var t Jobs
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	jobs := make(map[string]JobDescriptor)
	for _, v := range t.Jobs {
		jobs[v.Name] = v
	}
	return jobs, nil
}

// GetJobConfig retrieves the Jenkins jobs config for the named job.
func (client Client) GetJobConfig(jobName string) (JobConfig, error) {
	retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)

	var data []byte
	work := func() error {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/job/%s/config.xml", client.baseURL.String(), jobName), nil)
		log.Printf("jenkins.GetJobConfig URL: %s\n", req.URL)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/xml")
		req.SetBasicAuth(client.userName, client.password)

		var responseCode int
		responseCode, data, err = consumeResponse(req)
		if err != nil {
			return err
		}

		if responseCode != http.StatusOK {
			log.Printf("%s", string(data))
			return fmt.Errorf("%s", string(data))
		}
		return nil
	}
	if err := retry.Try(work); err != nil {
		return JobConfig{}, err
	}

	var config JobConfig
	reader := bytes.NewBuffer(data)
	if err := xml.NewDecoder(reader).Decode(&config); err != nil {
		return JobConfig{}, err
	}
	config.JobName = jobName
	return config, nil
}

// CreateJob creates a Jenkins job with the given name for the given XML job config.
func (client Client) CreateJob(jobName, jobConfigXML string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/createItem?name=%s", client.baseURL.String(), jobName), bytes.NewBuffer([]byte(jobConfigXML)))
	log.Printf("jenkins.CreateJob URL: %s\n", req.URL)
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/xml")
	req.SetBasicAuth(client.userName, client.password)

	responseCode, data, err := consumeResponse(req)
	if err != nil {
		return err
	}
	if responseCode != http.StatusOK {
		return fmt.Errorf("Error creating Jenkins job.  Status code: %d, response=%s\n", responseCode, string(data))
	}
	return nil
}

// DeleteJob creates a Jenkins job with the given name for the given XML job config.
func (client Client) DeleteJob(jobName string) error {
	retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)
	work := func() error {
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/job/%s/doDelete", client.baseURL.String(), jobName), bytes.NewBuffer([]byte("")))
		log.Printf("jenkins.DeleteJob URL: %s\n", req.URL)
		if err != nil {
			return err
		}
		req.Header.Set("Content-type", "application/xml")
		req.SetBasicAuth(client.userName, client.password)

		responseCode, data, err := consumeResponse(req)
		if err != nil {
			return err
		}
		if responseCode != http.StatusFound {
			return fmt.Errorf("Error deleting Jenkins job.  Status code: %d, response=%s\n", responseCode, string(data))
		}
		return nil
	}
	return retry.Try(work)
}

func consumeResponse(req *http.Request) (int, []byte, error) {
	var response *http.Response
	var err error
	/*
	   $ curl -i -d "" http://jenkins.example.com:8080/job/somejob/doDelete
	   HTTP/1.1 302 Found
	   Location: http://jenkins.example.com:8080/
	   Content-Length: 0
	   Server: Jetty(8.y.z-SNAPSHOT)
	*/
	// So 302 means it worked, but we don't want to follow the redirect.  Why use http.DefaultTransport.RoundTrip:
	// http://stackoverflow.com/questions/14420222/query-url-without-redirect-in-go
	response, err = http.DefaultTransport.RoundTrip(req)

	if err != nil {
		return 0, nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	return response.StatusCode, data, nil
}

func getJobType(xmlDocument []byte) (JobType, error) {
	decoder := xml.NewDecoder(bytes.NewBuffer(xmlDocument))

	var t string
	for {
		token, err := decoder.Token()
		if err != nil {
			return Unknown, err
		}
		if v, ok := token.(xml.StartElement); ok {
			t = v.Name.Local
			break
		}
	}

	switch t {
	case "maven2-moduleset":
		return Maven, nil
	case "project":
		return Freestyle, nil
	}
	return Unknown, nil
}

func dirExists(dirPath string) (bool, error) {
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

// findJobs is similar to "find <dir> -name somename -maxdepth d.
// Seeking jobname/config.xml:  resides in one level below root
// Discard config.xml:  resides at root
// Discard jobname/a/b/config.xml:  resides more than one level below root
func findJobs(root, fileName string, maxDepth int) ([]string, error) {
	files := make([]string, 0)
	markFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.Count(path, "/") >= maxDepth {
			return filepath.SkipDir
		}
		if strings.Count(path, "/") == maxDepth-1 && info.Mode().IsRegular() && info.Name() == fileName {
			files = append(files, path)
		}
		return nil
	}

	if err := os.Chdir(root); err != nil {
		return nil, err
	}

	err := filepath.Walk(".", markFn)
	if err != nil {
		return nil, err
	}

	return files, nil
}

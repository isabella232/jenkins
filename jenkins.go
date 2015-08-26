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
	log.Printf("jenkins.GetJobSummariesFromFilesystem from %s\n", root)

	if exists, err := dirExists(root); err != nil || !exists {
		if err != nil {
			return nil, err
		} else {
			return nil, fmt.Errorf("jenkins.GetJobSummariesFromFilesystem: root directory %s of Jenkins jobs does not exist.\n", root)
		}
	}

	jobConfigFiles, err := findJobsInFilesystem(root)
	if err != nil {
		return nil, err
	}

	summaries := make([]JobSummary, 0)
	for _, configFile := range jobConfigFiles {
		jobName, err := jobNameFromConfigFileName(configFile)
		if err != nil {
			log.Printf("Cannot acquire job name from config file name %s: %v.  Skipping.\n", configFile, err)
			continue
		}
		jobDescriptor := JobDescriptor{Name: jobName}

		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("Cannot read config file %s: %v.  Skipping.\n", configFile, err)
			continue
		}

		jobSummary, err := getSummaryFromConfigBytes(data, jobDescriptor)
		if err != nil {
			log.Printf("Cannot get job summary from config file data %s: %v.  Skipping.\n", configFile, err)
			continue
		}

		summaries = append(summaries, jobSummary)
	}
	return summaries, nil
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

	summary, err := getSummaryFromConfigBytes(data, jobDescriptor)
	if err != nil {
		return JobSummary{}, err
	}
	return summary, nil
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

// findJobsInFilesystem is similar to "find <dir> -name somename -maxdepth d.  We strictly want files at exactly depth.
// Seeking jobname/config.xml:  resides one level below root
// Discard config.xml:  resides at root
// Discard jobname/a/b/config.xml:  resides more than one level below root
// @root is the jobs directory, with a typical value of something like /srv/jenkins/jobs
func findJobsInFilesystem(root string) ([]string, error) {
	files := make([]string, 0)
	markFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// record this as a too-deep path we never want to traverse again
		if info.IsDir() && strings.Count(path, "/") > 1 {
			return filepath.SkipDir
		}

		// We compare again the depth to 1 because ./config.xml is an undesirable possibility.  For jenkins jobs,
		//  if root is /srv/jenkins/jobs, /srv/jenkins/jobs/config.xml is not a job config, whatever else it may be.
		if strings.Count(path, "/") == 1 && info.Mode().IsRegular() && info.Name() == "config.xml" {
			files = append(files, root+"/"+path)
		}
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.Chdir(pwd)
		if err != nil {
			log.Printf("jenkins.findJobsInFilesystem() cannot restore working directory to %s: %v\n", pwd, err)
		}
	}()

	// Change directory to root so we have no need to know how many "/" root itself contains.
	if err := os.Chdir(root); err != nil {
		return nil, err
	}

	// Walk relative to root.
	err = filepath.Walk(".", markFn)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func getSummaryFromConfigBytes(data []byte, jobDescriptor JobDescriptor) (JobSummary, error) {

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
			return JobSummary{}, fmt.Errorf("Maven-type job [%s] does not contain exactly one branch to build.  This is not supported.", jobDescriptor.Name)
		}
		if !referencesSingleGitRepo(maven.SCM) {
			return JobSummary{}, fmt.Errorf("Maven-type job [%s] does not contain exactly one Git repository URL.  This is not supported.", jobDescriptor.Name)
		}

		gitURL := maven.SCM.UserRemoteConfigs.UserRemoteConfig[0].URL
		if !strings.HasPrefix(gitURL, "ssh://") {
			return JobSummary{}, fmt.Errorf("Only ssh:// Git URLs are supported %s", jobDescriptor.Name)
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
			return JobSummary{}, fmt.Errorf("Freestyle-type job [%s] does not contain exactly one branch to build.  This is not supported.", jobDescriptor.Name)
		}
		if !referencesSingleGitRepo(freestyle.SCM) {
			return JobSummary{}, fmt.Errorf("Freestyle-type job [%s] does not contain exactly one Git repository URL.  This is not supported.", jobDescriptor.Name)
		}

		gitURL := freestyle.SCM.UserRemoteConfigs.UserRemoteConfig[0].URL
		if !strings.HasPrefix(gitURL, "ssh://") {
			return JobSummary{}, fmt.Errorf("Only ssh:// Git URLs are supported %s", jobDescriptor.Name)
		}
		return JobSummary{
			JobType:       Freestyle,
			JobDescriptor: jobDescriptor,
			GitURL:        gitURL,
			Branch:        freestyle.SCM.Branches.Branch[0].Name,
		}, nil
	}
	return JobSummary{}, fmt.Errorf("Unhandled job type for job name: %s\n", jobDescriptor)
}

// jobNameFromConfigFileName returns "jobname" from path1/path2/pathN/jobname/config.xml
func jobNameFromConfigFileName(configFileName string) (string, error) {
	parts := strings.Split(configFileName, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("Job config file name expected to have at least one /: %s", configFileName)
	}
	if parts[len(parts)-1] != "config.xml" {
		return "", fmt.Errorf("Job config file name does not end in config.xml: %s", configFileName)
	}
	return parts[len(parts)-2], nil
}

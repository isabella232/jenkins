package jenkins

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	httpClient *http.Client = &http.Client{}
)

// GetJobs retrieves the set of Jenkins jobs as a map indexed by job name.
func GetJobs(baseUrl string) (map[string]JobDescriptor, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/json/jobs", baseUrl), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	responseCode, data, err := consumeResponse(req)
	if err != nil {
		return nil, err
	}

	if responseCode != http.StatusOK {
		log.Printf("%s", string(data))
		return nil, fmt.Errorf("%s", string(data))
	}

	var t Jobs
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}

	jobs := make(map[string]JobDescriptor)
	for _, v := range t.Jobs {
		jobs[v.Name] = v
	}
	return jobs, nil
}

// GetJobConfig retrieves the Jenkins jobs config for the named job.
func GetJobConfig(baseUrl, jobName string) (JobConfig, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/job/%s/config.xml", baseUrl, jobName), nil)
	if err != nil {
		return JobConfig{}, err
	}
	req.Header.Set("Accept", "application/xml")

	responseCode, data, err := consumeResponse(req)
	if err != nil {
		return JobConfig{}, err
	}

	if responseCode != http.StatusOK {
		log.Printf("%s", string(data))
		return JobConfig{}, fmt.Errorf("%s", string(data))
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
func CreateJob(baseUrl, jobName, jobConfigXML string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/createItem?name=%s", baseUrl, jobName), bytes.NewBuffer([]byte(jobConfigXML)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/xml")
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
func DeleteJob(baseUrl, jobName string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/job/%s/doDelete", baseUrl, jobName), bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/xml")
	responseCode, data, err := consumeResponse(req)
	if err != nil {
		return err
	}
	// todo figure out what the response code is if the delete worked.
	if responseCode != http.StatusOK {
		//return fmt.Errorf("Error creating Jenkins job.  Status code: %d, response=%s\n", responseCode, string(data))
		fmt.Printf("Error creating Jenkins job.  Status code: %d, response=%s\n", responseCode, string(data))
	}
	return nil
}

func consumeResponse(req *http.Request) (int, []byte, error) {
	var response *http.Response
	var err error
	response, err = httpClient.Do(req)

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

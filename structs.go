package jenkins

import (
	"encoding/xml"
	"net/url"
)

type (
	Jenkins interface {
		GetJobs() (map[string]JobDescriptor, error)
		GetJobConfig(jobName string) (JobConfig, error)
		CreateJob(jobName, jobConfigXML string) error
		DeleteJob(jobName string) error
	}

	Client struct {
		baseURL *url.URL
		Jenkins
	}

	JobDescriptor struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	Jobs struct {
		Jobs []JobDescriptor `json:"jobs"`
	}

	JobConfig struct {
		XMLName    xml.Name   `xml:"maven2-moduleset"`
		SCM        Scm        `xml:"scm"`
		Publishers Publishers `xml:"publishers"`
		RootModule RootModule `xml:"rootModule"`
		JobName    string
	}

	Scm struct {
		XMLName           xml.Name          `xml:"scm"`
		Class             string            `xml:"class,attr"`
		UserRemoteConfigs UserRemoteConfigs `xml:userRemoteConfigs`
		Branches          Branches          `xml:"branches"`
	}

	Publishers struct {
		XMLName            xml.Name            `xml:"publishers"`
		RedeployPublishers []RedeployPublisher `xml:"hudson.maven.RedeployPublisher"`
	}

	RedeployPublisher struct {
		XMLName xml.Name `xml:"hudson.maven.RedeployPublisher"`
		URL     string   `xml:"url"`
	}

	UserRemoteConfigs struct {
		XMLName          xml.Name           `xml:"userRemoteConfigs"`
		UserRemoteConfig []UserRemoteConfig `xml:"hudson.plugins.git.UserRemoteConfig"`
	}

	UserRemoteConfig struct {
		XMLName xml.Name `xml:"hudson.plugins.git.UserRemoteConfig"`
		URL     string   `xml:"url"`
	}

	Branches struct {
		XMLName xml.Name `xml:"branches"`
		Branch  []Branch `xml:"hudson.plugins.git.BranchSpec"`
	}

	Branch struct {
		XMLName xml.Name `xml:"hudson.plugins.git.BranchSpec"`
		Name    string   `xml:"name"`
	}

	RootModule struct {
		XMLName    xml.Name `xml:"rootModule"`
		GroupID    string   `xml:"groupId"`
		ArtifactID string   `xml:"artifactId"`
	}
)

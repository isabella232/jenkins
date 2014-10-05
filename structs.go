package jenkins

import "encoding/xml"

type (
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
	}

	Scm struct {
		XMLName xml.Name `xml:"scm"`
		Class   string   `xml:"class,attr"`
	}

	UserRemoteConfigs struct {
		XMLName xml.Name `xml:"userRemoteConfigs"`
		Configs []UserRemoteConfig
	}

	UserRemoteConfig struct {
		XMLName xml.Name `xml:"hudson.plugins.git.UserRemoteConfig"`
		URL     string   `xml:"url"`
	}

	Branches struct {
		XMLName     xml.Name `xml:"branches"`
		GitBranches []Branch
	}

	Branch struct {
		XMLName xml.Name `xml:"hudson.plugins.git.BranchSpec"`
		Name    string   `xml:"name"`
	}

	Publishers struct {
		XMLName            xml.Name `xml:"publishers"`
		RedeployPublishers []RedeployPublisher
	}

	RedeployPublisher struct {
		XMLName xml.Name `xml:"hudson.maven.RedeployPublisher"`
		URL     string   `xml:"url"`
	}

	RootModule struct {
		XMLName    xml.Name `xml:"rootModule"`
		GroupID    string   `xml:"groupId"`
		ArtifactID string   `xml:"artifactId"`
	}
)

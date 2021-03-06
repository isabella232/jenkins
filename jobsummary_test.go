package jenkins

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var jobConfig1 string = `
<?xml version='1.0' encoding='UTF-8'?>
<maven2-moduleset plugin="maven-plugin@2.6">
  <actions/>
  <description>This will build a feature branche for the service</description>
  <logRotator class="hudson.tasks.LogRotator">
    <daysToKeep>30</daysToKeep>
    <numToKeep>30</numToKeep>
    <artifactDaysToKeep>-1</artifactDaysToKeep>
    <artifactNumToKeep>-1</artifactNumToKeep>
  </logRotator>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.plugins.git.GitSCM" plugin="git@2.2.4">
    <configVersion>2</configVersion>
    <userRemoteConfigs>
      <hudson.plugins.git.UserRemoteConfig>
        <url>ssh://example.com/proj/cool.git</url>
      </hudson.plugins.git.UserRemoteConfig>
    </userRemoteConfigs>
    <branches>
      <hudson.plugins.git.BranchSpec>
        <name>origin/develop</name>
      </hudson.plugins.git.BranchSpec>
    </branches>
    <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
    <submoduleCfg class="list"/>
    <extensions/>
  </scm>
  <quietPeriod>0</quietPeriod>
  <scmCheckoutRetryCount>3</scmCheckoutRetryCount>
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers>
    <hudson.triggers.SCMTrigger>
      <spec># Every 3 min.
H/3 * * * *
</spec>
      <ignorePostCommitHooks>false</ignorePostCommitHooks>
    </hudson.triggers.SCMTrigger>
  </triggers>
  <concurrentBuild>false</concurrentBuild>
  <rootModule>
    <groupId>com.example.widgets</groupId>
    <artifactId>widge</artifactId>
  </rootModule>
  <goals>clean install</goals>
  <mavenName>maven 3.2.1</mavenName>
  <aggregatorStyleBuild>true</aggregatorStyleBuild>
  <incrementalBuild>false</incrementalBuild>
  <localRepository class="hudson.maven.local_repo.PerJobLocalRepositoryLocator"/>
  <ignoreUpstremChanges>true</ignoreUpstremChanges>
  <archivingDisabled>false</archivingDisabled>
  <siteArchivingDisabled>false</siteArchivingDisabled>
  <fingerprintingDisabled>false</fingerprintingDisabled>
  <resolveDependencies>false</resolveDependencies>
  <processPlugins>false</processPlugins>
  <mavenValidationLevel>-1</mavenValidationLevel>
  <runHeadless>true</runHeadless>
  <disableTriggerDownstreamProjects>false</disableTriggerDownstreamProjects>
  <settings class="jenkins.mvn.DefaultSettingsProvider"/>
  <globalSettings class="jenkins.mvn.DefaultGlobalSettingsProvider"/>
  <reporters>
    <hudson.maven.reporters.MavenMailer>
      <recipients>build.failures@example.com</recipients>
      <dontNotifyEveryUnstableBuild>false</dontNotifyEveryUnstableBuild>
      <sendToIndividuals>true</sendToIndividuals>
      <perModuleEmail>true</perModuleEmail>
    </hudson.maven.reporters.MavenMailer>
  </reporters>
  <publishers>
    <hudson.maven.RedeployPublisher>
      <id>example-snapshots</id>
      <url>http://nexus.example.com/nexus/content/repositories/snapshots/</url>
      <uniqueVersion>false</uniqueVersion>
      <evenIfUnstable>false</evenIfUnstable>
    </hudson.maven.RedeployPublisher>
  </publishers>
  <buildWrappers/>
  <prebuilders/>
  <postbuilders>
    <hudson.tasks.Shell>
      <command>#!/bin/sh
echo &quot;Hello, world</command>
    </hudson.tasks.Shell>
  </postbuilders>
  <runPostStepsIfResult>
    <name>SUCCESS</name>
    <ordinal>0</ordinal>
    <color>BLUE</color>
    <completeBuild>true</completeBuild>
  </runPostStepsIfResult>
</maven2-moduleset>
`

var freestyle1 string = `
<?xml version='1.0' encoding='UTF-8'?>
<project>
  <actions/>
  <description>Builds branches under origin/story/...</description>
  <logRotator class="hudson.tasks.LogRotator">
    <daysToKeep>60</daysToKeep>
    <numToKeep>-1</numToKeep>
    <artifactDaysToKeep>-1</artifactDaysToKeep>
    <artifactNumToKeep>-1</artifactNumToKeep>
  </logRotator>
  <keepDependencies>false</keepDependencies>
  <scm class="hudson.plugins.git.GitSCM" plugin="git@2.2.2">
    <configVersion>2</configVersion>
    <userRemoteConfigs>
      <hudson.plugins.git.UserRemoteConfig>
        <url>ssh://example.com/proj/cool.git</url>
      </hudson.plugins.git.UserRemoteConfig>
    </userRemoteConfigs>
    <branches>
      <hudson.plugins.git.BranchSpec>
        <name>origin/develop</name>
      </hudson.plugins.git.BranchSpec>
    </branches>
    <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
    <submoduleCfg class="list"/>
    <extensions>
      <hudson.plugins.git.extensions.impl.PerBuildTag/>
      <hudson.plugins.git.extensions.impl.WipeWorkspace/>
      <hudson.plugins.git.extensions.impl.SubmoduleOption>
        <disableSubmodules>false</disableSubmodules>
        <recursiveSubmodules>true</recursiveSubmodules>
        <trackingSubmodules>false</trackingSubmodules>
      </hudson.plugins.git.extensions.impl.SubmoduleOption>
    </extensions>
  </scm>
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers>
    <hudson.triggers.SCMTrigger>
      <spec># Every 3 min.
H/3 * * * *</spec>
      <ignorePostCommitHooks>false</ignorePostCommitHooks>
    </hudson.triggers.SCMTrigger>
  </triggers>
  <concurrentBuild>false</concurrentBuild>
  <builders>
    <hudson.tasks.Ant plugin="ant@1.2">
    </hudson.tasks.Ant>
  </builders>
  <publishers>
  </publishers>
  <buildWrappers/>
</project>`

var notSSH string = `
<?xml version='1.0' encoding='UTF-8'?>
<maven2-moduleset plugin="maven-plugin@2.6">
  <actions/>
  <description>This will build a feature branche for the service</description>
  <logRotator class="hudson.tasks.LogRotator">
    <daysToKeep>30</daysToKeep>
    <numToKeep>30</numToKeep>
    <artifactDaysToKeep>-1</artifactDaysToKeep>
    <artifactNumToKeep>-1</artifactNumToKeep>
  </logRotator>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.plugins.git.GitSCM" plugin="git@2.2.4">
    <configVersion>2</configVersion>
    <userRemoteConfigs>
      <hudson.plugins.git.UserRemoteConfig>
        <url>http://example.com/proj/cool.git</url>
      </hudson.plugins.git.UserRemoteConfig>
    </userRemoteConfigs>
    <branches>
      <hudson.plugins.git.BranchSpec>
        <name>origin/develop</name>
      </hudson.plugins.git.BranchSpec>
    </branches>
    <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
    <submoduleCfg class="list"/>
    <extensions/>
  </scm>
  <quietPeriod>0</quietPeriod>
  <scmCheckoutRetryCount>3</scmCheckoutRetryCount>
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers>
    <hudson.triggers.SCMTrigger>
      <spec># Every 3 min.
H/3 * * * *
</spec>
      <ignorePostCommitHooks>false</ignorePostCommitHooks>
    </hudson.triggers.SCMTrigger>
  </triggers>
  <concurrentBuild>false</concurrentBuild>
  <rootModule>
    <groupId>com.example.widgets</groupId>
    <artifactId>widge</artifactId>
  </rootModule>
  <goals>clean install</goals>
  <mavenName>maven 3.2.1</mavenName>
  <aggregatorStyleBuild>true</aggregatorStyleBuild>
  <incrementalBuild>false</incrementalBuild>
  <localRepository class="hudson.maven.local_repo.PerJobLocalRepositoryLocator"/>
  <ignoreUpstremChanges>true</ignoreUpstremChanges>
  <archivingDisabled>false</archivingDisabled>
  <siteArchivingDisabled>false</siteArchivingDisabled>
  <fingerprintingDisabled>false</fingerprintingDisabled>
  <resolveDependencies>false</resolveDependencies>
  <processPlugins>false</processPlugins>
  <mavenValidationLevel>-1</mavenValidationLevel>
  <runHeadless>true</runHeadless>
  <disableTriggerDownstreamProjects>false</disableTriggerDownstreamProjects>
  <settings class="jenkins.mvn.DefaultSettingsProvider"/>
  <globalSettings class="jenkins.mvn.DefaultGlobalSettingsProvider"/>
  <reporters>
    <hudson.maven.reporters.MavenMailer>
      <recipients>build.failures@example.com</recipients>
      <dontNotifyEveryUnstableBuild>false</dontNotifyEveryUnstableBuild>
      <sendToIndividuals>true</sendToIndividuals>
      <perModuleEmail>true</perModuleEmail>
    </hudson.maven.reporters.MavenMailer>
  </reporters>
  <publishers>
    <hudson.maven.RedeployPublisher>
      <id>example-snapshots</id>
      <url>http://nexus.example.com/nexus/content/repositories/snapshots/</url>
      <uniqueVersion>false</uniqueVersion>
      <evenIfUnstable>false</evenIfUnstable>
    </hudson.maven.RedeployPublisher>
  </publishers>
  <buildWrappers/>
  <prebuilders/>
  <postbuilders>
    <hudson.tasks.Shell>
      <command>#!/bin/sh
echo &quot;Hello, world</command>
    </hudson.tasks.Shell>
  </postbuilders>
  <runPostStepsIfResult>
    <name>SUCCESS</name>
    <ordinal>0</ordinal>
    <color>BLUE</color>
    <completeBuild>true</completeBuild>
  </runPostStepsIfResult>
</maven2-moduleset>
`

func TestHttpMavenJobSummary(t *testing.T) {
	jobMap := make(map[string]string)
	jobMap["maven"] = jobConfig1
	jobMap["freestyle"] = freestyle1

	for jobType, v := range jobMap {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url := *r.URL
			if url.Path != "/job/thejob/config.xml" {
				t.Fatalf("getJobSummary() URL path expected to end with config.xml: %s\n", url.Path)
			}
			if r.Header.Get("Accept") != "application/xml" {
				t.Fatalf("getJobSummary() expected request Accept header to be application/xml but found %s\n", r.Header.Get("Accept"))
			}
			if r.Header.Get("Authorization") != "Basic dTpw" {
				t.Fatalf("Want Basic dTpw but got %s\n", r.Header.Get("Authorization"))
			}
			fmt.Fprintln(w, v)
		}))

		url, _ := url.Parse(testServer.URL)
		jenkinsClient := Client{baseURL: url, userName: "u", password: "p"}
		summary, err := jenkinsClient.getJobSummary(JobDescriptor{Name: "thejob"})
		if err != nil {
			t.Fatalf("Unexpected error: %v\n", err)
		}
		if summary.JobDescriptor.Name != "thejob" {
			t.Fatalf("Want thejob but got: %s\n", summary.JobDescriptor.Name)
		}
		switch jobType {
		case "maven":
			if summary.JobType != Maven {
				t.Fatalf("Want Maven type but got: %v\n", summary.JobType)
			}
		case "freestyle":
			if summary.JobType != Freestyle {
				t.Fatalf("Want Freestyle type but got: %d\n", summary.JobType)
			}
		}
		if summary.GitURL != "" {
			t.Fatalf("Want empty Git URL but but got: %s\n", summary.GitURL)
		}
		if summary.Branch != "" {
			t.Fatalf("Want empty branch but got: %s\n", summary.Branch)
		}

		testServer.Close()
	}

}

func TestHttpUnknownJobSummary(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := *r.URL
		if url.Path != "/job/thejob/config.xml" {
			t.Fatalf("getJobSummary() URL path expected to end with config.xml: %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/xml" {
			t.Fatalf("getJobSummary() expected request Accept header to be application/xml but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want Basic dTpw but got %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, "<foo/>")
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	jenkinsClient := Client{baseURL: url, userName: "u", password: "p"}
	_, err := jenkinsClient.getJobSummary(JobDescriptor{Name: "thejob"})
	if err == nil {
		t.Fatalf("Expected error owing to unknown job type\n")
	}
}

func TestJobSummariesFromFilesystem(t *testing.T) {
	root, err := extractTestConfigs()
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer func() {
		os.RemoveAll(root)
	}()

	jenkinsClient := Client{baseURL: nil, userName: "u", password: "p"}

	summaries, err := jenkinsClient.GetJobSummariesFromFilesystem(root)
	if len(summaries) != 2 {
		t.Fatalf("Want 2 but got %d\n", len(summaries))
	}

	for _, v := range summaries {
		if !(v.JobDescriptor.Name == "a" || v.JobDescriptor.Name == "x") {
			t.Fatalf("Want job name a or x but got %s\n", v.JobDescriptor.Name)
		}
		switch v.JobDescriptor.Name {
		case "a":
			if v.JobType != Maven {
				t.Fatalf("Want Maven job type but got %d\n", v.JobType)
			}
			if v.GitURL != "" {
				t.Fatalf("Want empty git repository but got %s\n", v.GitURL)
			}
		case "x":
			if v.JobType != Freestyle {
				t.Fatalf("Want Freestyle job type but got %d\n", v.JobType)
			}
			if v.GitURL != "" {
				t.Fatalf("Want empty git repository but got %s\n", v.GitURL)
			}
		}
	}
}

func TestJobSummariesFromFilesystemNoSuchRoot(t *testing.T) {
	root, err := extractTestConfigs()
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	defer func() {
		os.RemoveAll(root)
	}()

	jenkinsClient := Client{baseURL: nil, userName: "u", password: "p"}

	_, err = jenkinsClient.GetJobSummariesFromFilesystem(root + "/nosuchdirectory")
	if err == nil {
		t.Fatalf("Want an error when getting summaries from a non-existent directory\n")
	}
}

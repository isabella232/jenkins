package jenkins

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	jobConfigXML string = `
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
        <name>*/develop</name>
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
)

func TestGetJobConfig(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := *r.URL
		if url.Path != "/job/thejob/config.xml" {
			t.Fatalf("GetJobs() URL path expected to end with config.xml: %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/xml" {
			t.Fatalf("GetJobs() expected request Accept header to be application/xml but found %s\n", r.Header.Get("Accept"))
		}
		fmt.Fprintln(w, jobConfigXML)
	}))
	defer testServer.Close()

	cfg, err := GetJobConfig(testServer.URL, "thejob")
	if err != nil {
		t.Fatalf("GetJobConfig() not expecting an error, but received: %v\n", err)
	}

	if cfg.SCM.Class != "hudson.plugins.git.GitSCM" {
		t.Fatalf("Wanted SCM.Class == hudson.plugins.git.GitSCM but found %d\n", cfg.SCM.Class)
	}

	if len(cfg.Publishers.RedeployPublishers) != 1 {
		t.Fatalf("Wanted Publishers.RedeployPublishers slice of length 1 but found %d\n", len(cfg.Publishers.RedeployPublishers))
	}

	if cfg.RootModule.GroupID != "com.example.widgets" {
		t.Fatalf("Wanted RootModule.GroupID == com.example.com but found %d\n", cfg.RootModule.GroupID)
	}

	if cfg.RootModule.ArtifactID != "widge" {
		t.Fatalf("Wanted RootModule.ArtifactID == widge but found %d\n", cfg.RootModule.ArtifactID)
	}

	/*
		{XMLName:{Space: Local:maven2-moduleset} SCM:{XMLName:{Space: Local:scm} Class:hudson.plugins.git.GitSCM} Publishers:{XMLName:{Space: Local:publishers} RedeployPublishers:[]} RootModule:{XMLName:{Space: Local:rootModule} GroupID:com.example.widgets ArtifactID:widge}}
	*/
}

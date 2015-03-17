package jenkins

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	fooJob string = `
<?xml version='1.0' encoding='UTF-8'?>
<maven2-moduleset plugin="maven-plugin@2.5">
  <actions/>
  <description>This will build the develop branch for User microservice.&#xd;
&lt;p/&gt;&#xd;
NOTE: This will be deployed to nexus, so if the feature version isn&apos;t used, then it will overwrite other snapshots</description>
  <logRotator class="hudson.tasks.LogRotator">
    <daysToKeep>30</daysToKeep>
    <numToKeep>30</numToKeep>
    <artifactDaysToKeep>-1</artifactDaysToKeep>
    <artifactNumToKeep>-1</artifactNumToKeep>
  </logRotator>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.plugins.git.GitSCM" plugin="git@2.2.2">
    <configVersion>2</configVersion>
    <userRemoteConfigs>
      <hudson.plugins.git.UserRemoteConfig>
        <url>ssh://git.example.com/proj/u.git</url>
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
    <artifactId>user</artifactId>
  </rootModule>
  <goals>clean install -Dspring.datasource.url=jdbc:mysql://localhost:3703/mgdb?zeroDateTimeBehavior=convertToNull -Dspring.datasource.username=root -Dspring.datasource.password=r00t -Dspring.datasource.driverClassName=com.mysql.jdbc.Driver -DprofileResourceUrl=http://localhost:29288/papi/v1/profiles</goals>
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
      <recipients></recipients>
      <dontNotifyEveryUnstableBuild>false</dontNotifyEveryUnstableBuild>
      <sendToIndividuals>true</sendToIndividuals>
      <perModuleEmail>true</perModuleEmail>
    </hudson.maven.reporters.MavenMailer>
  </reporters>
  <publishers>
    <hudson.maven.RedeployPublisher>
      <id>snapshots</id>
      <url>http://nexus.example.com:8081/nexus/content/repositories/snapshots/</url>
      <uniqueVersion>false</uniqueVersion>
      <evenIfUnstable>false</evenIfUnstable>
    </hudson.maven.RedeployPublisher>
  </publishers>
  <buildWrappers/>
  <prebuilders/>
  <postbuilders/>
  <runPostStepsIfResult>
    <name>SUCCESS</name>
    <ordinal>0</ordinal>
    <color>BLUE</color>
    <completeBuild>true</completeBuild>
  </runPostStepsIfResult>
</maven2-moduleset>
`
)

func TestCreateJenkinsJobsNoError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/createItem" {
			t.Fatalf("wanted URL path /createItem but found %s\n", url.Path)
		}
		if r.Header.Get("Content-type") != "application/xml" {
			t.Fatalf("wanted  Content-type header application/xml but found %s\n", r.Header.Get("Content-type"))
		}
        if r.Header.Get("Authorization") != "Basic dTpw" {
            t.Fatalf("Want Basic dTpw but got %s\n", r.Header.Get("Authorization"))
        }
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error reading POST body: %v\n", err)
		}
		if bytes.Compare([]byte(fooJob), data) != 0 {
			t.Fatalf("received unexpected []byte.  Expecting exactly the same as []byte(jenkinsJobConfig)")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	jenkinsClient := NewClient(url, "u", "p")
	err := jenkinsClient.CreateJob("job-name", fooJob)
	if err != nil {
		t.Fatalf("JenkinsJobCreate() not expecting an error, but received: %v\n", err)
	}
}

func TestCreateJenkinsJobs500(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/createItem" {
			t.Fatalf("wanted URL path /createItem but found %s\n", url.Path)
		}
		if r.Header.Get("Content-type") != "application/xml" {
			t.Fatalf("wanted  Content-type header application/xml but found %s\n", r.Header.Get("Content-type"))
		}
        if r.Header.Get("Authorization") != "Basic dTpw" {
            t.Fatalf("Want Basic dTpw but got %s\n", r.Header.Get("Authorization"))
        }
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error reading POST body: %v\n", err)
		}
		if bytes.Compare([]byte(fooJob), data) != 0 {
			t.Fatalf("received unexpected []byte.  Expecting exactly the same as []byte(jenkinsJobConfig)")
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
    jenkinsClient := NewClient(url, "u", "p")
	if err := jenkinsClient.CreateJob("job-name", fooJob); err == nil {
		t.Fatalf("JenkinsJobCreate() expecting an error, but received none\n")
	}
}

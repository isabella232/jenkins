package jenkins

import "testing"

var maven string = `<?xml version='1.0' encoding='UTF-8'?>
<maven2-moduleset plugin="maven-plugin@2.7.1">
    <description>Build maven</description>
    </maven2-moduleset>`

var freestyle string = `<?xml version='1.0' encoding='UTF-8'?>
<project>
  <description>Builds freestyle</description>
</project>`

var unknown string = `<?xml version='1.0' encoding='UTF-8'?>
<nope>
  <description>Builds freestyle</description>
</nope>`

func TestDocumentType(t *testing.T) {
	if jobType, _ := getJobType([]byte(maven)); jobType != Maven {
		t.Fatalf("Want Maven type but got %v\n", jobType)
	}

	if jobType, _ := getJobType([]byte(freestyle)); jobType != Freestyle {
		t.Fatalf("Want Freestyle type but got %v\n", jobType)
	}

	if jobType, _ := getJobType([]byte(unknown)); jobType != Unknown {
		t.Fatalf("Want Unknown type but got %v\n", jobType)
	}

	if jobType, _ := getJobType([]byte("")); jobType != Unknown {
		t.Fatalf("Want Unknown type but got %v\n", jobType)
	}

	if _, err := getJobType([]byte("{}")); err == nil {
		t.Fatalf("Expecting an error parsing not-XMLt %v\n", err)
	}
}

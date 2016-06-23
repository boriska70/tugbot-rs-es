package junitxml

import (
    "fmt"
    "strings"
    xj "github.com/basgys/goxml2json"
    "time"
    _"strconv"
    elastic "gopkg.in/olivere/elastic.v3"
    "net/http"
    "bytes"
)

var junitResultsClient *elastic.Client
var indexName = "tbresults"
var junitResultType = "junitxml"
var indexTemplate = `{"template":"tbresults*","mappings":{"_default_":{"dynamic_templates":[{"strings":{"match_mapping_type":"string","mapping":{"type":"string","index":"not_analyzed"}}},{"timestamp_field":{"match":"testsuites.testsuite.timestamp","mapping":{"type":"date"}}}]}}}`
var templateApplied = false;

var ch = make(chan string)

type Message struct {
    When  time.Time
    What  string
    Level string
}

func (m Message) Message() {
    if len(m.Level) == 0 {
	m.Level = "INFO"
    }
    fmt.Println(fmt.Sprintf("%v %s %s", m.When, strings.ToUpper(m.Level), m.What))
}

func push(s string) {

    r := strings.NewReader(s)
    json, err := xj.Convert(r)
//    fmt.Println(len(json.String()))
    json1 := strings.Replace(json.String(), "\"-", "\"", -1)
//    fmt.Println(len(json1))
    if err != nil {
	Message{When:time.Now(), Level:"Error", What:"Something is bad"}.Message()
	ch <- ""
    } else {
//	Message{When:time.Now(), Level: "Trace", What:"JSON is ready with length of " + strconv.Itoa(len(json1))}.Message()
	ch <- json1
    }

}

func prepareJUnitResultsClient() {
    if junitResultsClient == nil {
	junitResultsClient, _ = elastic.NewClient()
	exists, err := junitResultsClient.IndexExists("tbresults").Do()
	if err != nil {
	    fmt.Println("Problem when checking index")
	}
	if !exists {
	    applyTemplate()
	} else {
	    fmt.Println("Index exists, continue using")
	}
    }
}

//create template to prevent analyzing fields, if no index created yet
func applyTemplate() {
    var templateString = []byte(indexTemplate);
    req, err := http.NewRequest("POST", "http://localhost:9200" + "/_template/tbresults_template", bytes.NewBuffer(templateString))
    req.Header.Set("Content-Type", "application/json")
    httpClient := &http.Client{}
    resp, err := httpClient.Do(req)
    defer resp.Body.Close()
    if err != nil {
	panic(err)
    } else {
	fmt.Println("Template created for the index")
	templateApplied = true
    }
}

func HandleJUnitXml(input []byte) string{
    inp := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><testsuites><testsuite name=\"JUnitXmlReporter\" errors=\"0\" tests=\"0\" failures=\"0\" time=\"0\" timestamp=\"2016-06-21T10:23:58\" /><testsuite name=\"JUnitXmlReporter.constructor\" errors=\"0\" skipped=\"1\" tests=\"3\" failures=\"1\" time=\"0.006\" timestamp=\"2016-06-21T10:23:58\"><properties><property name=\"java.vendor\" value=\"Sun Microsystems Inc.\" /><property name=\"compiler.debug\" value=\"on\" /><property name=\"project.jdk.classpath\" value=\"jdk.classpath.1.6\" /></properties><testcase classname=\"JUnitXmlReporter.constructor\" name=\"should default path to an empty string\" time=\"0.006\"><failure message=\"test failure\">Assertion failed</failure></testcase><testcase classname=\"JUnitXmlReporter.constructor\" name=\"should default consolidate to true\" time=\"0\"><skipped /></testcase><testcase classname=\"JUnitXmlReporter.constructor\" name=\"should default useDotNotation to true\" time=\"0\" /></testsuite></testsuites>"
    go push(inp)
    v := <-ch
    defer func() {
	//	r := recover();
	//	if r != nil {
	//	    fmt.Println("Everything is done with ERRORS!", r)
	//	}
//	fmt.Println("Everything is done")
    }()

    if len(v) > 0 {
//	fmt.Println(v)
	//just for the case that Elastic wasn't available on startup
	if junitResultsClient == nil {
	    prepareJUnitResultsClient()
	}
	if (!templateApplied) {
	    applyTemplate()
	}
	indexResult, _ := junitResultsClient.Index().Index(indexName).Type(junitResultType).BodyJson(v).Do()
//	fmt.Println(errindex)
//	fmt.Println(indexResult.Id)
	return  indexResult.Id
    } else {
	Message{When:time.Now(), Level:"warning", What:"Empty output received"}.Message()
    }
    return  "bbb"
}



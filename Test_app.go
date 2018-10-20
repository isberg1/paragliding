package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const localProjectURLroot = "http://localhost:8080/"
const localProjectURLbase = "http://localhost:8080/paragliding/"
const localProjectURLinfo1 = "http://localhost:8080/paragliding/api"
const localProjectURLinfo2 = "http://localhost:8080/paragliding/api/"
const localProjectURLarray1 = "http://localhost:8080/paragliding/api/igc/"
const localProjectURLarray2 = "http://localhost:8080/paragliding/api/igc"
const validIgcURL1 = "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
const validIgcURL2 = "https://raw.githubusercontent.com/marni/goigc/master/testdata/optimize-long-flight-1.igc"

const contentType = "application/json"

// start the server so other tests can be run later
func Test_startServer(t *testing.T) {
	// start local server
	go main()
	time.Sleep(2 * time.Second)
}

// wait for server to be established
func Test_httpConnection(t *testing.T) {
	for i := 0; i < 5; i++ {
		_, err := http.Get(localProjectURLroot)
		// if connection was successful
		if err == nil {
			i = 5 // abort loop
		} else {
			time.Sleep(2 * time.Second)
		}
		// stop trying to connect
		if i == 4 {
			t.Error("error unable to connect to  ", localProjectURLroot, err)
		}
	}
}

// checks if correct http status code 404 is received from expected URL's
func Test_rubbishURL_local(t *testing.T) {
	expected := http.StatusNotFound

	temp := make([]string, 0)
	temp = append(
		temp,
		localProjectURLroot,
		localProjectURLbase,
		localProjectURLbase+"rubbish",
		localProjectURLinfo1+"rubbish",
		localProjectURLarray1+"rubbish",
	)
	// iterates over URL's and checks responds value
	for _, res := range temp {
		get, err := http.Get(res)
		if err != nil {
			t.Error("error getting content from URL", res, err)
		}
		defer get.Body.Close()
		if get.StatusCode != expected {
			t.Error("error incorrect status code from: ", res, err)
		}
	}
}

// Test_igcinfoapi_local check if the responding json is as expected
func Test_igcinfoapi_local(t *testing.T) {
	expectedSatusCode := http.StatusOK
	expectedVersion := unavalabeVersinNr
	expectedInfo := infoSting
	notExpectedUptime1 := ""
	notExpectedUptime2 := "PT"

	get, err := http.Get(localProjectURLinfo1)
	if err != nil {
		t.Error("error getting from URL", err)
	}
	defer get.Body.Close()

	res, err2 := ioutil.ReadAll(get.Body)
	if err2 != nil {
		t.Error("error reading body", err2)
	}
	var appInfo GetIgcinfoAPI

	// check if values are correct
	err3 := json.Unmarshal(res, &appInfo)
	if err3 != nil {
		t.Error("error umarshaling ", err3)
	}
	if get.StatusCode != expectedSatusCode {
		t.Error("error invalid http status code")
	}
	if appInfo.Version != expectedVersion {
		t.Error("error invalid version nr ")
	}
	if appInfo.Info != expectedInfo {
		t.Error("error invalid information string ")
	}
	if appInfo.Uptime == notExpectedUptime1 || appInfo.Uptime == notExpectedUptime2 {
		t.Error("error invalid uptime value ")
	}
}

// tries to post at an invalid URL
func Test_PostAtInvalidURL(t *testing.T) {
	expected := make([]int, 0)

	expected = append(expected,
		http.StatusMethodNotAllowed,
		http.StatusMethodNotAllowed,
		http.StatusNotFound,
		http.StatusNotFound,
	)

	temp := make([]string, 0)
	temp = append(
		temp,
		localProjectURLinfo1,
		localProjectURLinfo2,
		localProjectURLarray1+"rubbish",
		localProjectURLarray1+"rubbish/pilot",
	)

	jsonVar, err := json.Marshal(InputURL{URL: validIgcURL1})
	if err != nil {
		t.Error("error marshaling into json")
	}
	// iterate over slice with URL's
	for nr, res := range temp {
		post, err := http.Post(res, contentType, bytes.NewBuffer(jsonVar))
		if err != nil {
			t.Error("error unable to POST for:", res)
		}
		defer post.Body.Close()
		if post.StatusCode != expected[nr] {
			t.Error("error illegal POST permitted form: ", res, post.StatusCode)
		}
	}
}

// tries to post invalid content to correct URL
func Test_PostInvalidContent(t *testing.T) {
	expected := http.StatusBadRequest

	invalidIgcURL := make([]string, 0)
	invalidIgcURL = append(
		invalidIgcURL,
		"https://github.com/",
		"https://raw.githubusercontent.com/marni/goigc/master/testdata/parse-0-invalid-record.0.igc",
		"https://raw.githubusercontent.com/marni/goigc/master/testdata/parse-c-invalid-finish.0.igc",
	)
	// iterate over slice with URL's
	for _, res := range invalidIgcURL {
		jsonVar, err := json.Marshal(InputURL{URL: res})
		if err != nil {
			t.Error("error marshaling into json")
		}
		post, err2 := http.Post(localProjectURLarray1, contentType, bytes.NewBuffer(jsonVar))
		if err2 != nil {
			t.Error("error unsuccessful post attempt for ", res)
		}
		defer post.Body.Close()
		if post.StatusCode != expected {
			t.Error("error illegal POST permitted for: ", res)
		}
	} // test if IgcMap length is as expected
	if len(IgcMap) != 0 {
		t.Error("error invalid content posted in data structure: IgcMap")
	}
}

// tries to post valid content at valid URL
func Test_PostValidContent(t *testing.T) {
	expected := http.StatusCreated
	// IGC files URL's to be posted
	igcULR := make([]string, 0)
	igcULR = append(
		igcULR,
		validIgcURL1,
		validIgcURL2,
	)
	// URL to send POST content to
	postURL := make([]string, 0)
	postURL = append(
		postURL,
		localProjectURLarray1,
		localProjectURLarray2,
	)
	// for all IGC files
	for _, res := range igcULR {
		// make json object to be posted
		jsonVar, err := json.Marshal(InputURL{URL: res})
		if err != nil {
			t.Error("error marshaling into json")
		} // for all  URL to send POST content to
		for _, pURL := range postURL {
			// try posing
			post, err2 := http.Post(pURL, contentType, bytes.NewBuffer(jsonVar))
			if err2 != nil {
				t.Error("error unable to POST from: ", res)
			}
			defer post.Body.Close()
			if post.StatusCode != expected {
				t.Error("error legal post not permitted for: ", res)
			}
			// more validation tests
			err3 := validatePostResponse(post)
			if err3 != nil {
				t.Error("error ", err3, " from: ", res)
			}
		}
	}
	// check that nr of entries in IgcMap is correct
	if len(IgcMap) != (len(postURL) + len(igcULR)) {
		t.Error("error data structure does not contain expected nr of values: ", len(IgcMap))
	}
}

// checks that ioutil can read content, content can be unmarshaled and that content string is correct via regex
func validatePostResponse(p *http.Response) error {
	read, err := ioutil.ReadAll(p.Body)
	if err != nil {
		return errors.New("can not read content body ")
	}

	strc := ResponsID{}
	err2 := json.Unmarshal(read, &strc)
	if err2 != nil {
		return errors.New("can not unmarshal ")
	}

	test, err4 := regexp.MatchString(idPrefix+"[1-9]*", strc.ID)
	if err4 != nil {
		return errors.New("defective regex compilation ")
	}
	if !test {
		return errors.New("incorrect return ID string based on regex match ")
	}
	return nil
}

// tries to get the json array of all stored Igc file ID's
func Test_getAllIDs(t *testing.T) {
	expected := http.StatusOK

	get, err := http.Get(localProjectURLarray1)
	if err != nil {
		t.Error("error getting content from URL", localProjectURLarray1)
	}
	defer get.Body.Close()
	if get.StatusCode != expected {
		t.Error("error incorrect status code", get.StatusCode)
	}
	res, err2 := ioutil.ReadAll(get.Body)
	if err2 != nil {
		t.Error("error reading body", err2)
	}
	slice := make([]ResponsID, 0)
	err3 := json.Unmarshal(res, &slice)
	if err3 != nil {
		t.Error("error unable to unmarshal json array", err3)
	}
	if len(slice) != len(IgcMap) {
		t.Error("error not the same nr of objects in local and global data structures")
	}
}

// tries to get single fields form URL's
func Test_getFields(t *testing.T) {
	expected := http.StatusOK

	metaKey := make([]string, 0)
	metaKey = append(
		metaKey,
		"h_date",
		"pilot",
		"glider",
		"glider_id",
		"track_length",
	)

	for key := range IgcMap {
		for _, field := range metaKey {

			myURL := localProjectURLarray1 + key + "/" + field

			get, err := http.Get(myURL)
			if err != nil {
				t.Error("error getting from URL", err)
			}
			defer get.Body.Close()
			if get.StatusCode != expected {
				fmt.Println(get.StatusCode)
				t.Error("error invalid status code from", myURL)
			}
			res, err2 := ioutil.ReadAll(get.Body)
			if err2 != nil {
				t.Error("error reading body", err2)
			}
			str := string(res)
			if str == "" {
				t.Error("error illegal content has been successfully posted")
			}
			// if the field is "track_length"
			if field == metaKey[4] {
				// try to convert the field "track_length" into int
				str = strings.TrimSpace(str)
				_, err3 := strconv.Atoi(str)
				if err3 != nil {
					t.Error("error illegal value(not int) for", err3, field, str)
				}
			}
		}
	}

}
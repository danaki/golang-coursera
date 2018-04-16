package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type XMLRow struct {
	ID            int    `xml:"id"`
	GUID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type XMLRoot struct {
	Rows []XMLRow `xml:"row"`
}

func (x *XMLRow) Name() string {
	return x.FirstName + " " + x.LastName
}

func (x *XMLRow) toUser() User {
	return User{Id: x.ID, Name: x.Name(), Age: x.Age, About: x.About, Gender: x.Gender}
}

func (u User) String() string {
	return strconv.Itoa(u.Id)
}

type ById []User

func (a ById) Len() int           { return len(a) }
func (a ById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ById) Less(i, j int) bool { return a[i].Id < a[j].Id }

type ByAge []User

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }

type ByName []User

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func TimeoutedServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	}
}

func ErroneousServer(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "blabla://1", http.StatusFound)
}

func CurvedHandsServer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusInternalServerError)
}

func BadJsonServer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(""))
}

func UnknownBadRequestServer(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

func BadUserJsonResponseServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(""))
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("AccessToken") != "token" {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := r.FormValue("query")
	orderField := r.FormValue("order_field")

	if orderField == "" {
		orderField = "Name"
	}

	if orderField != "Id" && orderField != "Age" && orderField != "Name" {
		resp := SearchErrorResponse{Error: "ErrorBadOrderField"}
		js, _ := json.Marshal(resp)
		http.Error(w, "", http.StatusBadRequest)
		w.Write([]byte(js))
		return
	}

	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	var x XMLRoot
	xml.Unmarshal(b, &x)

	users := make([]User, 0)
	for _, row := range x.Rows {
		if !(strings.Contains(row.Name(), query) || strings.Contains(row.About, query)) {
			continue
		}

		users = append(users, row.toUser())
	}

	var by sort.Interface

	switch orderField {
	case "Id":
		by = ById(users)
	case "Age":
		by = ByAge(users)
	case "Name":
		by = ByName(users)
	}

	if orderBy == OrderByDesc {
		sort.Sort(by)
	} else if orderBy == OrderByAsc {
		sort.Sort(sort.Reverse(by))
	}

	max := offset + limit
	if max > len(users) {
		max = len(users)
	}

	js, err := json.Marshal(users[offset:max])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(js)
}

func TestSearchServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}

	req := SearchRequest{}

	_, err := cli.FindUsers(req)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestLimitNegativeProducesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{Limit: -1}

	_, err := cli.FindUsers(req)
	if err == nil {
		t.Errorf("Must be error")
	}
}

func TestLimitMoreThan25ProducesNoError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{Limit: 26}

	_, err := cli.FindUsers(req)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestNextPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{Offset: 30, Limit: 25}

	resp, err := cli.FindUsers(req)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(resp.Users) != 5 {
		t.Errorf("Offset doesn't work %v", len(resp.Users))
	}
}

func TestOffsetNegativeProducesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{Offset: -1}

	_, err := cli.FindUsers(req)
	if err == nil {
		t.Errorf("Must be error")
	}
}

func TestTimeoutProducesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(TimeoutedServer))
	defer ts.Close()

	//client.Timeout = time.Millisecond
	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{}

	_, err := cli.FindUsers(req)
	if err == nil {
		t.Errorf("Must be error")
	}
}

func TestOtherErrorProducesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(ErroneousServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{}

	_, err := cli.FindUsers(req)
	if err == nil {
		t.Errorf("Must be error")
	}
}

func TestUnknownTokenProducesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cli := &SearchClient{"blabla", ts.URL}
	req := SearchRequest{}

	_, err := cli.FindUsers(req)
	if err == nil {
		t.Errorf("Must be error")
	}
}

func TestBadOrderFieldError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{OrderField: "Blabla"}

	_, err := cli.FindUsers(req)

	if err == nil {
		t.Errorf("Must be error")
	}
}

func testErrorServer(t *testing.T, f func(w http.ResponseWriter, r *http.Request)) {
	ts := httptest.NewServer(http.HandlerFunc(f))
	defer ts.Close()

	cli := &SearchClient{"token", ts.URL}
	req := SearchRequest{}

	_, err := cli.FindUsers(req)
	if err == nil {
		t.Errorf("Must be error")
	}
}

func TestUnknownBadRequestServerError(t *testing.T) {
	testErrorServer(t, UnknownBadRequestServer)
}

func TestBadUserJsonResponseServerError(t *testing.T) {
	testErrorServer(t, BadUserJsonResponseServer)
}

func TestCurvedHandsServerError(t *testing.T) {
	testErrorServer(t, CurvedHandsServer)
}

func TestBadJsonServerError(t *testing.T) {
	testErrorServer(t, BadJsonServer)
}

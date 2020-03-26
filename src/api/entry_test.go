package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

type getEntriesResponse struct {
	Entries []database.Entry `json:"entries"`
}

const UserHasAccessEmail = "user1@user.com"
const UserNoAccessEmail = "user2@user.com"

func SetupUsers() {
	err := database.GetDB().Unscoped().Delete(database.User{})
	if err.Error != nil {
		fmt.Println(err.Error.Error())
	}
	var user1 = database.User{
		BaseModel: database.BaseModel{},
		Email:     UserHasAccessEmail,
		Password:  "azer",
	}
	database.Insert(&user1)
	var user2 = database.User{
		BaseModel: database.BaseModel{},
		Email:     UserNoAccessEmail,
		Password:  "azer",
	}
	database.Insert(&user2)
}

func TestGetEntries(t *testing.T) {
	database.GetDB().Unscoped().Delete(database.Entry{})
	SetupUsers()
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	println(user.Email, user.ID)
	for i := 0; i < 13; i++ {
		entry := database.Entry{
			PartialEntry: database.PartialEntry{
				Content: strconv.Itoa(i),
				Title:   "Entry " + strconv.Itoa(i),
			},
			UserID:user.ID,
		}
		_ = database.Insert(&entry)
	}

	t.Run("No limit", caseNoLimit)
	t.Run("Bad limit", caseBadLimit)
	t.Run("Bad page", caseBadPage)
	t.Run("Limit Ok", caseLimitOk)
}

func caseNoLimit(t *testing.T) {
	e := echo.New()
	request, err := http.NewRequest("GET", "/entries", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	err = GetEntries(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v", http.StatusOK, recorder.Code)
	}

	var response getEntriesResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error("Could not read response")
	}

	if len(response.Entries) != 10 {
		t.Errorf("Bad number of entries, expected %v, got %v", 10, len(response.Entries))
	}
	if response.Entries[3].Content != "" {
		t.Errorf("Expected empty content, got %v", response.Entries[0].Content)
	}
}

func caseBadLimit(t *testing.T) {
	e := echo.New()
	q:= make(url.Values)
	q.Set("limit", "@&é")
	request, err := http.NewRequest("GET", "/entries?" + q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	err = GetEntries(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v", http.StatusBadRequest, recorder.Code)
	}
}

func caseBadPage(t *testing.T) {
	e := echo.New()
	q:= make(url.Values)
	q.Set("page", "@&é")
	request, err := http.NewRequest("GET", "/entries?" + q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	err = GetEntries(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v", http.StatusBadRequest, recorder.Code)
	}
}

func caseLimitOk(t *testing.T) {
	e := echo.New()
	q:= make(url.Values)
	q.Set("page", "2")
	q.Set("limit", "3")
	request, err := http.NewRequest("GET", "/entries?" + q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	err = GetEntries(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v", http.StatusOK, recorder.Code)
	}

	var response getEntriesResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error("Could not read response")
	}

	if len(response.Entries) != 3 {
		t.Errorf("Bad number of entries, expected %v, got %v", 10, len(response.Entries))
	}

	if response.Entries[2].Title != "Entry 7" {
		t.Errorf("Bad Entry title, expected %v, got %v", "Entry 5", response.Entries[2].Title)
	}
}

var validEntry = database.Entry{
	BaseModel:    database.BaseModel{
		ID: 432,
	},
	PartialEntry: database.PartialEntry{
		Content: "The entry content",
		Title:   "The title",
	},
}

var invalidEntry = database.Entry{
	BaseModel:    database.BaseModel{
		ID: 432,
	},
	PartialEntry: database.PartialEntry{
		Content: "The entry content",
		Title:   "T",
	},
}

const invalidJSON = "{"

func TestDeleteEntry(t *testing.T) {
	t.Run("Ensure ok", testDeleteEntryOK)
	t.Run("Ensure ko", testDeleteEntryKO)
}

func runDeleteEntry(id uint, t *testing.T) *httptest.ResponseRecorder {
	e := echo.New()

	/*
		Very ugly fix caused by echo internal problems
		A maintainer suggests this
		https://github.com/labstack/echo/pull/1463#issuecomment-581107410
	*/
	r := e.Router()
	r.Add("DELETE", "/entries/:id", func(ctx echo.Context) error {return nil})

	request := httptest.NewRequest("DELETE", "/entries/" + strconv.Itoa(int(id)), nil)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(id)))

	err := DeleteEntry(context)
	if err != nil {
		t.Fatal(err)
	}

	return recorder
}

func testDeleteEntryOK(t *testing.T) {
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)

	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "",
			Title:   "The entry to remove title",
		},
		UserID: user.ID,
	}
	_ = database.Insert(&entry)
	recorder := runDeleteEntry(entry.ID, t)
	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v", http.StatusOK, recorder.Code)
	}

	var resultEntry database.Entry
	result := database.GetDB().Where("ID = ?", entry.ID).First(&resultEntry)
	if result.RecordNotFound() == false {
		t.Error("Record was not deleted from the database")
	}
}

func testDeleteEntryKO(t *testing.T) {
	recorder := runDeleteEntry(456545, t)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("Bad status, expected %v, got %v", http.StatusNotFound, recorder.Code)
	}
}

func TestEditEntry(t *testing.T) {
	t.Run("Valid arg", testEditValidEntry)
	t.Run("Invalid Arg", testEditInvalidEntry)
}

func runEditEntry(id uint, arg []byte, t *testing.T) *httptest.ResponseRecorder {
	e := echo.New()

	/*
	Very ugly fix caused by echo internal problems
	A maintainer suggests this
	https://github.com/labstack/echo/pull/1463#issuecomment-581107410
	 */
	r := e.Router()
	r.Add("PUT", "/entries/:id", func(ctx echo.Context) error {return nil})

	request := httptest.NewRequest("PUT", "/entries/" + strconv.Itoa(int(id)), bytes.NewReader(arg))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(id)))
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	err := EditEntry(context)
	if err != nil {
		t.Fatal(err)
	}

	return recorder
}

func testEditInvalidEntry(t *testing.T) {
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)

	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "",
			Title:   "The entry to edit title",
		},
		UserID:user.ID,
	}
	_ = database.Insert(&entry)
	marshall, _ := json.Marshal(invalidEntry)
	recorder := runEditEntry(entry.ID, marshall, t)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v", http.StatusBadRequest, recorder.Code)
	}

	var resultEntry database.Entry
	database.GetDB().Where("ID = ?", entry.ID).First(&resultEntry)

	if resultEntry.Title != "The entry to edit title" {
		t.Errorf("Bad title, got %v, expected %v", resultEntry.Title, "The entry to edit title")
	}
}

func testEditValidEntry(t *testing.T) {
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)

	//should be in setup code
	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "",
			Title:   "The entry to edit title",
		},
		UserID:user.ID,
	}
	_ = database.Insert(&entry)
	marshall, _ := json.Marshal(validEntry)
	recorder := runEditEntry(entry.ID, marshall, t)

	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v", http.StatusOK, recorder.Code)
	}

	var response response

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	var resultEntry database.Entry
	fmt.Println(response.Entry.ID)
	result := database.GetDB().Where("ID = ?", response.Entry.ID).First(&resultEntry)

	if result.RecordNotFound() {
		t.Error("Could not find created entry")
	}
	if resultEntry.Title != "The title" {
		t.Errorf("Bad title, got %v, expected %v", resultEntry.Title, "The title")
	}
}

func TestAddEntry(t *testing.T) {
	t.Run("Valid arg", testAddValidEntry)
	t.Run("Invalid arg", testAddInvalidEntry)
	t.Run("Invalid json", testAddInvalidJson)
}

func runAddEntry(arg []byte, t *testing.T) *httptest.ResponseRecorder {
	e := echo.New()
	request, _ := http.NewRequest("POST", "/entries", bytes.NewReader(arg))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	context.Set("user", user)

	err := AddEntry(context)
	if err != nil {
		t.Fatal(err)
	}

	return recorder
}

type response struct {
	Entry database.Entry `json:"entry"`
}

func testAddValidEntry(t *testing.T) {
	marshall, _ := json.Marshal(validEntry)
	recorder := runAddEntry(marshall, t)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Bad status, expected %v, got %v", http.StatusCreated, recorder.Code)
	}

	var response response

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	var entries []database.Entry
	result := database.GetDB().First(&entries).Where("ID = ?", response.Entry.ID)

	if result.RecordNotFound() {
		t.Error("Could not find created entry")
	}
	if response.Entry.Title != "The title" {
		t.Errorf("Bad title, got %v, expected %v", response.Entry.Title, "The title")
	}
}

func testAddInvalidEntry(t *testing.T) {
	marshall, _ := json.Marshal(invalidEntry)
	recorder := runAddEntry(marshall, t)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v", http.StatusBadRequest, recorder.Code)
	}

	str := "Validation failed on field: 'Title'. Expected to respect rule: 'min 3'. Got value: 'T'.\n"
	if recorder.Body.String() != str {
		t.Errorf("Bad status, expected (%v), got (%v)", str, recorder.Body.String())
	}
}

func testAddInvalidJson(t *testing.T) {
	recorder := runAddEntry([]byte(invalidJSON), t)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v", http.StatusBadRequest, recorder.Code)
	}
}
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	asserthelper "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

type getEntriesResponse struct {
	Entries []database.Entry `json:"entries"`
}

type getEntryResponse struct {
	Entry database.Entry `json:"entry"`
}

func TestGetEntry(t *testing.T) {
	SetupUsers()
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)

	//should be in setup code
	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "",
			Title:   "An awesome tile",
		},
		UserID:user.ID,
	}
	_ = database.Insert(&entry)

	e := echo.New()

	/*
		Very ugly fix caused by echo internal problems
		A maintainer suggests this
		https://github.com/labstack/echo/pull/1463#issuecomment-581107410
	*/
	r := e.Router()
	r.Add("DELETE", "/entries/:id", func(ctx echo.Context) error {return nil})
	request, err := http.NewRequest("GET", "/entries/" + strconv.Itoa(int(entry.ID)), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	context.Set("user", user)

	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(entry.ID)))

	err = GetEntry(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v (%v)", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response getEntryResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error("Could not read response")
	}

	if response.Entry.Title != "An awesome tile" {
		t.Errorf("Expected %v, got %v", "An awesome tile", response.Entry.Title)
	}
}

func TestGetEntries(t *testing.T) {
	database.GetDB().Unscoped().Delete(database.Entry{})
	SetupUsers()
	var user database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user)
	var label database.Label = database.Label{
		PartialLabel: database.PartialLabel{
			Name:  "Work",
			Color: "#FF0000",
		},
		UserID:       user.ID,
	}
	database.GetDB().Create(&label)
	for i := 0; i < 13; i++ {
		entry := database.Entry{
			PartialEntry: database.PartialEntry{
				Content: strconv.Itoa(i),
				Title:   "Entry " + strconv.Itoa(i),
			},
			UserID:user.ID,
			Labels: []database.Label{label},
		}
		_ = database.Insert(&entry)
	}

	t.Run("No limit", caseNoLimit)
	t.Run("Bad limit", caseBadLimit)
	t.Run("Bad page", caseBadPage)
	t.Run("Limit Ok", caseLimitOk)
}

func caseNoLimit(t *testing.T) {
	assert := asserthelper.New(t)

	context, recorder := BuildEchoContext([]byte(""))

	err := GetEntries(context)
	assert.Nil(err)

	assert.Equal(http.StatusOK, recorder.Code)

	var response getEntriesResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(err)
	assert.Equal(10, len(response.Entries))

	// By design we don't fill content
	assert.Equal("", response.Entries[3].Content)

	if assert.NotNil(response.Entries[0].Labels) {
		assert.Equal("Work", response.Entries[0].Labels[0].Name)
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
	assert := asserthelper.New(t)

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
	assert.Equal(http.StatusOK, recorder.Code)

	var response getEntriesResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error("Could not read response")
	}
	assert.Equal(3, len(response.Entries))

	assert.Equal("Entry 7", response.Entries[2].Title)
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
	SetupUsers()


	t.Run("Valid arg", testAddValidEntry)
	t.Run("Invalid arg", testAddInvalidEntry)
	t.Run("Invalid json", testAddInvalidJson)

	t.Run("Associate existing labels", testAssociateLabels)
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

func testAssociateLabels(t *testing.T) {
	assert := asserthelper.New(t)

	var user1 database.User
	database.GetDB().Where("email = ?", UserHasAccessEmail).First(&user1)
	var user2 database.User
	database.GetDB().Where("email = ?", UserNoAccessEmail).First(&user2)

	labelWorkUsr1 := database.Label{
		PartialLabel: database.PartialLabel{
			Name: "Work",
			Color: "#123456",
		},
		UserID:       user1.ID,
	}
	labelFamilyUsr1 := database.Label{
		PartialLabel: database.PartialLabel{
			Name: "Family",
			Color: "#654321",
		},
		UserID:       user1.ID,
	}
	labelFamilyUsr2 := database.Label{
		PartialLabel: database.PartialLabel{
			Name: "Family",
			Color: "#AABBCC",
		},
		UserID:       user2.ID,
	}
	database.GetDB().Create(&labelWorkUsr1)

	database.GetDB().Create(&labelFamilyUsr1)

	database.GetDB().Create(&labelFamilyUsr2)

	marshall, _ := json.Marshal(AddEntryRequestBody{
		PartialEntry: database.PartialEntry{
			Title: "Entry with labels",
		},
		LabelsID:     []uint{labelWorkUsr1.ID, labelFamilyUsr2.ID, labelFamilyUsr1.ID},
	})
	recorder := runAddEntry(marshall, t)

	assert.Equal(http.StatusCreated, recorder.Code, recorder.Body.String())

	var response response

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	var entries []database.Entry
	result := database.GetDB().First(&entries).Where("ID = ?", response.Entry.ID)

	assert.Equal(false, result.RecordNotFound())
	assert.Equal("Entry with labels", response.Entry.Title)

	assert.NotNil(response.Entry.Labels)
	assert.Equal(2, len(response.Entry.Labels))
	assert.Equal("#654321", response.Entry.Labels[1].Color)

}

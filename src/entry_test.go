package src

import (
	"bytes"
	"encoding/json"
	"github.com/Yuruh/encrypted-diary/src/database"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetEntries(t *testing.T) {
	e := echo.New()
	request, err := http.NewRequest("GET", "/entries", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	err = GetEntries(context)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("Bad status, expected %v, got %v", http.StatusOK, recorder.Code)
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

func TestEditEntry(t *testing.T) {

	t.Run("Valid arg", testEditValidEntry)
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
//	context.SetPath("/entries/:id")
	context.SetParamNames("id")
	context.SetParamValues(strconv.Itoa(int(id)))

	err := EditEntry(context)
	if err != nil {
		t.Fatal(err)
	}

	return recorder
}

func testEditValidEntry(t *testing.T) {
	//should be in setup code
	entry := database.Entry{
		PartialEntry: database.PartialEntry{
			Content: "",
			Title:   "The entry to edit title",
		},
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
	result := database.GetDB().First(&resultEntry).Where("ID = ?", response.Entry.ID)

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
}

func testAddInvalidJson(t *testing.T) {
	recorder := runAddEntry([]byte(invalidJSON), t)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Bad status, expected %v, got %v", http.StatusBadRequest, recorder.Code)
	}
}
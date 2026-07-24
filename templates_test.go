package todoist

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreateProjectFromFileMultipart(t *testing.T) {
	var gotPath, gotName, gotWorkspace, gotFile string
	var gotContentType string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		gotName = r.FormValue("name")
		gotWorkspace = r.FormValue("workspace_id")
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile: %v", err)
		} else {
			b, _ := io.ReadAll(f)
			gotFile = string(b)
		}
		_, _ = w.Write([]byte(`{"status":"ok","project_id":"p1","template_type":"csv"}`))
	})

	resp, err := c.CreateProjectFromFile(context.Background(), "My Project", Ptr("ws1"), strings.NewReader("TYPE,CONTENT\ntask,Buy milk\n"), "template.csv")
	if err != nil {
		t.Fatalf("CreateProjectFromFile: %v", err)
	}
	if gotPath != "/templates/create_project_from_file" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("content-type = %q", gotContentType)
	}
	if gotName != "My Project" || gotWorkspace != "ws1" {
		t.Errorf("name=%q workspace=%q", gotName, gotWorkspace)
	}
	if !strings.Contains(gotFile, "Buy milk") {
		t.Errorf("file content = %q", gotFile)
	}
	if resp.ProjectID != "p1" || resp.Status != "ok" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestImportIntoProjectFromFile(t *testing.T) {
	var gotProject string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(1 << 20)
		gotProject = r.FormValue("project_id")
		_, _ = w.Write([]byte(`{"status":"ok","template_type":"csv"}`))
	})

	if _, err := c.ImportIntoProjectFromFile(context.Background(), "p9", strings.NewReader("data"), "t.csv"); err != nil {
		t.Fatalf("ImportIntoProjectFromFile: %v", err)
	}
	if gotProject != "p9" {
		t.Errorf("project_id = %q", gotProject)
	}
}

func TestImportIntoProjectFromTemplateID(t *testing.T) {
	var gotPath string
	var body map[string]string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"status":"ok","template_type":"tpl"}`))
	})

	if _, err := c.ImportIntoProjectFromTemplateID(context.Background(), "p1", "tpl123", Ptr("en")); err != nil {
		t.Fatalf("ImportIntoProjectFromTemplateID: %v", err)
	}
	if gotPath != "/templates/import_into_project_from_template_id" {
		t.Errorf("path = %q", gotPath)
	}
	if body["project_id"] != "p1" || body["template_id"] != "tpl123" || body["locale"] != "en" {
		t.Errorf("body = %+v", body)
	}
}

func TestGetTemplateFileRawText(t *testing.T) {
	var gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("TYPE,CONTENT\ntask,Hello\n"))
	})

	out, err := c.GetTemplateFile(context.Background(), "p1", true)
	if err != nil {
		t.Fatalf("GetTemplateFile: %v", err)
	}
	if gotQuery != "project_id=p1&use_relative_dates=true" {
		t.Errorf("query = %q", gotQuery)
	}
	if !strings.Contains(out, "Hello") {
		t.Errorf("out = %q", out)
	}
}

func TestGetTemplateFileError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := c.GetTemplateFile(context.Background(), "p1", false); err == nil {
		t.Fatal("expected error from GetTemplateFile")
	}
}

func TestMultipartFileReadError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {})
	// errBody (from internal_test.go) fails on Read, so io.Copy in doMultipart errors.
	_, err := c.ImportIntoProjectFromFile(context.Background(), "p1", errBody{}, "t.csv")
	if err == nil {
		t.Fatal("expected file read error")
	}
}

func TestMultipartRequestError(t *testing.T) {
	c := New("t", WithBaseURL("http://%zz"))
	_, err := c.ImportIntoProjectFromFile(context.Background(), "p1", strings.NewReader("x"), "t.csv")
	if err == nil {
		t.Fatal("expected request construction error")
	}
}

func TestGetTemplateURL(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "project_id=p1" {
			t.Errorf("query = %q (use_relative_dates should be omitted when false)", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"file_name":"p1.csv","file_url":"https://x/p1.csv"}`))
	})

	out, err := c.GetTemplateURL(context.Background(), "p1", false)
	if err != nil {
		t.Fatalf("GetTemplateURL: %v", err)
	}
	if out.FileName != "p1.csv" || out.FileURL != "https://x/p1.csv" {
		t.Errorf("out = %+v", out)
	}
}

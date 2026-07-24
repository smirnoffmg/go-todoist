package todoist

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestUploadFileMultipart(t *testing.T) {
	var gotPath, gotName, gotProject, gotFile, gotContentType string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		gotName = r.FormValue("file_name")
		gotProject = r.FormValue("project_id")
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile: %v", err)
		} else {
			b, _ := io.ReadAll(f)
			gotFile = string(b)
		}
		_, _ = w.Write([]byte(`{"file_url":"https://x/f.txt","file_name":"f.txt","file_type":"text/plain","file_size":5,"resource_type":"file","upload_state":"completed"}`))
	})

	up, err := c.UploadFile(context.Background(), strings.NewReader("hello"), "f.txt", Ptr("p1"))
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if gotPath != "/uploads" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("content-type = %q", gotContentType)
	}
	if gotName != "f.txt" || gotProject != "p1" || gotFile != "hello" {
		t.Errorf("name=%q project=%q file=%q", gotName, gotProject, gotFile)
	}
	if up.FileURL != "https://x/f.txt" || up.FileSize != 5 {
		t.Errorf("up = %+v", up)
	}
}

func TestUploadResultAttachment(t *testing.T) {
	up := UploadResult{
		FileURL:      "https://x/f.txt",
		FileName:     "f.txt",
		FileType:     "text/plain",
		FileSize:     5,
		ResourceType: "file",
		UploadState:  "completed",
	}
	a := up.Attachment()
	if a.FileURL != up.FileURL || a.FileName != up.FileName || a.FileType != up.FileType ||
		a.FileSize != up.FileSize || a.ResourceType != up.ResourceType || a.UploadState != up.UploadState {
		t.Errorf("attachment = %+v", a)
	}
}

func TestDeleteUpload(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`"ok"`))
	})

	if err := c.DeleteUpload(context.Background(), "https://x/f.txt"); err != nil {
		t.Fatalf("DeleteUpload: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/uploads" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}
	if gotQuery != "file_url=https%3A%2F%2Fx%2Ff.txt" {
		t.Errorf("query = %q", gotQuery)
	}
}

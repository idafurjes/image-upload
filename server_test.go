package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUploadFile(t *testing.T) {
	cases := []struct {
		title          string
		filePath       string
		responseStatus int
	}{
		{
			title:          "success",
			filePath:       "test-data/penguin.png",
			responseStatus: http.StatusCreated,
		},
		{
			title:          "extension not allowed",
			filePath:       "test-data/penguin.csv",
			responseStatus: http.StatusBadRequest,
		},
	}

	for _, c := range cases {
		service := Service{imageDir: "test"}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(service.uploadFile)

		file, contentType := createMultipartFormData(t, "fileupload", c.filePath)
		req, err := http.NewRequest("POST", "/image", &file)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Add("Content-Type", contentType)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != c.responseStatus {
			t.Errorf("returned wrong status code: got %v want %v",
				status, c.responseStatus)
		}

		if rr.Body.String() != "" && rr.Code == http.StatusCreated {
			if !strings.Contains(rr.Body.String(), "id") {
				t.Errorf("returned unexpected body: got %v want id inside of the body",
					rr.Body.String())
			}
		}
	}
}

func TestAccessFile(t *testing.T) {
	cases := []struct {
		title          string
		fileID         string
		responseStatus int
		expResponse    []byte
	}{
		{
			title:          "success png",
			fileID:         "5577006791947779410",
			responseStatus: http.StatusOK,
			expResponse:    loadTestImage(t, "test-data/5577006791947779410.png"),
		},
		{
			title:          "not found",
			fileID:         "file_name",
			responseStatus: http.StatusNotFound,
			expResponse:    []byte("Not Found: File not found\n"),
		},
	}

	for _, c := range cases {
		service := Service{imageDir: "test-data"}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(service.accessFile)
		req, err := http.NewRequest("GET", "/image/"+c.fileID, nil)
		if err != nil {
			t.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != c.responseStatus {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, c.responseStatus)
		}

		if bytes.Compare(rr.Body.Bytes(), c.expResponse) != 0 {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), c.expResponse)
		}
	}
}

func loadTestImage(t *testing.T, filePath string) []byte {
	file := openTestImage(t, filePath)
	res, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("Couldn't read read file")
	}

	return res
}

func openTestImage(t *testing.T, filePath string) *os.File {
	file, err := os.Open(filePath)
	if err != nil {
		t.Errorf("Couldn't open test file: %v", err)
	}
	// defer file.Close()

	return file
}

func createMultipartFormData(
	t *testing.T,
	fieldName, fileName string,
) (bytes.Buffer, string) {
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	var fw io.Writer
	file := openTestImage(t, fileName)
	if fw, err = w.CreateFormFile(fieldName, file.Name()); err != nil {
		t.Errorf("Error creating writer: %v", err)
	}
	if _, err = io.Copy(fw, file); err != nil {
		t.Errorf("Error with io.Copy: %v", err)
	}

	w.Close()
	return b, w.FormDataContentType()
}

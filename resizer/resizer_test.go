package resizer

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"staply_img_resizer/config"
	"sync"
	"testing"
)

type clientMockGetImage struct {
}

func (c *clientMockGetImage) Get(str string) (*http.Response, error) {
	inputBuf, err := ioutil.ReadFile(str)
	if err != nil {
		return nil, err
	}
	resp := &http.Response{
		StatusCode: 200,
		Body: ioutil.NopCloser(
			bytes.NewReader(inputBuf)),
	}
	return resp, nil
}

func (c *clientMockGetImage) Head(str string) (*http.Response, error) {
	inputBuf, err := ioutil.ReadFile(str)
	if err != nil {
		return nil, err
	}
	resp := &http.Response{
		StatusCode:    200,
		ContentLength: int64(len(inputBuf)),
	}
	return resp, nil
}

func TestResize(t *testing.T) {
	config.Set(config.FileSaveDir, "test_out")
	os.MkdirAll(config.GetString(config.FileSaveDir), os.ModePerm)
	defer os.RemoveAll(config.GetString(config.FileSaveDir))
	// decoder wants []byte, so read the whole file into a buffer
	inputBuf, err := ioutil.ReadFile("test_data/test_image.jpg")
	if err != nil {
		t.Fatalf("failed to read input file, %s\n", err)
	}
	r := NewImgResizer()
	if err = r.ResizeImg(inputBuf); err != nil {
		t.Fatal(err)
	}
}

func TestImgFromUrl(t *testing.T) {
	config.Set(config.FileSaveDir, "test_out")
	os.MkdirAll(config.GetString(config.FileSaveDir), os.ModePerm)
	defer os.RemoveAll(config.GetString(config.FileSaveDir))

	r := NewImgResizer()
	client = &clientMockGetImage{}
	if err := r.FromUrl("test_data/test_image.jpg"); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkResizeAndSaveConcurency(b *testing.B) {
	config.Set(config.FileSaveDir, "test_out")
	os.MkdirAll(config.GetString(config.FileSaveDir), os.ModePerm)
	defer os.RemoveAll(config.GetString(config.FileSaveDir))

	r := NewImgResizer()
	inputBuf, _ := ioutil.ReadFile("test_data/test_image.jpg")
	b.N = 200
	wg := sync.WaitGroup{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			err := r.ResizeImg(inputBuf)
			if err != nil {
				b.Error(err.Error())
			}
			wg.Done()
		}()
		wg.Add(1)
	}
	wg.Wait()
	//b.StopTimer()
}

func BenchmarkResizeConcureny(b *testing.B) {
	inputBuf, _ := ioutil.ReadFile("test_data/test_image.jpg")
	b.N = 200
	inJob := imgJob{
		img: inputBuf,
		err: make(chan error),
	}
	inChan := make(chan imgJob)
	outChan := make(chan imgJob, b.N)
	startResizeWorkerPool(&sync.WaitGroup{}, inChan, outChan)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inChan <- inJob
	}
	for i := 0; i < b.N; i++ {
		<-outChan
	}
}

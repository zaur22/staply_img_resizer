package resizer

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"staply_img_resizer/config"
	"sync"
	"time"

	"github.com/davidbyttow/govips/pkg/vips"
	satori "github.com/satori/go.uuid"
)

func init() {
	client = config.HTTPClient
}

var client Client

type Client interface {
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
}

type Resizer interface {
	FromUrl(url string) error
	ResizeImg(img []byte) error
}

type ImgResizer struct {
	resizeChan     chan imgJob
	fileSaveChan   chan imgJob
	requestImgChan chan requestJob
	wg             sync.WaitGroup
}

type imgJob struct {
	img          []byte
	imgExtension string
	err          chan error
}

type requestJob struct {
	url string
	err chan error
}

//NewImgResizer создаёт resizer с запущенными воркерами и настраивает vips
func NewImgResizer() *ImgResizer {
	resizer := ImgResizer{
		resizeChan:     make(chan imgJob, config.GetInt(config.ResizeChannelSize)),
		fileSaveChan:   make(chan imgJob, config.GetInt(config.FileSaveChannelSize)),
		requestImgChan: make(chan requestJob, config.GetInt(config.RequestImgChannelSize)),
		wg:             sync.WaitGroup{},
	}

	log.Printf("Resize channel size: %v", config.GetInt(config.ResizeChannelSize))
	log.Printf("File save channel size: %v", config.GetInt(config.FileSaveChannelSize))
	log.Printf("Request image channel size: %v", config.GetInt(config.RequestImgChannelSize))

	vips.Startup(
		&vips.Config{
			ConcurrencyLevel: config.GetInt(config.VipsConcurrencyLevel),
			MaxCacheFiles:    config.GetInt(config.VipsMaxCashFiles),
			MaxCacheSize:     config.GetInt(config.VipsMaxCashSize),
			MaxCacheMem:      config.GetInt(config.VipsMaxCacheMem),
			CacheTrace:       config.GetBool(config.VipsCacheTrace),
			CollectStats:     config.GetBool(config.VipsCollectStats),
		},
	)

	startResizeWorkerPool(&resizer.wg, resizer.resizeChan, resizer.fileSaveChan)
	startFileSaveWorkerPool(&resizer.wg, resizer.fileSaveChan)
	startRequestImgWorkerPool(&resizer.wg, resizer.requestImgChan, resizer.resizeChan)
	return &resizer
}

func (r *ImgResizer) FromUrl(url string) error {
	var errChan = make(chan error, 1)
	r.requestImgChan <- requestJob{
		url: url,
		err: errChan,
	}
	select {
	case err := <-errChan:
		close(errChan)
		return err
	case <-time.After(time.Second * config.GetDuration(
		config.JobTimeoutSec)):
		close(errChan)
		return fmt.Errorf("Timout for request job")
	}
}

func (r *ImgResizer) ResizeImg(img []byte) error {
	var errChan = make(chan error, 1)
	r.resizeChan <- imgJob{
		img: img,
		err: errChan,
	}
	select {
	case err := <-errChan:
		close(errChan)
		return err
	case <-time.After(time.Second * config.GetDuration(
		config.JobTimeoutSec)):
		close(errChan)
		return fmt.Errorf("Timout for resize job")
	}
}

//Stop останавливает все воркеры и ждёт их завершения
func (r *ImgResizer) Stop() {
	close(r.resizeChan)
	close(r.requestImgChan)
	close(r.fileSaveChan)
	r.wg.Wait()
}

func resizeWorker(wg *sync.WaitGroup, in <-chan imgJob, out chan<- imgJob) {
	defer wg.Done()
	var err error
	var imgType vips.ImageType
	for job := range in {

		if len(job.img) == 0 {
			writeErr(job.err, fmt.Errorf("image is missing"))
			continue
		}

		job.img, imgType, err = vips.NewTransform().
			LoadBuffer(job.img).
			ResizeStrategy(vips.ResizeStrategyCrop).
			Resize(100, 100).
			OutputBytes().
			Apply()

		if err != nil {
			writeErr(job.err, fmt.Errorf("resize error: %v", err))
			continue
		}

		job.imgExtension = imgType.OutputExt()
		out <- job
	}
}

func fileSaveWorker(wg *sync.WaitGroup, in <-chan imgJob) {
	defer wg.Done()

	for job := range in {
		name, err := genName()
		if err != nil {
			writeErr(job.err, fmt.Errorf("can't gen name; error %v", err))
			continue
		}
		err = ioutil.WriteFile(
			path.Join(
				config.GetString(config.FileSaveDir),
				name+job.imgExtension),
			job.img,
			0644)
		if err != nil {
			writeErr(job.err, err)
			continue
		}
		writeErr(job.err, nil)
	}
}

func requestImgWorker(wg *sync.WaitGroup, in <-chan requestJob, out chan<- imgJob) {
	defer wg.Done()

	for job := range in {

		resp, err := client.Head(job.url)
		if err != nil {
			writeErr(job.err, err)
			continue
		}

		if resp.ContentLength > config.GetInt64(config.MaxImageSizeByte) {
			writeErr(job.err, fmt.Errorf("Image size is too large"))
			continue
		}

		resp, err = client.Get(job.url)
		if err != nil {
			writeErr(job.err, err)
			continue
		}

		if resp.StatusCode/100 != 2 {
			writeErr(job.err,
				fmt.Errorf("failed to get %s: status %d",
					job.url,
					resp.StatusCode),
			)
			continue
		}

		img := imgJob{
			err: job.err,
		}
		img.img, err = ioutil.ReadAll(resp.Body)

		if int64(len(img.img)) > config.GetInt64(config.MaxImageSizeByte) {
			writeErr(job.err, fmt.Errorf("Image size is too large."))
			continue
		}

		if err != nil {
			writeErr(job.err, err)
			continue
		}
		out <- img
		resp.Body.Close()
	}
}

func genName() (string, error) {
	var errCount int
	var err = fmt.Errorf("")
	var res string
	var uuid satori.UUID

	for errCount < 5 && err != nil {
		uuid, err = satori.NewV4()
		res = uuid.String()
		errCount++
	}
	return res, err
}

func writeErr(errChan chan<- error, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic: error channel closed. panic value: %v", r)
		}
	}()
	errChan <- err
}

func startResizeWorkerPool(wg *sync.WaitGroup, in <-chan imgJob, out chan<- imgJob) {
	for i := 0; i < config.GetInt(config.ResizeWorkerCount); i++ {
		go resizeWorker(wg, in, out)
		wg.Add(1)
	}
	log.Printf("The count of running resize workers: %v",
		config.GetInt(config.ResizeWorkerCount),
	)
}

func startFileSaveWorkerPool(wg *sync.WaitGroup, in <-chan imgJob) {
	for i := 0; i < config.GetInt(config.FileSaveWorkerCount); i++ {
		go fileSaveWorker(wg, in)
		wg.Add(1)
	}
	log.Printf("The count of running file save workers: %v",
		config.GetInt(config.FileSaveWorkerCount),
	)
}

func startRequestImgWorkerPool(wg *sync.WaitGroup, in <-chan requestJob, out chan<- imgJob) {
	for i := 0; i < config.GetInt(config.RequestImgWorkerCount); i++ {
		go requestImgWorker(wg, in, out)
		wg.Add(1)
	}
	log.Printf("The count of running request img workers: %v",
		config.GetInt(config.RequestImgWorkerCount),
	)
}

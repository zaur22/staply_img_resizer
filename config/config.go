package config

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
)

func GetInt(name string) int {
	return viper.GetInt(name)
}

func GetInt32(name string) int32 {
	return viper.GetInt32(name)
}

func GetInt64(name string) int64 {
	return viper.GetInt64(name)
}

func GetString(name string) string {
	return viper.GetString(name)
}

func GetBool(name string) bool {
	return viper.GetBool(name)
}

func GetDuration(name string) time.Duration {
	return viper.GetDuration(name)
}

func Set(name string, value interface{}) {
	viper.Set(name, value)
}

const (
	//ResizeWorkerCount количество запускаемых воркеров для ресайза
	ResizeWorkerCount = "resize_worker_count"

	//FileSaveWorkerCount количество запускаемых воркеров для сохранения изображения
	FileSaveWorkerCount = "file_save_worker_count"

	//RequestImgWorkerCount количество запускаемых воркеров для загрузки изображения по urls
	RequestImgWorkerCount = "request_img_worker_count"

	//ResizeChannelSize размер канала для задач по ресайзу
	ResizeChannelSize = "resize_channel_size"

	//FileSaveChannelSize размер канала для задач по сохранению файла
	FileSaveChannelSize = "file_save_channel_size"

	//RequestImgChannelSize размер канала для задач по загрузкам изображений по url
	RequestImgChannelSize = "request_img_channel_size"

	//JobTimeout таймаут ожидания выполнения задачи воркером в секундах
	JobTimeoutSec = "job_timeout_sec"

	//FileSaveDir директория для сохранения миниатюрок файлов
	FileSaveDir = "file_save_dir"

	//MaxImageSizeByte максимально допустимый размер изображения в байтах
	MaxImageSizeByte = "max_image_size_byte"

	//VipsConcurrencyLevel количество запущенных воркеров в vips. по умолчанию равен количеству ядер
	VipsConcurrencyLevel = "vips_concurrency_level"

	//VipsMaxCashFiles максимальнео количество файлов, которые vips будет кэшировать
	VipsMaxCashFiles = "vips_max_cache_files"

	//VipsMaxCashSize максимальное количество операций, которые будут храниться в vips
	VipsMaxCashSize = "vips_max_cache_size"

	//VipsMaxCacheMem максимальный размер кэша в байтах для vips
	VipsMaxCacheMem = "vips_max_cache_mem"

	//VipsCacheTrace включить трассировку кэша для vips
	VipsCacheTrace = "vips_cache_trace"

	//VipsCollectStats что за хрень сам не в курсе. По-моему что-то, что наблюдает за утечкой памяти.
	VipsCollectStats = "vips_collect_stats"

	// IdleConnTimeout таймаут сброса соединения при его неиспользовании
	IdleConnTimeoutSec = "idle_conn_timeout_sec"

	//MaxIdleConns максимальное количество одновременных подключений
	MaxIdleConns = "max_idle_conns"

	//MaxIdleConnsPerHost максимальное количество одновременнх подключений для одного хоста
	MaxIdleConnsPerHost = "max_idle_conns_per_host"
)

func init() {
	viper.SetDefault(ResizeWorkerCount, 10)
	viper.SetDefault(FileSaveWorkerCount, 10)
	viper.SetDefault(RequestImgWorkerCount, 100)
	viper.SetDefault(ResizeChannelSize,
		viper.GetInt(ResizeWorkerCount))
	viper.SetDefault(FileSaveChannelSize,
		viper.GetInt(FileSaveWorkerCount))
	viper.SetDefault(RequestImgChannelSize,
		viper.GetInt(RequestImgWorkerCount))
	viper.SetDefault(JobTimeoutSec, 10)
	viper.SetDefault(IdleConnTimeoutSec, 90)
	viper.SetDefault(MaxIdleConns, 100)
	viper.SetDefault(MaxIdleConnsPerHost, 100)
	viper.SetDefault(FileSaveDir, "")
	viper.SetDefault(MaxImageSizeByte, 15*1024*1024)

	HTTPClient = &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout:     time.Second * GetDuration(IdleConnTimeoutSec),
			MaxIdleConns:        GetInt(MaxIdleConns),
			MaxIdleConnsPerHost: GetInt(MaxIdleConnsPerHost),
		},
		Timeout: time.Second * GetDuration(JobTimeoutSec),
	}
}

//HTTPClient настроенный клиент
var HTTPClient *http.Client

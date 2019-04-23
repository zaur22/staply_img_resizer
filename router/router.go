package router

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"staply_img_resize/config"
	"staply_img_resize/resizer"
	"strconv"
)

type Router struct {
	Resizer resizer.Resizer
}

func NewRouter(r resizer.Resizer) Router {
	return Router{
		Resizer: r,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		router.imgFromUrl(w, r)
	case http.MethodPost:
		bodySize, _ := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if bodySize > config.GetInt64(config.MaxImageSizeByte) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Image size is too large."))
			return
		}

		mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch mt {
		case "multipart/form-data":
			router.imgFromMultiPart(w, r)
		case "application/json":
			router.imgFromJson(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Underfined content-type for POST method:" + r.Header.Get("Content-Type")))
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The path with this method is missing."))
	}
}

func (router *Router) imgFromUrl(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	urlVal := keys.Get("url")
	if urlVal == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required parameter 'url'"))
		return
	}

	err := router.Resizer.FromUrl(urlVal)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (router *Router) imgFromMultiPart(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Can't take image: " + err.Error()))
		return
	}

	img, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	err = router.Resizer.ResizeImg(img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (router *Router) imgFromJson(w http.ResponseWriter, r *http.Request) {
	var jsonImage struct {
		Image []byte `json:"image"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&jsonImage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if len(jsonImage.Image) == 0 {
		http.Error(w,
			"Field 'image' cannot be empty",
			http.StatusBadRequest,
		)
		return
	}

	err = router.Resizer.ResizeImg(jsonImage.Image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

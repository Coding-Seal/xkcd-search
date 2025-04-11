package handlers

import (
	"errors"
	"net/http"
	"strconv"

	http_util "yadro-go-course/pkg/http-util"

	"yadro-go-course/internal/core/services"
)

func Update(fetcherSrv *services.Fetcher) http_util.ErrHandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := fetcherSrv.Update(r.Context())
		if err != nil {
			return errors.Join(http_util.ErrInternal, err)
		}

		return nil
	}
}

type outputComic struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	ImgURL string `json:"img"`
}

func Search(searchSrv *services.Search) http_util.ErrHandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		phrase := r.URL.Query().Get("search")
		comics := searchSrv.SearchComics(r.Context(), phrase, 10)

		if len(comics) == 0 {
			return http_util.ErrNotFound
		}

		out := make([]outputComic, 0, len(comics))

		for _, comic := range comics {
			out = append(out, outputComic{ID: comic.ID, Title: comic.Title, ImgURL: comic.ImgURL})
		}

		err := http_util.WriteJson(w, out)
		if err != nil {
			return errors.Join(http_util.ErrInternal, err)
		}

		return nil
	}
}

func Comic(comicSrv *services.Comic) http_util.ErrHandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			return errors.Join(http_util.ErrBadRequest, err)
		}

		comic, err := comicSrv.Comic(r.Context(), id)
		if err != nil {
			return errors.Join(http_util.ErrInternal, err)
		}

		out := outputComic{ID: comic.ID, Title: comic.Title, ImgURL: comic.ImgURL}
		err = http_util.WriteJson(w, out)
		if err != nil {
			return errors.Join(http_util.ErrInternal, err)
		}

		return nil
	}
}

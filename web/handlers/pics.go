package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	httputil "yadro-go-course/pkg/http-util"

	"yadro-go-course/web/rest"
	"yadro-go-course/web/templates"
)

func Pics(c *rest.Client) httputil.ErrHandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		search := r.FormValue("query")

		comics, err := c.SearchPics(r.Context(), search)
		if err != nil {
			return errors.Join(err, httputil.ErrInternal)
		}

		favIDs := make([]int, 0)
		fav, err := r.Cookie("Favorites")
		if err == nil {
			err := json.Unmarshal([]byte(fav.Value), &favIDs)
			if err != nil {
				return errors.Join(err, httputil.ErrBadRequest)
			}
		}
		outComics := make([]templates.Comic, 0, len(comics))
		for _, c := range comics {
			outComics = append(outComics, templates.Comic{
				ID:       c.ID,
				ImgURL:   c.ImgURL,
				Title:    c.Title,
				Favorite: slices.Contains(favIDs, c.ID),
			})
		}

		err = templates.Pics(w, templates.PicsParams{
			Comics: outComics,
			Layout: templates.Layout{},
		})
		if err != nil {
			return errors.Join(err, httputil.ErrInternal)
		}

		return nil
	}
}

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	httputil "yadro-go-course/pkg/http-util"
	"yadro-go-course/web/rest"
	"yadro-go-course/web/templates"
)

func ToggleFavorites() httputil.ErrHandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return errors.Join(err, httputil.ErrBadRequest)
		}

		favIDs := make([]int, 0)
		fav, err := r.Cookie("Favorites")
		if err == nil {
			err := json.Unmarshal([]byte(fav.Value), &favIDs)
			if err != nil {
				return errors.Join(err, httputil.ErrBadRequest)
			}
		}

		if i := slices.Index(favIDs, id); i != -1 {
			favIDs = append(favIDs[:i], favIDs[i+1:]...)
		} else {
			favIDs = append(favIDs, id)
		}
		res, err := json.Marshal(favIDs)
		if err != nil {
			return errors.Join(err, httputil.ErrInternal)
		}
		cookie := &http.Cookie{
			Name:  "Favorites",
			Path:  "/",
			Value: string(res),
		}
		http.SetCookie(w, cookie)

		return nil
	}
}
func Favorites(c *rest.Client) httputil.ErrHandleFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		favIDs := make([]int, 0)
		fav, err := r.Cookie("Favorites")
		if err == nil {
			err := json.Unmarshal([]byte(fav.Value), &favIDs)
			if err != nil {
				return errors.Join(err, httputil.ErrBadRequest)
			}
		}
		outComics := make([]templates.Comic, 0, len(favIDs))
		for _, id := range favIDs {
			c, err := c.Comic(r.Context(), id)
			if err != nil {
				return errors.Join(err, httputil.ErrInternal)
			}
			outComics = append(outComics, templates.Comic{ID: c.ID, ImgURL: c.ImgURL, Title: c.Title, Favorite: true})
		}

		return templates.Favorites(w, templates.FavoritesParams{
			Comics: outComics,
			Layout: templates.Layout{},
		})
	}
}

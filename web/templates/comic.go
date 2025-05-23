package templates

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed *

var files embed.FS

func parse(file string) *template.Template {
	return template.Must(
		template.New("layout.html").ParseFS(files, "layout.html", file))
}

var (
	pics     = parse("comics.html")
	login    = parse("login.html")
	update   = parse("update.html")
	favorite = parse("favorites.html")
	main     = parse("main.html")
)

type Layout struct {
	PageTitle string
}
type Comic struct {
	ID       int
	Title    string
	ImgURL   string
	Favorite bool
}

type PicsParams struct {
	Comics []Comic
	Layout
}
type LoginParams struct {
	Layout
}
type UpdateParams struct {
	Layout
}
type MainParams struct {
	Layout
}

type FavoritesParams struct {
	Comics []Comic
	Layout
}

func Pics(w http.ResponseWriter, params PicsParams) error {
	return pics.Execute(w, params)
}

func Login(w http.ResponseWriter, params LoginParams) error {
	return login.Execute(w, params)
}

func Update(w http.ResponseWriter, params UpdateParams) error {
	return update.Execute(w, params)
}

func Favorites(w http.ResponseWriter, params FavoritesParams) error {
	return favorite.Execute(w, params)
}
func Main(w http.ResponseWriter, params MainParams) error {
	return main.Execute(w, params)
}

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alexsergivan/transliterator"
)

var tmplIndex = template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
var tmplGallery = template.Must(template.ParseFiles("templates/layout.html", "templates/gallery.html"))
var tmplArticles = template.Must(template.ParseFiles("templates/layout.html", "templates/articles.html"))
var tmplArticle = template.Must(template.ParseFiles("templates/layout.html", "templates/article.html"))
var tmplProjects = template.Must(template.ParseFiles("templates/layout.html", "templates/projects.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		FirstName, LastName, NickName, Quote string
		Age                                  int
		Description                          []string
		Links                                struct {
			GitHub, Telegram, TelegramChl, TelegramShitPost, TikTok, Discord, Email, Reddit, Snapchat string
		}
	}{
		FirstName: "Gregory",
		LastName:  "Mikhalkin",
		NickName:  "witttccchhher",
		Age:       16,
		Quote:     "you have every damn right to be fucked up tired, but you cant fucking quit",
		Description: []string{
			"go/python backend dev | linux junkie | rei ayanami simp | ngl kinda a dumb fuck vibin' with mignight rain and metal on blast",
			"also i really love dark souls II (peak)",
		},
		Links: struct {
			GitHub, Telegram, TelegramChl, TelegramShitPost, TikTok, Discord, Email, Reddit, Snapchat string
		}{
			GitHub:           "https://github.com/witttccchhher",
			Telegram:         "https://t.me/witttccchhher",
			TelegramChl:      "https://t.me/witttccchhher_blog",
			TelegramShitPost: "https://t.me/swampwithdicks",
			TikTok:           "https://tiktok.com/@witttccchhher",
			Discord:          "@witttccchhher",
			Email:            "mailto:somokill650@gmail.com",
			Reddit:           "https://reddit.com/user/somokill",
			Snapchat:         "https://snapchat.com/add/witttccchhher",
		},
	}

	if err := tmplIndex.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadImages(dir string) ([]string, error) {
	var images []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			ext := filepath.Ext(entry.Name())
			if ext == ".jpg" || ext == ".png" || ext == ".jpeg" || ext == ".gif" {
				images = append(images, entry.Name())
			}
		}
	}

	return images, nil
}

func shuffleStrings(slice []string) {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

func galleryHandler(w http.ResponseWriter, r *http.Request) {
	type GalleryData struct {
		Images []string
	}

	images, _ := loadImages("static/gallery")
	shuffleStrings(images)
	data := GalleryData{Images: images}

	if err := tmplGallery.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func slugify(title string) string {
	tr := transliterator.NewTransliterator(nil)

	slug := tr.Transliterate(title, "ru")
	slug = strings.ToLower(slug)

	forbidden := `'",.:;!?/%*@#№$^&()[]{}\|><+=`
	slug = strings.Map(func(r rune) rune {
		if strings.ContainsRune(forbidden, r) {
			return -1
		}
		return r
	}, slug)

	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}

func SimpleMarkdownRenderer(input string) string {
	blockCodeRegex := regexp.MustCompile("(?s)```(\\w*)\n(.+?)```")
	blocks := blockCodeRegex.FindAllStringSubmatchIndex(input, -1)

	var output strings.Builder
	lastIndex := 0

	for _, b := range blocks {
		textBefore := input[lastIndex:b[0]]
		textBefore = strings.ReplaceAll(textBefore, "\n", "<br>")
		output.WriteString(textBefore)

		lang := input[b[2]:b[3]]
		code := input[b[4]:b[5]]
		output.WriteString(fmt.Sprintf(`<pre><code class="language-%s">%s</code></pre>`, lang, code))

		lastIndex = b[1]
	}

	remainder := input[lastIndex:]
	remainder = strings.ReplaceAll(remainder, "\n", "<br>")
	output.WriteString(remainder)

	html := output.String()
	patterns := []struct {
		regex   *regexp.Regexp
		replace string
	}{
		{regexp.MustCompile(`\*\*(.+?)\*\*`), `<strong class="renderer">$1</strong>`},
		{regexp.MustCompile(`\*(.+?)\*`), `<em class="renderer">$1</em>`},
		{regexp.MustCompile(`~~(.+?)~~`), `<del class="renderer">$1</del>`},
		{regexp.MustCompile("`(.+?)`"), `<code class="renderer">$1</code>`},
		{regexp.MustCompile(`__(.+?)__`), `<u class="renderer">$1</u>`},
	}

	for _, p := range patterns {
		html = p.regex.ReplaceAllString(html, p.replace)
	}

	return html
}

func renderMarkdown(path string) (string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	mdContent := string(source)
	htmlContent := SimpleMarkdownRenderer(mdContent)

	return htmlContent, nil
}

type Article struct {
	Title string `json:"title"`
	Date  string `json:"date"`
	File  string `json:"file"`
	Slug  string
}

func loadArticles() ([]Article, error) {
	var articles []Article

	data, err := os.ReadFile("static/articles/articles.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &articles); err != nil {
		return nil, err
	}

	for i := range articles {
		articles[i].Slug = slugify(articles[i].Title)
	}

	return articles, nil
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	type ArticlesData struct {
		Articles []Article
	}

	articles, _ := loadArticles()
	data := ArticlesData{Articles: articles}

	if err := tmplArticles.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/articles/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	articles, err := loadArticles()
	if err != nil {
		http.Error(w, "Failed to load articles", http.StatusInternalServerError)
	}

	var found *Article
	for i := range articles {
		if articles[i].Slug == slug {
			found = &articles[i]
			break
		}
	}

	if found == nil {
		http.NotFound(w, r)
		return
	}

	htmlContent, err := renderMarkdown(found.File)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title       string
		Date        string
		HTMLContent template.HTML
	}{
		Title:       found.Title,
		Date:        found.Date,
		HTMLContent: template.HTML(htmlContent),
	}

	if err := tmplArticle.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type Project struct {
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Stars       int    `json:"stars"`
	Description string `json:"description"`
}

func loadProjects() ([]Project, error) {
	data, err := os.ReadFile("static/projects.json")
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	type ProjectsData struct {
		Projects []Project
	}

	projects, _ := loadProjects()
	data := ProjectsData{Projects: projects}

	if err := tmplProjects.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gallery", galleryHandler)
	http.HandleFunc("/articles", articlesHandler)
	http.HandleFunc("/articles/", articleHandler)
	http.HandleFunc("/projects", projectsHandler)

	const port string = ":8000"
	log.Printf("Server starting on port %v", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

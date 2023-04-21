package contacts

import (
	"net/http"
	"os"
	"strconv"

	"github.com/Masterminds/sprig/v3"

	"github.com/gin-contrib/multitemplate"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Server struct {
	db       *Db
	archiver *Archiver
}

func newServer() *Server {
	db := NewDb()
	err := db.LoadDB()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading contacts DB")
	}
	return &Server{
		db:       db,
		archiver: NewArchiver(),
	}
}

func createRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles(
		"index",
		"templates/layout.html",
		"templates/index.html",
		"templates/rows.html",
		"templates/archive_ui.html",
	)
	r.AddFromFiles("rows", "templates/rows.html")
	return r
}

func Run() {
	server := newServer()
	router := gin.Default()
	router.SetFuncMap(sprig.FuncMap())
	router.HTMLRender = createRenderer()
	router.Static("/static", "./static")

	sessionKey := os.Getenv("CONTACTS_SESSION_KEY")
	if sessionKey == "" {
		log.Fatal().Msg("CONTACTS_SESSION_KEY environment variable is required")
	}

	store := cookie.NewStore([]byte(sessionKey))
	router.Use(sessions.Sessions("mysession", store))

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/contacts")
	})

	router.GET("/contacts", server.contacts)

	//router.POST("/contacts/archive", server.startArchive)
	//
	//router.GET("/contacts/archive", server.archiveStatus)

	// Add other routes here as needed

	err := router.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("error running server")
	}
}

func (sr *Server) contacts(c *gin.Context) {
	search := c.Query("q")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.String(
			http.StatusBadRequest,
			"Invalid page number: %s",
			c.Query("page"),
		)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var contactsSet []*Contact
	log.Info().Int("page", page).Str("search", search).Msg("contacts query")

	if search != "" {
		contactsSet, err := sr.db.Search(search)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				"Invalid search text: %s",
				c.Query("q"),
			)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if c.Request.Header.Get("HX-Trigger") == "search" {
			c.HTML(http.StatusOK, "rows", gin.H{
				"contacts": contactsSet,
			})
			return
		}
	} else {
		contactsSet, err = sr.db.All(1)
	}

	c.HTML(http.StatusOK, "index", gin.H{
		"contacts":        contactsSet,
		"archiver":        sr.archiver,
		"requestArgsQ":    c.Query("q"),
		"flashedMessages": getFlashMessages(c),
	})
}

//func (sr *Server) startArchive(c *gin.Context) {
//	archiver := ArchiverGet()
//	archiver.Run()
//	c.HTML(http.StatusOK, "archive_ui.html", gin.H{
//		"archiver": archiver,
//	})
//}
//
//func (sr *Server) archiveStatus(c *gin.Context) {
//	archiver := ArchiverGet()
//	c.HTML(http.StatusOK, "archive_ui.html", gin.H{
//		"archiver": archiver,
//	})
//}

func flashMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.AddFlash(message)
	if err := session.Save(); err != nil {
		log.Error().Err(err).Str("message", message).Msg("error saving session")
	}
}

func getFlashMessages(c *gin.Context) []any {
	session := sessions.Default(c)
	flashes := session.Flashes()
	if len(flashes) != 0 {
		if err := session.Save(); err != nil {
			log.Printf("error in flashes saving session: %s", err)
		}
	}
	return flashes
}

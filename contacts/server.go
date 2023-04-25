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
	r.AddFromFiles("archive", "templates/archive_ui.html")
	r.AddFromFiles("new", "templates/layout.html", "templates/new.html")
	r.AddFromFiles("show", "templates/layout.html", "templates/show.html")
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
		c.Redirect(http.StatusFound, "/contacts")
	})
	router.GET("/contacts", server.contacts)
	router.POST("/contacts/archive", server.startArchive)
	router.GET("/contacts/archive", server.getArchiveStatus)
	router.DELETE("/contacts/archive", server.getArchiveStatus)
	router.GET("/contacts/archive/file", server.downloadArchiveFile)
	router.POST("/contacts/new", server.newContact)
	router.GET("/contacts/:contact_id", server.contactsView)

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

func (sr *Server) startArchive(c *gin.Context) {
	sr.archiver.Run()
	c.HTML(http.StatusOK, "archive", gin.H{
		"archiver": sr.archiver,
	})
}

func (sr *Server) getArchiveStatus(c *gin.Context) {
	c.HTML(http.StatusOK, "archive", gin.H{
		"archiver": sr.archiver,
	})
}

func (sr *Server) resetArchive(c *gin.Context) {
	sr.archiver.Reset()
	c.HTML(http.StatusOK, "archive", gin.H{
		"archiver": sr.archiver,
	})
}

func (sr *Server) newContact(c *gin.Context) {
	type ContactForm struct {
		First string `form:"first_name" binding:"required"`
		Last  string `form:"last_name"  binding:"required"`
		Phone string `form:"phone"      binding:"required"`
		Email string `form:"email"      binding:"required"`
	}
	var form ContactForm
	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, "bad request: %v", err)
		return
	}
	contact := NewContact(form.First, form.Last, form.Phone, form.Email)
	err := sr.db.SaveContact(contact)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			"error saving contact: %v",
			err,
		)
		c.HTML(http.StatusOK, "new", gin.H{
			"contact": contact,
		})
		return
	}
	flashMessage(c, "Created New Contact!")
	c.Redirect(http.StatusFound, "/contacts")
}

func (sr *Server) contactsView(c *gin.Context) {
	idStr := c.Param("contact_id")
	contactID, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			"Invalid contact id: %s",
			c.Param("contact_id"),
		)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	contact, err := sr.db.Find(contactID)
	if err != nil {
		c.String(
			http.StatusNotFound,
			"Contact not found: %s",
			idStr,
		)
		return
	}
	c.HTML(http.StatusOK, "show", gin.H{
		"contact": contact,
	})
}

func (sr *Server) downloadArchiveFile(c *gin.Context) {
	fileName := sr.archiver.ArchiveFile()
	curDir, err := os.Getwd()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	//content, err := os.ReadFile(fileName)
	//if err != nil {
	//	c.String(
	//		http.StatusInternalServerError,
	//		"Error reading archive file: %s: %v",
	//		fileName,
	//		err,
	//	)
	//	c.AbortWithStatus(http.StatusInternalServerError)
	//	return
	//}
	//c.Header("Content-Disposition", "attachment; filename="+fileName)
	//c.Header("Content-Type", "application/text/plain")
	//c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
	c.FileAttachment(curDir, fileName)
}

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

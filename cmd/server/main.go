package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/gommon/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var id = 0

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Contact struct {
	Id    int
	Name  string
	Email string
}

func NewContact(name, email string) Contact {
	id++
	return Contact{
		Id:    id,
		Name:  name,
		Email: email,
	}
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
}

func (d *Data) IndexOf(id int) int {
	for i, c := range d.Contacts {
		if c.Id == id {
			return i
		}
	}
	return -1
}

func (d *Data) HasEmail(email string) bool {
	for _, c := range d.Contacts {
		if c.Email == email {
			return true
		}
	}
	return false
}

func NewData() Data {
	return Data{
		Contacts: []Contact{
			NewContact("Alice", "alice@gmail.com"),
			NewContact("Bob", "bob@gmail.com"),
		},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func NewFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func NewPage() Page {
	return Page{
		Data: NewData(),
		Form: NewFormData(),
	}
}

func main() {
	e := echo.New()
	e.Renderer = NewTemplates()
	e.Use(middleware.Logger())
	e.Static("/images", "images")
	e.Static("/css", "css")

	page := NewPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.HasEmail(email) {
			formData := NewFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"
			page.Form = formData

			return c.Render(http.StatusUnprocessableEntity, "form", page.Form)
		}

		newContact := NewContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, newContact)

		err := c.Render(http.StatusOK, "form", NewFormData())
		if err != nil {
			log.Error(err)
		}
		return c.Render(http.StatusOK, "oob-contact", newContact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid ID")
		}

		index := page.Data.IndexOf(id)
		if index == -1 {
			return c.String(http.StatusNotFound, "Contact not found")
		}

		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)
		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":7777"))
}

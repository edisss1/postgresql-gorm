package main

import (
	"github.com/edisss1/pg-tutorial-go/models"
	"github.com/edisss1/pg-tutorial-go/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repo struct {
	DB *gorm.DB
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file" + err.Error())
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("Error connecting to database" + err.Error())
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("Error migrating books" + err.Error())
	}

	r := Repo{
		DB: db,
	}

	app := fiber.New()

	r.SetupRoutes(app)

	log.Fatal(app.Listen(":3000"))

}

func (r *Repo) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_books/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
	api.Patch("/update_title/:id", r.UpdateTitle)
	api.Get("/get_books_by_id", r.GetBooksByID)
}

func (r *Repo) CreateBook(c *fiber.Ctx) error {
	book := Book{}

	if err := c.BodyParser(&book); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(err.Error())
	}

	if err := r.DB.Create(&book).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}

	return c.Status(http.StatusCreated).JSON(book)

}

func (r *Repo) DeleteBook(c *fiber.Ctx) error {
	bookModel := &models.Book{}
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON("id is required")
	}
	if err := r.DB.Delete(bookModel, id).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"msg": "Book deleted successfully"})
}

func (r *Repo) GetBooks(c *fiber.Ctx) error {
	bookModels := &[]models.Book{}

	if err := r.DB.Find(bookModels).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"books": bookModels})

}

func (r *Repo) GetBookByID(c *fiber.Ctx) error {
	bookModel := &models.Book{}
	id := c.Params("id")

	if id == "" {
		return c.Status(http.StatusBadRequest).JSON("id is required")
	}

	if err := r.DB.Where("id = ?", id).First(bookModel).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"book": bookModel})

}

func (r *Repo) UpdateTitle(c *fiber.Ctx) error {
	id := c.Params("id")
	bookModel := &models.Book{}
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON("id is required")
	}

	var body struct {
		Title string `json:"title"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}

	if err := r.DB.Where("id = ?", id).First(bookModel).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}
	bookModel.Title = &body.Title

	if err := r.DB.Save(bookModel).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"book": bookModel})

}

// GetBooksByID for getting books where id is greater than 5
func (r *Repo) GetBooksByID(c *fiber.Ctx) error {
	bookModels := &[]models.Book{}

	if err := r.DB.Where("id > ? ", 5).Order("id desc").Find(bookModels).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(err.Error())
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"books": bookModels})
}

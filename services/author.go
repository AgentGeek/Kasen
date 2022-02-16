package services

import (
	"database/sql"
	"strings"

	. "kasen/database"

	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/rs1703/logger"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var AuthorCols = models.AuthorColumns

// This function simply calls CreateAuthorEx with the global Write connection.
func CreateAuthor(name string) (*modext.Author, error) {
	return CreateAuthorEx(WriteDB, name)
}

// CreateAuthorEx creates new author and returns it
// or the existing author if an author with the same name already exists.
func CreateAuthorEx(e boil.Executor, name string) (*modext.Author, error) {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return nil, errs.ErrAuthorNameRequired
	} else if len(name) > 128 {
		return nil, errs.ErrAuthorNameTooLong
	}

	a, err := models.Authors(Where("name ILIKE ?", name)).One(e)
	if err == sql.ErrNoRows {
		a = &models.Author{Name: name}
		if err = a.Insert(e, boil.Infer()); err != nil {
			logger.Err.Println(err)
			return nil, errs.ErrUnknown
		}
	} else if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewAuthor(a), nil
}

// This function simply calls GetAuthorsEx with the global Read connection.
func GetAuthors() ([]*modext.Author, error) {
	return GetAuthorsEx(ReadDB)
}

// GetAuthorsEx gets all authors, results are sorted by name in ascending order.
func GetAuthorsEx(e boil.Executor) ([]*modext.Author, error) {
	authors, err := models.Authors(OrderBy("name ASC")).All(e)
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	result := make([]*modext.Author, len(authors))
	for i, a := range authors {
		result[i] = modext.NewAuthor(a)
	}

	return result, nil
}

// This function simply calls GetAuthorEx with the global Read connection.
func GetAuthor(id int64) (*modext.Author, error) {
	return GetAuthorEx(ReadDB, id)
}

// GetAuthorEx gets an author.
func GetAuthorEx(e boil.Executor, id int64) (*modext.Author, error) {
	a, err := models.FindAuthor(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewAuthor(a), nil
}

// This function simply calls GetAuthorByNameEx with the global Read connection.
func GetAuthorByName(name string) (*modext.Author, error) {
	return GetAuthorByNameEx(ReadDB, name)
}

// GetAuthorByNameEx gets an author by name.
func GetAuthorByNameEx(e boil.Executor, name string) (*modext.Author, error) {
	a, err := models.Authors(Where("name ILIKE ?", name)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewAuthor(a), nil
}

// This function simply calls GetAuthorBySlugEx with the global Read connection.
func GetAuthorBySlug(slug string) (*modext.Author, error) {
	return GetAuthorBySlugEx(ReadDB, slug)
}

// GetAuthorBySlugEx gets an author by slug.
func GetAuthorBySlugEx(e boil.Executor, slug string) (*modext.Author, error) {
	a, err := models.Authors(Where("slug ILIKE ?", slug)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return modext.NewAuthor(a), nil
}

// This function simply calls GetAuthorBySlugOrNameEx with the global Read connection.
func GetAuthorBySlugOrName(slugOrName string) (*modext.Author, error) {
	return GetAuthorBySlugOrNameEx(ReadDB, slugOrName)
}

// GetAuthorBySlugOrNameEx gets an author by slug or name.
func GetAuthorBySlugOrNameEx(e boil.Executor, slugOrName string) (*modext.Author, error) {
	a, err := models.Authors(Where("slug ILIKE ? OR name ILIKE ?", slugOrName, slugOrName)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return modext.NewAuthor(a), nil
}

// This function simply calls UpdateAuthorEx with the global Write connection.
func UpdateAuthor(id int64, name string) (*modext.Author, error) {
	return UpdateAuthorEx(WriteDB, id, name)
}

// UpdateAuthorEx updates an author.
func UpdateAuthorEx(e boil.Executor, id int64, name string) (*modext.Author, error) {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return nil, errs.ErrAuthorNameRequired
	} else if len(name) > 128 {
		return nil, errs.ErrAuthorNameTooLong
	}

	a, err := models.FindAuthor(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	a.Name = name
	if err = a.Update(e, boil.Infer()); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewAuthor(a), nil
}

// This function simply calls DeleteAuthorEx with the global Write connection.
func DeleteAuthor(id int64) error {
	return DeleteAuthorEx(WriteDB, id)
}

// DeleteAuthorEx deletes an author.
func DeleteAuthorEx(e boil.Executor, id int64) error {
	a, err := models.FindAuthor(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = a.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteAuthorByNameEx with the global Write connection.
func DeleteAuthorByName(name string) error {
	return DeleteAuthorByNameEx(WriteDB, name)
}

// DeleteAuthorByNameEx deletes an author by name.
func DeleteAuthorByNameEx(e boil.Executor, name string) error {
	a, err := models.Authors(Where("name ILIKE ?", name)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = a.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteAuthorBySlugEx with the global Write connection.
func DeleteAuthorBySlug(slug string) error {
	return DeleteAuthorBySlugEx(WriteDB, slug)
}

// DeleteAuthorBySlugEx deletes an author by slug.
func DeleteAuthorBySlugEx(e boil.Executor, slug string) error {
	a, err := models.Authors(Where("slug ILIKE ?", slug)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = a.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteAuthorBySlugOrNameEx with the global Write connection.
func DeleteAuthorBySlugOrName(slugOrName string) error {
	return DeleteAuthorBySlugOrNameEx(WriteDB, slugOrName)
}

// DeleteAuthorBySlugOrNameEx deletes an author by slug or name.
func DeleteAuthorBySlugOrNameEx(e boil.Executor, slugOrName string) error {
	a, err := models.Authors(Where("slug ILIKE ? OR name ILIKE ?", slugOrName, slugOrName)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrAuthorNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = a.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

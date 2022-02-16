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

var TagCols = models.TagColumns

// This function simply calls CreateTagEx with the global Write connection.
func CreateTag(name string) (*modext.Tag, error) {
	return CreateTagEx(WriteDB, name)
}

// CreateTagEx creates a tag.
// Returns the existing tag if a tag with the same name already exists.
func CreateTagEx(e boil.Executor, name string) (*modext.Tag, error) {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return nil, errs.ErrTagNameRequired
	} else if len(name) > 32 {
		return nil, errs.ErrTagNameTooLong
	}

	t, err := models.Tags(Where("name ILIKE ?", name)).One(e)
	if err == sql.ErrNoRows {
		t = &models.Tag{Name: name}
		if err = t.Insert(e, boil.Infer()); err != nil {
			logger.Err.Println(err)
			return nil, errs.ErrUnknown
		}
	} else if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewTag(t), nil
}

// This function simply calls GetTagEx with the global Read connection.
func GetTag(id int64) (*modext.Tag, error) {
	return GetTagEx(ReadDB, id)
}

// GetTagEx gets a tag by id.
func GetTagEx(e boil.Executor, id int64) (*modext.Tag, error) {
	t, err := models.FindTag(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return modext.NewTag(t), nil
}

// This function simply calls GetTagByNameEx with the global Read connection.
func GetTagByName(name string) (*modext.Tag, error) {
	return GetTagByNameEx(ReadDB, name)
}

// GetTagByNameEx gets a tag by name.
func GetTagByNameEx(e boil.Executor, name string) (*modext.Tag, error) {
	t, err := models.Tags(Where("name ILIKE ?", name)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return modext.NewTag(t), nil
}

// This function simply calls GetTagBySlugEx with the global Read connection.
func GetTagBySlug(slug string) (*modext.Tag, error) {
	return GetTagBySlugEx(ReadDB, slug)
}

// GetTagBySlugEx gets a tag by.
func GetTagBySlugEx(e boil.Executor, slug string) (*modext.Tag, error) {
	t, err := models.Tags(Where("slug ILIKE ?", slug)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return modext.NewTag(t), nil
}

// This function simply calls GetTagBySlugOrNameEx with the global Read connection.
func GetTagBySlugOrName(slugOrName string) (*modext.Tag, error) {
	return GetTagBySlugOrNameEx(ReadDB, slugOrName)
}

// GetTagBySlugOrNameEx gets a tag by slug or name.
func GetTagBySlugOrNameEx(e boil.Executor, slugOrName string) (*modext.Tag, error) {
	t, err := models.Tags(Where("slug ILIKE ? OR name ILIKE ?", slugOrName, slugOrName)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return modext.NewTag(t), nil
}

// This function simply calls GetTagsEx with the global Read connection.
func GetTags() ([]*modext.Tag, error) {
	return GetTagsEx(ReadDB)
}

// GetTagsEx gets all tags ordered by name.
func GetTagsEx(e boil.Executor) ([]*modext.Tag, error) {
	tags, err := models.Tags(OrderBy("name ASC")).All(e)
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	result := make([]*modext.Tag, len(tags))
	for i, t := range tags {
		result[i] = modext.NewTag(t)
	}

	return result, nil
}

// This function simply calls UpdateTagEx with the global Write connection.
func UpdateTag(id int64, name string) (*modext.Tag, error) {
	return UpdateTagEx(WriteDB, id, name)
}

// UpdateTagEx updates a tag.
func UpdateTagEx(e boil.Executor, id int64, name string) (*modext.Tag, error) {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return nil, errs.ErrTagNameRequired
	} else if len(name) > 32 {
		return nil, errs.ErrTagNameTooLong
	}

	t, err := models.FindTag(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	t.Name = name
	if err = t.Update(e, boil.Infer()); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewTag(t), nil
}

// This function simply calls DeleteTagEx with the global Write connection.
func DeleteTag(id int64) error {
	return DeleteTagEx(WriteDB, id)
}

// DeleteTagEx deletes a tag.
func DeleteTagEx(e boil.Executor, id int64) error {
	t, err := models.FindTag(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = t.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteTagByNameEx with the global Write connection.
func DeleteTagByName(name string) error {
	return DeleteTagByNameEx(WriteDB, name)
}

// DeleteTagByNameEx deletes a tag by name.
func DeleteTagByNameEx(e boil.Executor, name string) error {
	t, err := models.Tags(Where("name ILIKE ?", name)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = t.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteTagBySlugEx with the global Write connection.
func DeleteTagBySlug(slug string) error {
	return DeleteTagBySlugEx(WriteDB, slug)
}

// DeleteTagBySlugEx deletes a tag by slug.
func DeleteTagBySlugEx(e boil.Executor, slug string) error {
	t, err := models.Tags(Where("slug ILIKE ?", slug)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = t.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteTagBySlugOrNameEx with the global Write connection.
func DeleteTagBySlugOrName(slugOrName string) error {
	return DeleteTagBySlugOrNameEx(WriteDB, slugOrName)
}

// DeleteTagBySlugOrNameEx deletes a tag by slug or name.
func DeleteTagBySlugOrNameEx(e boil.Executor, slugOrName string) error {
	t, err := models.Tags(Where("slug ILIKE ? OR name ILIKE ?", slugOrName, slugOrName)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrTagNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = t.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

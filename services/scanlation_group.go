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

var ScanlationGroupCols = models.ScanlationGroupColumns

// This function simply calls CreateScanlationGroupEx with the global Write connection.
func CreateScanlationGroup(name string) (*modext.ScanlationGroup, error) {
	return CreateScanlationGroupEx(WriteDB, name)
}

// CreateScanlationGroupEx creates a new scanlation group,
// and returns it or the existing scanlation group if
// a scanlation group with the same name already exists.
func CreateScanlationGroupEx(e boil.Executor, name string) (*modext.ScanlationGroup, error) {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return nil, errs.ErrScanlationGroupNameRequired
	} else if len(name) > 128 {
		return nil, errs.ErrScanlationGroupNameTooLong
	}

	g, err := models.ScanlationGroups(Where("name ILIKE ?", name)).One(e)
	if err == sql.ErrNoRows {
		g = &models.ScanlationGroup{Name: name}
		if err = g.Insert(e, boil.Infer()); err != nil {
			logger.Err.Println(err)
			return nil, errs.ErrUnknown
		}
	} else if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewScanlationGroup(g), nil
}

// This function simply calls GetScanlationGroupsEx with the global Read connection.
func GetScanlationGroups() ([]*modext.ScanlationGroup, error) {
	return GetScanlationGroupsEx(ReadDB)
}

// GetScanlationGroupsEx gets all scanlation groups, results are sorted by name in ascending order.
func GetScanlationGroupsEx(e boil.Executor) ([]*modext.ScanlationGroup, error) {
	groups, err := models.ScanlationGroups(OrderBy("name ASC")).All(e)
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	result := make([]*modext.ScanlationGroup, len(groups))
	for i, g := range groups {
		result[i] = modext.NewScanlationGroup(g)
	}
	return result, nil
}

// This function simply calls GetScanlationGroupEx with the global Read connection.
func GetScanlationGroup(id int64) (*modext.ScanlationGroup, error) {
	return GetScanlationGroupEx(ReadDB, id)
}

// GetScanlationGroupEx gets a scanlation group.
func GetScanlationGroupEx(e boil.Executor, id int64) (*modext.ScanlationGroup, error) {
	g, err := models.FindScanlationGroup(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewScanlationGroup(g), nil
}

// This function simply calls GetScanlationGroupByNameEx with the global Read connection.
func GetScanlationGroupByName(name string) (*modext.ScanlationGroup, error) {
	return GetScanlationGroupByNameEx(ReadDB, name)
}

// GetScanlationGroupByNameEx gets a scanlation group by name.
func GetScanlationGroupByNameEx(e boil.Executor, name string) (*modext.ScanlationGroup, error) {
	g, err := models.ScanlationGroups(Where("name ILIKE ?", name)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewScanlationGroup(g), nil
}

// This function simply calls GetScanlationGroupBySlugEx with the global Read connection.
func GetScanlationGroupBySlug(slug string) (*modext.ScanlationGroup, error) {
	return GetScanlationGroupBySlugEx(ReadDB, slug)
}

// GetScanlationGroupBySlugEx gets a scanlation group by slug.
func GetScanlationGroupBySlugEx(e boil.Executor, slug string) (*modext.ScanlationGroup, error) {
	g, err := models.ScanlationGroups(Where("slug ILIKE ?", slug)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewScanlationGroup(g), nil
}

// This function simply calls GetScanlationGroupBySlugOrNameEx with the global Read connection.
func GetScanlationGroupBySlugOrName(slugOrName string) (*modext.ScanlationGroup, error) {
	return GetScanlationGroupBySlugOrNameEx(ReadDB, slugOrName)
}

// GetScanlationGroupBySlugOrNameEx gets a scanlation group by slug or name.
func GetScanlationGroupBySlugOrNameEx(e boil.Executor, slugOrName string) (*modext.ScanlationGroup, error) {
	g, err := models.ScanlationGroups(Where("slug ILIKE ? OR name ILIKE ?", slugOrName, slugOrName)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewScanlationGroup(g), nil
}

// This function simply calls UpdateScanlationGroupEx with the global Write connection.
func UpdateScanlationGroup(id int64, name string) (*modext.ScanlationGroup, error) {
	return UpdateScanlationGroupEx(WriteDB, id, name)
}

// UpdateScanlationGroupEx updates a scanlation group.
func UpdateScanlationGroupEx(e boil.Executor, id int64, name string) (*modext.ScanlationGroup, error) {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return nil, errs.ErrScanlationGroupNameRequired
	} else if len(name) > 128 {
		return nil, errs.ErrScanlationGroupNameTooLong
	}

	g, err := models.FindScanlationGroup(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	g.Name = name
	if err = g.Update(e, boil.Infer()); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewScanlationGroup(g), nil
}

// This function simply calls DeleteScanlationGroupEx with the global Write connection.
func DeleteScanlationGroup(id int64) error {
	return DeleteScanlationGroupEx(WriteDB, id)
}

// DeleteScanlationGroupEx deletes a scanlation group.
func DeleteScanlationGroupEx(e boil.Executor, id int64) error {
	g, err := models.FindScanlationGroup(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = g.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteScanlationGroupBySlugEx with the global Write connection.
func DeleteScanlationGroupBySlug(slug string) error {
	return DeleteScanlationGroupBySlugEx(WriteDB, slug)
}

// DeleteScanlationGroupBySlugEx deletes a scanlation group by slug.
func DeleteScanlationGroupBySlugEx(e boil.Executor, slug string) error {
	g, err := models.ScanlationGroups(Where("slug ILIKE ?", slug)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = g.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteScanlationGroupByNameEx with the global Write connection.
func DeleteScanlationGroupByName(name string) error {
	return DeleteScanlationGroupByNameEx(WriteDB, name)
}

// DeleteScanlationGroupByNameEx deletes a scanlation group by name.
func DeleteScanlationGroupByNameEx(e boil.Executor, name string) error {
	g, err := models.ScanlationGroups(Where("name ILIKE ?", name)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = g.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// This function simply calls DeleteScanlationGroupBySlugOrNameEx with the global Write connection.
func DeleteScanlationGroupBySlugOrName(slugOrName string) error {
	return DeleteScanlationGroupBySlugOrNameEx(WriteDB, slugOrName)
}

// DeleteScanlationGroupBySlugOrNameEx deletes a scanlation group by slug or name.
func DeleteScanlationGroupBySlugOrNameEx(e boil.Executor, slugOrName string) error {
	g, err := models.ScanlationGroups(Where("slug ILIKE ? OR name ILIKE ?", slugOrName, slugOrName)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrScanlationGroupNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if err = g.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

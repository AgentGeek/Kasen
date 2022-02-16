package modext

import (
	"fmt"

	"kasen/models"
)

type Cover struct {
	ID        int64  `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	FileName  string `json:"fileName"`
}

func NewCover(cover *models.Cover) *Cover {
	if cover == nil {
		return nil
	}
	return &Cover{
		ID:        cover.ID,
		CreatedAt: cover.CreatedAt.Unix(),
		UpdatedAt: cover.UpdatedAt.Unix(),
		FileName:  cover.FileName,
	}
}

func (c *Cover) Path(p *Project) string {
	return fmt.Sprintf("covers/%s/%s", p.Slug, c.FileName)
}

package modext

import "kasen/models"

type ScanlationGroup struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`

	Chapters []*Chapter `json:"-"`
}

func NewScanlationGroup(scanlationGroup *models.ScanlationGroup) *ScanlationGroup {
	if scanlationGroup == nil {
		return nil
	}
	return &ScanlationGroup{
		ID:   scanlationGroup.ID,
		Slug: scanlationGroup.Slug,
		Name: scanlationGroup.Name,
	}
}

func (g *ScanlationGroup) LoadChapters(group *models.ScanlationGroup) *ScanlationGroup {
	if group == nil || group.R == nil || len(group.R.Chapters) == 0 {
		return g
	}

	g.Chapters = make([]*Chapter, len(group.R.Chapters))
	for i, chapter := range group.R.Chapters {
		exists := false
		for _, c := range g.Chapters {
			if c.ID == chapter.ID {
				exists = true
				break
			}
		}
		if !exists {
			g.Chapters[i] = NewChapter(chapter).LoadRels(chapter)
		}
	}

	return g
}

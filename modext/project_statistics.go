package modext

import "kasen/models"

type ProjectStats struct {
	*Statistics
	ProjectID int64 `json:"-"`
}

func NewProjectStats(stat *models.Statistic) *ProjectStats {
	if stat == nil {
		return nil
	}
	return &ProjectStats{
		Statistics: NewStatistics(stat),
		ProjectID:  stat.ProjectID.Int64,
	}
}

func (ProjectStats) CacheIdentifier() string {
	return "p"
}

func (ProjectStats) PrimaryKeyName() string {
	return "project_id"
}

func (p *ProjectStats) PrimaryKey() int64 {
	return p.ProjectID
}

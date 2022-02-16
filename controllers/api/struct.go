package api

type CreateEntityPayload struct {
	Name string `json:"name"`
}

type GetEntityQueries struct {
	ID   int64  `form:"id"`
	Slug string `form:"slug"`
	Name string `form:"name"`
}

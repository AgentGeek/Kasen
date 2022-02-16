package api

import (
	. "kasen/constants"
	"kasen/server"
)

func Init() {
	var (
		GET    = server.GET
		POST   = server.POST
		PATCH  = server.PATCH
		DELETE = server.DELETE

		WithAuthorization = server.WithAuthorization
		WithPermissions   = server.WithPermissions
		WithRateLimit     = server.WithRateLimit
	)

	PATCH("/api/config/meta",
		WithPermissions(PermManage), UpdateMeta)
	GET("/api/config/service",
		WithPermissions(PermManage),
		GetServiceConfig)
	PATCH("/api/config/service",
		WithPermissions(PermManage),
		UpdateServiceConfig)
	POST("/api/util/remap",
		WithPermissions(PermManage),
		RemapSymlinks)
	POST("/api/util/refreshTemplates",
		WithPermissions(PermManage),
		RefreshTemplates)

	POST("/api/author",
		WithPermissions(PermCreateProject, PermEditProject),
		CreateAuthor)
	DELETE("/api/author/:identifier",
		WithPermissions(PermCreateProject, PermEditProject),
		DeleteAuthor)
	GET("/api/author",
		WithAuthorization(nil),
		GetAuthor)
	GET("/api/authors",
		WithRateLimit("api-global", "5-S"),
		GetAuthors)

	POST("/api/project/:id/chapter",
		WithPermissions(PermCreateChapter),
		CreateChapter)
	DELETE("/api/chapter/:id",
		WithPermissions(PermDeleteChapter, PermDeleteChapters),
		DeleteChapter)
	GET("/api/chapter/:id",
		WithRateLimit("api-global", "5-S"),
		GetChapter)
	GET("/api/chapter/:id/stats",
		WithRateLimit("api-global", "5-S"),
		GetChapterStats)
	GET("/api/chapter",
		WithRateLimit("api-global", "5-S"),
		GetChapters)
	GET("/api/project/:id/chapters",
		WithRateLimit("api-global", "5-S"),
		GetChaptersByProject)
	PATCH("/api/chapter/:id/lock",
		WithPermissions(PermLockChapter, PermLockChapters),
		LockChapter)
	PATCH("/api/chapter/:id/publish",
		WithPermissions(PermPublishChapter, PermPublishChapters),
		PublishChapter)
	PATCH("/api/chapter/:id/unlock",
		WithPermissions(PermUnlockChapter, PermUnlockChapters),
		UnlockChapter)
	PATCH("/api/chapter/:id/unpublish",
		WithPermissions(PermUnpublishChapter, PermUnpublishChapters),
		UnpublishChapter)
	PATCH("/api/chapter/:id",
		WithPermissions(PermEditChapter, PermEditChapters),
		UpdateChapter)

	DELETE("/api/chapter/:id/pages/:fileName",
		WithPermissions(PermCreateChapter, PermEditChapter),
		DeletePage)
	GET("/api/chapter/:id/pages",
		WithRateLimit("api-global", "5-S"),
		GetPages)
	POST("/api/chapter/:id/pages",
		WithPermissions(PermCreateChapter, PermEditChapter),
		UploadPage)

	GET("/api/project/exists",
		WithAuthorization(nil),
		CheckProjectExists)
	POST("/api/project",
		WithPermissions(PermCreateProject),
		CreateProject)
	DELETE("/api/project/:id",
		WithPermissions(PermDeleteProject),
		DeleteProject)
	GET("/api/project/:id",
		WithRateLimit("api-global", "5-S"),
		GetProject)
	GET("/api/project/:id/stats",
		WithRateLimit("api-global", "5-S"),
		GetProjectStats)
	GET("/api/project",
		WithRateLimit("api-global", "5-S"),
		GetProjects)
	PATCH("/api/project/:id/lock",
		WithPermissions(PermLockProject),
		LockProject)
	PATCH("/api/project/:id/publish",
		WithPermissions(PermPublishProject),
		PublishProject)
	PATCH("/api/project/:id/unlock",
		WithPermissions(PermUnlockProject),
		UnlockProject)
	PATCH("/api/project/:id/unpublish",
		WithPermissions(PermUnpublishProject),
		UnpublishProject)
	PATCH("/api/project/:id",
		WithPermissions(PermEditProject),
		UpdateProject)

	DELETE("/api/project/:id/cover/:cid",
		WithPermissions(PermDeleteCover),
		DeleteCover)
	GET("/api/project/:id/cover",
		WithRateLimit("api-global", "5-S"),
		GetCover)
	GET("/api/project/:id/covers",
		WithRateLimit("api-global", "5-S"),
		GetCovers)
	PATCH("/api/project/:id/cover/:cid",
		WithPermissions(PermSetCover),
		SetCover)
	POST("/api/project/:id/cover",
		WithPermissions(PermUploadCover),
		UploadCover)

	POST("/api/scanlation_group",
		WithPermissions(PermCreateChapter, PermEditChapter),
		CreateScanlationGroup)
	DELETE("/api/scanlation_group/:identifier",
		WithPermissions(PermCreateChapter, PermEditChapter),
		DeleteScanlationGroup)
	GET("/api/scanlation_group",
		WithAuthorization(nil),
		GetScanlationGroup)
	GET("/api/scanlation_groups",
		WithRateLimit("api-global", "5-S"),
		GetScanlationGroups)

	POST("/api/tag",
		WithPermissions(PermCreateProject, PermEditProject),
		CreateTag)
	DELETE("/api/tag/:identifier",
		WithPermissions(PermCreateProject, PermEditProject),
		DeleteTag)
	GET("/api/tag",
		WithAuthorization(nil),
		GetTag)
	GET("/api/tags",
		WithRateLimit("api-global", "5-S"),
		GetTags)

	DELETE("/api/user",
		WithPermissions(PermDeleteUser),
		DeleteUser)
	DELETE("/api/user/:id",
		WithPermissions(PermDeleteUsers),
		DeleteUserById)
	GET("/api/user",
		WithAuthorization(nil),
		GetUser)
	GET("/api/users",
		WithPermissions(PermEditUsers, PermDeleteUsers, PermManage),
		GetUsers)
	PATCH("/api/user/name",
		WithPermissions(PermEditUser),
		UpdateUserName)
	PATCH("/api/user/:id/name",
		WithPermissions(PermEditUsers),
		UpdateUserNameById)
	PATCH("/api/user/password",
		WithPermissions(PermEditUser),
		UpdateUserPassword)
	PATCH("/api/user/:id/password",
		WithPermissions(PermEditUsers),
		UpdateUserPasswordById)
	PATCH("/api/user/:id/permissions",
		WithPermissions(PermManage),
		UpdateUserPermissions)

	GET("/api/stats/pages",
		WithPermissions(PermManage),
		GetPagesCacheStats)
	GET("/api/stats/project",
		WithPermissions(PermManage),
		GetProjectCacheStats)
	GET("/api/stats/cover",
		WithPermissions(PermManage),
		GetCoverCacheStats)
	GET("/api/stats/chapter",
		WithPermissions(PermManage),
		GetChapterCacheStats)

	GET("/api/md/chapter/:id",
		WithAuthorization(nil),
		GetChapterMd)
	GET("/api/md/chapter/:id/pages",
		WithAuthorization(nil),
		GetPagesMd)
	GET("/api/md/project/:id",
		WithAuthorization(nil),
		GetProjectMd)
}

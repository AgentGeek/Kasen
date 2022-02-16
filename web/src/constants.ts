export const Limit = 20;
export const MaxPages = 10;

export enum ProjectCols {
  ProjectStatus = "projectStatus",
  SeriesStatus = "seriesStatus",
  Demographic = "demographic",
  Rating = "rating"
}

export const ProjectColsKeys = Object.keys(ProjectCols);
export const ProjectColsValues = Object.values(ProjectCols);

export enum Demographic {
  None = "none",
  Shounen = "shounen",
  Shoujo = "shoujo",
  Josei = "josei",
  Seinen = "seinen"
}

export const DemographicKeys = Object.keys(Demographic);
export const DemographicValues = Object.values(Demographic);

export enum Permission {
  CreateProject = "create_project",
  EditProject = "edit_project",
  PublishProject = "publish_project",
  UnpublishProject = "unpublish_project",
  LockProject = "lock_project",
  UnlockProject = "unlock_project",
  DeleteProject = "delete_project",

  UploadCover = "upload_cover",
  SetCover = "set_cover",
  DeleteCover = "delete_cover",

  CreateChapter = "create_chapter",
  EditChapter = "edit_chapter",
  EditChapters = "edit_chapters",
  PublishChapter = "publish_chapter",
  PublishChapters = "publish_chapters",
  UnpublishChapter = "unpublish_chapter",
  UnpublishChapters = "unpublish_chapters",
  LockChapter = "lock_chapter",
  LockChapters = "lock_chapters",
  UnlockChapter = "unlock_chapter",
  UnlockChapters = "unlock_chapters",
  DeleteChapter = "delete_chapter",
  DeleteChapters = "delete_chapters",

  EditUser = "edit_user",
  EditUsers = "edit_users",
  DeleteUser = "delete_user",
  DeleteUsers = "delete_users",

  Manage = "manage"
}

export const PermissionKeys = Object.keys(Permission);
export const PermissionValues = Object.values(Permission);

export enum ProjectStatus {
  Ongoing = "ongoing",
  Finished = "finished",
  Dropped = "dropped"
}

export const ProjectStatusKeys = Object.keys(ProjectStatus);
export const ProjectStatusValues = Object.values(ProjectStatus);

export enum Rating {
  None = "none",
  Safe = "safe",
  Suggestive = "suggestive",
  Erotica = "erotica",
  Pornographic = "pornographic"
}

export const RatingKeys = Object.keys(Rating);
export const RatingValues = Object.values(Rating);

export enum SeriesStatus {
  Ongoing = "ongoing",
  Completed = "completed",
  Hiatus = "hiatus",
  Cancelled = "cancelled"
}

export const SeriesStatusKeys = Object.keys(SeriesStatus);
export const SeriesStatusValues = Object.values(SeriesStatus);

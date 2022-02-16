import { Permission } from "./constants";

export default {
  projects: {
    path: "/projects",
    name: "Projects",
    permissions: [
      Permission.CreateProject,
      Permission.EditProject,
      Permission.PublishProject,
      Permission.UnpublishProject,
      Permission.LockProject,
      Permission.UnlockProject,
      Permission.DeleteProject,
      Permission.UploadCover,
      Permission.SetCover,
      Permission.DeleteCover,
      Permission.CreateChapter,
      Permission.EditChapter,
      Permission.EditChapters,
      Permission.PublishChapter,
      Permission.PublishChapters,
      Permission.UnpublishChapter,
      Permission.UnpublishChapters,
      Permission.LockChapter,
      Permission.LockChapters,
      Permission.UnlockChapter,
      Permission.UnlockChapters,
      Permission.DeleteChapter,
      Permission.DeleteChapters
    ]
  },
  chapters: {
    path: "/chapters",
    name: "Chapters",
    permissions: [
      Permission.CreateChapter,
      Permission.EditChapter,
      Permission.EditChapters,
      Permission.PublishChapter,
      Permission.PublishChapters,
      Permission.UnpublishChapter,
      Permission.UnpublishChapters,
      Permission.LockChapter,
      Permission.LockChapters,
      Permission.UnlockChapter,
      Permission.UnlockChapters,
      Permission.DeleteChapter,
      Permission.DeleteChapters
    ]
  },
  settings: {
    path: "/settings",
    name: "Settings",
    permissions: []
  }
};

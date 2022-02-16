export enum Order {
  ASC = "asc",
  DESC = "desc"
}

export const OrderKeys = Object.keys(Order);
export const OrderValues = Object.values(Order);

export enum ChapterPreloads {
  Project = "project",
  Uploader = "uploader",
  ScanlationGroups = "scanlationGroups",
  Statistic = "statistic"
}

export enum ChapterSort {
  ID = "id",
  CreatedAt = "created_at",
  UpdatedAt = "updated_at",
  PublishedAt = "published_at",
  Chapter = "chapter",
  Volume = "volume",
  Title = "title"
}

export const ChapterSortKeys = Object.keys(ChapterSort);
export const ChapterSortValues = Object.values(ChapterSort);

export enum ProjectPreloads {
  Cover = "cover",
  Artists = "artists",
  Authors = "authors",
  Statistic = "Statistic",
  Tags = "tags"
}

export enum ProjectSort {
  ID = "id",
  CreatedAt = "created_at",
  UpdatedAt = "updated_at",
  PublishedAt = "published_at",
  Title = "title"
}

export const ProjectSortKeys = Object.keys(ProjectSort);
export const ProjectSortValues = Object.values(ProjectSort);

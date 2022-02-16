import { Demographic, ProjectStatus, Rating, SeriesStatus } from "../constants";
import { Order, ProjectPreloads, ProjectSort } from "./enums";
import SendRequest from "./xhr";

interface CheckProjectOptions {
  id?: number;
  slug?: string;
  title?: string;
}

export const CheckProjectExists = ({ id, slug, title }: CheckProjectOptions) =>
  SendRequest<{ id: string; slug: string }>(
    "GET",
    `/api/project/exists?id=${id || ""}&slug=${slug || ""}&title=${title || ""}`
  );

export const CreateProject = (draft: ProjectDraft) =>
  SendRequest<Project>("POST", `/api/project`, JSON.stringify(draft));

export const DeleteProject = (id: number) => SendRequest("DELETE", `/api/project/${id}`);

interface GetProjectOptions {
  preloads?: ProjectPreloads[];
  includesDrafts?: boolean;
}

export const GetProject = (id: number, o: GetProjectOptions) => {
  const searchParams = new URLSearchParams();

  if (o.preloads?.length) o.preloads.forEach(preload => searchParams.append("preload", preload));
  if (o.includesDrafts) searchParams.set("includesDrafts", "true");

  let url = `/api/project/${id}`;
  const params = searchParams.toString();
  if (params.length) url += `?${params}`;

  return SendRequest<Project>("GET", url);
};

export const GetProjectMd = (id: string) => SendRequest<ProjectDraft>("GET", `/api/md/project/${id}`);

interface GetProjectsOptions {
  title?: string;

  projectStatus?: ProjectStatus;
  seriesStatus?: SeriesStatus;
  demographic?: Demographic;
  rating?: Rating;

  excludedProjectStatus?: string[];
  excludedSeriesStatus?: string[];
  excludedDemographic?: string[];
  excludedRating?: string[];

  artists?: string[];
  authors?: string[];
  tags?: string[];

  limit?: number;
  offset?: number;
  preloads?: ProjectPreloads[];
  sort?: ProjectSort;
  order?: Order;
  includesDrafts?: boolean;
}

interface GetProjectsResult {
  data?: Project[];
  total?: number;
}

export const GetProjects = (o: GetProjectsOptions) => {
  const searchParams = new URLSearchParams();

  if (o.title) searchParams.set("title", o.title);
  if (o.projectStatus) searchParams.set("projectStatus", o.projectStatus);
  if (o.demographic) searchParams.set("demographic", o.demographic);
  if (o.rating) searchParams.set("rating", o.rating);

  if (o.excludedProjectStatus?.length)
    o.excludedProjectStatus.forEach(v => searchParams.append("excludeProjectStatus", v));
  if (o.excludedSeriesStatus?.length)
    o.excludedSeriesStatus.forEach(v => searchParams.append("excludeSeriesStatus", v));
  if (o.excludedDemographic?.length) o.excludedDemographic.forEach(v => searchParams.append("excludeDemographic", v));
  if (o.excludedRating?.length) o.excludedRating.forEach(v => searchParams.append("excludeRating", v));

  if (o.artists?.length) o.artists.forEach(v => searchParams.append("artist", v));
  if (o.authors?.length) o.authors.forEach(v => searchParams.append("author", v));
  if (o.tags?.length) o.tags.forEach(v => searchParams.append("tag", v));

  if (o.limit > 0) searchParams.set("limit", o.limit.toString());
  if (o.offset >= 0) searchParams.set("offset", o.offset.toString());
  if (o.sort) searchParams.set("sort", o.sort);
  if (o.order) searchParams.set("order", o.order);

  if (o.preloads?.length) o.preloads.forEach(rel => searchParams.append("preload", rel));
  if (o.includesDrafts) searchParams.set("includesDrafts", "true");

  let url = "/api/project";
  const params = searchParams.toString();
  if (params.length) url += `?${params}`;

  return SendRequest<GetProjectsResult>("GET", url);
};

export const LockProject = (id: number) => SendRequest<Project>("PATCH", `/api/project/${id}/lock`);

export const PublishProject = (id: number) => SendRequest<Project>("PATCH", `/api/project/${id}/publish`);

export const UnlockProject = (id: number) => SendRequest<Project>("PATCH", `/api/project/${id}/unlock`);

export const UnpublishProject = (id: number) => SendRequest<Project>("PATCH", `/api/project/${id}/unpublish`);

export const UpdateProject = (id: number, draft: ProjectDraft) =>
  SendRequest<Project>("PATCH", `/api/project/${id}`, JSON.stringify(draft));

import { ChapterPreloads, ChapterSort, Order } from "./enums";
import SendRequest from "./xhr";

export const CreateChapter = (projectId: number, draft: ChapterDraft) =>
  SendRequest<Chapter>("POST", `/api/project/${projectId}/chapter`, JSON.stringify(draft));

export const DeleteChapter = (id: number) => SendRequest("DELETE", `/api/chapter/${id}`);

interface GetChapterOptions {
  preloads?: ChapterPreloads[];
  includesDrafts?: boolean;
}

export const GetChapter = (id: number, o: GetChapterOptions = {}) => {
  const searchParams = new URLSearchParams();

  if (o.preloads?.length) o.preloads.forEach(preload => searchParams.append("preload", preload));
  if (o.includesDrafts) searchParams.set("includesDrafts", "true");

  let url = `/api/chapter/${id}`;
  const params = searchParams.toString();
  if (params.length) url += `?${params}`;

  return SendRequest<Chapter>("GET", url);
};

export const GetChapterMd = (id: string) => SendRequest<ChapterDraft>("GET", `/api/md/chapter/${id}`);

interface GetChaptersOptions {
  uploader?: string;
  scanlationGroups?: string[];
  limit?: number;
  offset?: number;
  preloads?: ChapterPreloads[];
  sort?: ChapterSort;
  order?: Order;
  includesDrafts?: boolean;
}

export const GetChapters = (o: GetChaptersOptions = {}) => {
  const searchParams = new URLSearchParams();

  if (o.uploader) searchParams.set("uploader", o.uploader);
  if (o.scanlationGroups?.length) {
    o.scanlationGroups.forEach(group => searchParams.append("scanlation_group", group));
  }

  if (o.limit > 0) searchParams.set("limit", o.limit.toString());
  if (o.offset >= 0) searchParams.set("offset", o.offset.toString());
  if (o.sort) searchParams.set("sort", o.sort);
  if (o.order) searchParams.set("order", o.order);

  if (o.preloads?.length) o.preloads.forEach(preload => searchParams.append("preload", preload));
  if (o.includesDrafts) searchParams.set("includesDrafts", "true");

  let url = `/api/chapter`;
  const params = searchParams.toString();
  if (params.length) url += `?${params}`;

  return SendRequest<{ data?: Chapter[]; total?: number }>("GET", url);
};

export const GetChaptersByProject = (projectId: number, o: GetChaptersOptions) => {
  const searchParams = new URLSearchParams();

  if (o.limit > 0) searchParams.set("limit", o.limit.toString());
  if (o.offset >= 0) searchParams.set("offset", o.offset.toString());
  if (o.sort) searchParams.set("sort", o.sort);
  if (o.order) searchParams.set("order", o.order);

  if (o.preloads?.length) o.preloads.forEach(preload => searchParams.append("preload", preload));
  if (o.includesDrafts) searchParams.set("includesDrafts", "true");

  let url = `/api/project/${projectId}/chapters`;
  const params = searchParams.toString();
  if (params.length) url += `?${params}`;

  return SendRequest<{ data?: Chapter[]; total?: number }>("GET", url);
};

export const LockChapter = (id: number) => SendRequest<Chapter>("PATCH", `/api/chapter/${id}/lock`);

export const PublishChapter = (id: number) => SendRequest<Chapter>("PATCH", `/api/chapter/${id}/publish`);

export const UnlockChapter = (id: number) => SendRequest<Chapter>("PATCH", `/api/chapter/${id}/unlock`);

export const UnpublishChapter = (id: number) => SendRequest<Chapter>("PATCH", `/api/chapter/${id}/unpublish`);

export const UpdateChapter = (id: number, draft: ChapterDraft) =>
  SendRequest<Chapter>("PATCH", `/api/chapter/${id}`, JSON.stringify(draft));

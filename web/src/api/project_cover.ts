import SendRequest from "./xhr";

export const DeleteCover = (projectId: number, coverId: number) =>
  SendRequest("DELETE", `/api/project/${projectId}/cover/${coverId}`);

export const GetCover = (projectId: number) => SendRequest<Cover>("GET", `/api/project/${projectId}/cover`);

export const GetCovers = (projectId: number) => SendRequest<Cover[]>("GET", `/api/project/${projectId}/covers`);

export const SetCover = (projectId: number, coverId: number) =>
  SendRequest("PATCH", `/api/project/${projectId}/cover/${coverId}`);

interface UploadCoverOptions {
  data?: File;
  url?: string;
  isInitialCover?: boolean;
  setAsMainCover?: boolean;
}

export const UploadCover = (projectId: number, o: UploadCoverOptions) => {
  const formData = new FormData();

  if (o.data) formData.set("data", o.data);
  if (o.url) formData.set("url", o.url);
  if (o.isInitialCover) formData.set("isInitialCover", "true");
  if (o.setAsMainCover) formData.set("setAsMainCover", "true");

  return SendRequest<Cover>("POST", `/api/project/${projectId}/cover`, formData);
};

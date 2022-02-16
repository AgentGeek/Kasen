import SendRequest from "./xhr";

export const DeletePage = (chapterId: number, fileName: string) =>
  SendRequest<string[]>("DELETE", `/api/chapter/${chapterId}/pages/${fileName}`);

export const GetPages = (chapterId: number) => SendRequest<string[]>("GET", `/api/chapter/${chapterId}/pages`);

export const GetPagesMd = (chapterId: string) =>
  SendRequest<{
    baseUrl: string;
    hash: string;
    pages: string[];
  }>("GET", `/api/md/chapter/${chapterId}/pages`);

interface UploadPageOptions {
  data?: File;
  url?: string;
}

export const UploadPage = (chapterId: number, o: UploadPageOptions) => {
  const formData = new FormData();

  if (o.data) formData.set("data", o.data);
  if (o.url) formData.set("url", o.url);

  return SendRequest<string[]>("POST", `/api/chapter/${chapterId}/pages`, formData);
};

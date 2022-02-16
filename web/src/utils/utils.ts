import { MaxPages } from "../constants";

export const HasPerms = (user: User, ...perms: string[]) =>
  perms.some(p => user.permissions.some(perm => perm.toLowerCase() === p.toLowerCase()));

export const FormatUnix = (unix: number) => {
  const time = new Date(unix * 1000);
  const month = (time.getMonth() + 1).toString().padStart(2, "0");
  const date = time.getDate().toString().padStart(2, "0");
  const year = time.getFullYear().toString().slice(-2);
  const hours = time.getHours().toString().padStart(2, "0");
  const minutes = time.getMinutes().toString().padStart(2, "0");
  return `${month}/${date}/${year} ${hours}:${minutes}`;
};

export const FormatChapter = (chapter: Chapter) => {
  let str = "";
  if (chapter.volume) {
    str += `Vol. ${chapter.volume} `;
  }
  str += `Ch. ${chapter.chapter}`;
  if (chapter.title) {
    str += ` - ${chapter.title}`;
  }
  return str;
};

const kb = 1024;
const units = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

export const FormatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 Bytes";
  const i = Math.floor(Math.log(bytes) / Math.log(kb));
  return `${(bytes / kb ** i).toFixed(1)} ${units[i]}`;
};

export const GetCoverURL = (p: Project, c: Cover): string => `/covers/${p.slug}/${c.fileName}`;

export const GetPageURL = (c: Chapter, fileName: string): string => `/pages/${c.id}/${fileName}`;

export const CreatePagination = (currentPage: number, totalPages: number) => {
  if (currentPage < 1) {
    currentPage = 1;
  } else if (currentPage > totalPages) {
    currentPage = totalPages;
  }

  let first: number;
  let last: number;

  if (totalPages <= MaxPages) {
    first = 1;
    last = totalPages;
  } else {
    const min = Math.floor(MaxPages / 2);
    const max = Math.ceil(MaxPages / 2) - 1;
    if (currentPage <= min) {
      first = 1;
      last = MaxPages;
    } else if (currentPage + max >= totalPages) {
      first = totalPages - MaxPages + 1;
      last = totalPages;
    } else {
      first = currentPage - min;
      last = currentPage + max;
    }
  }

  return Array.from(Array(last + 1 - first).keys()).map(i => first + i);
};

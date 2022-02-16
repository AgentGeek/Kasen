interface Api<T> extends Promise<ApiResult<T>> {
  addEventListener<K extends keyof XMLHttpRequestEventMap>(
    type: K,
    listener: (this: XMLHttpRequest, ev: XMLHttpRequestEventMap[K]) => any
  ): void;
  upload: {
    addEventListener<K extends keyof XMLHttpRequestEventTargetEventMap>(
      type: K,
      listener: (this: XMLHttpRequestUpload, ev: XMLHttpRequestEventTargetEventMap[K]) => any
    ): void;
  };
  cancel(): void;
}

declare interface ApiError {
  message?: string;
  cause?: string;
}

declare interface ApiResult<T> {
  response?: T;
  error?: ApiError;
  status: number;
}

declare interface User {
  id: number;
  name: string;
  email?: string;
  permissions: string[];
}

interface Entity {
  id: string;
  slug: string;
  name: string;
}

interface GetEntityOptions {
  id?: number;
  slug?: string;
  name?: string;
}

declare type Author = Entity;
declare type ScanlationGroup = Entity;
declare type Tag = Entity;

declare interface Project {
  id: number;
  slug: string;
  locked?: boolean;
  createdAt: number;
  updatedAt: number;
  publishedAt?: number;
  title: string;
  description?: string;
  projectStatus: string;
  seriesStatus: string;
  demographic?: string;
  rating?: string;
  artists?: Author[];
  authors?: Author[];
  tags?: Tag[];
  cover?: Cover;
  covers?: Cover[];
  chapters?: Chapter[];
  stats?: Statistics;
}

declare interface ProjectDraft {
  title: string;
  description?: string;
  coverUrl?: string;
  projectStatus: string;
  seriesStatus: string;
  demographic?: string;
  rating?: string;
  artists?: string[];
  authors?: string[];
  tags?: string[];
}

declare interface Cover {
  id: number;
  createdAt: number;
  updatedAt: number;
  fileName: string;
}

declare interface Chapter {
  id: number;
  locked?: boolean;
  createdAt: number;
  updatedAt: number;
  publishedAt?: number;
  chapter: string;
  volume?: string;
  title?: string;
  pages?: string[];
  project?: Project;
  uploader?: User;
  scanlationGroups: ScanlationGroup[];
  stats?: Statistics;
}

declare interface ChapterDraft {
  chapter: string;
  volume?: string;
  title?: string;
  scanlationGroups?: string[];
}

declare interface Statistics {
  viewCount?: number;
}

declare interface Menu {
  id: number;
  name: string;
  url: string;
  priority: number;
}

declare interface ViewData {
  baseURL: string;
  title: string;
  description: string;
  language: string;

  user: User;
}

declare const viewData: ViewData;

declare interface ReaderData {
  chapter?: Chapter;
  chapters?: Chapter[];
  pagination: ReaderPagination;
}

declare interface ReaderPagination {
  Current?: Chapter;
  Previous?: Chapter;
  Next?: Chapter;
}
declare const readerData: ReaderData;

declare interface Route {
  path: string;
  name: string;
  permissions: string[];
}

type Dispatcher<T = any> = React.Dispatch<React.SetStateAction<T>>;
type Props<T = any> = React.DetailedHTMLProps<React.HTMLAttributes<T>, T>;

declare interface Mutable<T> {
  current: T;
}

type Renderer = () => void;

declare interface PageState {
  isViewing?: boolean;
  isDownloaded?: boolean;
  isDownloading?: boolean;
  isFailed?: boolean;

  ref: Mutable<HTMLDivElement>;

  url?: string;
  fileName: string;
  index: number;
}

declare interface ServiceConfig {
  disableRegistration: boolean;
  coverMaxFileSize: number;
  pageMaxFileSize: number;
}

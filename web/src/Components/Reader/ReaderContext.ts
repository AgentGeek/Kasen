import { createContext } from "react";

export default createContext<{
  pref: Preference;

  chapter?: Chapter;
  chapters?: Chapter[];
  pagination?: ReaderPagination;

  pagesRef: Mutable<PageState[]>;
  currentPageRef: Mutable<PageState>;

  isLoadingRef: Mutable<boolean>;
  render: Renderer;
}>(undefined);

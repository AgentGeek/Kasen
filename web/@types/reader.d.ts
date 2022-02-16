import { PageDirection, PageScale, SidebarPosition } from "../src/Components/Reader/constants";

declare global {
  interface Preference {
    showSidebar: boolean;
    sidebarPosition: SidebarPosition;

    navigateOnClick: boolean;
    direction: PageDirection;
    scale: PageScale;
    maxWidth: string;
    maxHeight: string;
    gaps: string;
    zoom: string;

    maxPreloads: number;
    maxParallel: number;

    keybinds: Keybinds;
  }

  interface Keybinds {
    previousChapter: string;
    nextChapter: string;
    previousPage: string;
    nextPage: string;
  }
}

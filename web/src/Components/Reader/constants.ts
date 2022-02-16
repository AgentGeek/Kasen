export enum PageDirection {
  TopToBottom = 1,
  RightToLeft,
  LeftToRight
}

export enum PageScale {
  Default,
  Original,
  Width,
  Height,
  Stretch,
  FitWidth,
  FitHeight,
  StretchWidth,
  StretchHeight
}

export enum SidebarPosition {
  Left = 1,
  Right
}

export const SidebarPositionKeys = Object.keys(SidebarPosition);
export const SidebarPositionValues = Object.values(SidebarPosition);

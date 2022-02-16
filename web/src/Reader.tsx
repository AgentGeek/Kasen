import React from "react";
import { render } from "react-dom";
import { BrowserRouter as Router } from "react-router-dom";
import Reader from "./Components/Reader";
import { PageDirection, PageScale, SidebarPosition } from "./Components/Reader/constants";
import "./styles/reader.less";

const initReader = () => {
  if (!readerData) return;
  const target = document.querySelector(".reader");
  if (!target) return;

  let pref: Preference = {
    showSidebar: false,
    sidebarPosition: SidebarPosition.Left,

    navigateOnClick: true,
    direction: PageDirection.LeftToRight,
    scale: PageScale.Default,
    maxWidth: "800",
    maxHeight: "0",
    gaps: "20",
    zoom: "1.0",

    maxPreloads: 6,
    maxParallel: 3,

    keybinds: {
      previousChapter: "Comma",
      nextChapter: "Period",

      previousPage: "ArrowLeft",
      nextPage: "ArrowRight"
    }
  };

  const str = localStorage.getItem("pref");
  if (str) {
    pref = JSON.parse(str);
  } else {
    localStorage.setItem("pref", JSON.stringify(pref));
  }

  render(
    <Router>
      <Reader data={readerData} pref={pref} />
    </Router>,
    target
  );
};

if (document.readyState === "complete") {
  initReader();
} else {
  document.addEventListener("DOMContentLoaded", initReader);
}

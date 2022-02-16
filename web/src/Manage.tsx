import React from "react";
import { render } from "react-dom";
import Manage from "./Components/Manage";
import "./styles/manage.less";

const initManage = () => {
  const target = document.getElementById("manage");
  if (!target) return;
  render(<Manage data={JSON.parse(JSON.stringify(viewData))} />, target);
};

if (document.readyState === "complete") {
  initManage();
} else {
  document.addEventListener("DOMContentLoaded", initManage);
}

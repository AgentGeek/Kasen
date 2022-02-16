import feather from "feather-icons";
import "./styles/common.less";

const initCommon = () => {
  if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("/serviceWorker.js");
  }
  feather.replace({ "stroke-width": 3 });
};

if (document.readyState === "complete") {
  initCommon();
} else {
  document.addEventListener("DOMContentLoaded", initCommon);
}

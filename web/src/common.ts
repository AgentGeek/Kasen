import feather from "feather-icons";
import "./styles/common.less";

const initCommon = () => {
  if ("serviceWorker" in navigator) {
    navigator.serviceWorker.getRegistrations().then(registrations => {
      for (let i = 0; i < registrations.length; i++) {
        registrations[i].unregister();
      }
    });
  }
  feather.replace({ "stroke-width": 3 });
};

if (document.readyState === "complete") {
  initCommon();
} else {
  document.addEventListener("DOMContentLoaded", initCommon);
}

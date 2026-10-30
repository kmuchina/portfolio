/* Inlined in <head> before any CSS: applies the stored theme with no flash.
   Preference: "dark" | "system" | "light" (default "dark"). */
(function () {
  var d = document.documentElement, p = "dark";
  try {
    var s = localStorage.getItem("theme");
    if (s === "system" || s === "light") p = s;
  } catch (e) {}
  d.dataset.themePref = p;
  d.dataset.theme = p === "system"
    ? (matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark")
    : p;
  d.classList.add("js");
})();

// Language tabs interactivity for index.astro
// Only runs in browser (not during SSR)
if (typeof document !== "undefined") {
  document.querySelectorAll(".lang-tab").forEach((tab) => {
    tab.addEventListener("click", () => {
      const lang = tab.getAttribute("data-lang");
      document.querySelectorAll(".lang-tab").forEach((t) => {
        t.classList.remove("active");
        t.setAttribute("aria-selected", "false");
      });
      tab.classList.add("active");
      tab.setAttribute("aria-selected", "true");
      document.querySelectorAll(".lang-panel").forEach((p) => {
        p.classList.remove("active");
        if (p.getAttribute("data-lang") === lang) {
          p.classList.add("active");
        }
      });
    });
  });
}

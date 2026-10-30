/* main.js — deferred. Theme toggles, header scrolled state, mobile menu,
   career graph, skill evidence toggles, copy email, the contact clock, the
   footer year and scroll reveal. Progressive enhancement: the page is
   complete without it. */
(() => {
  "use strict";

  const root = document.documentElement;

  /* ---- Theme ----------------------------------------------------------- */
  const KEY = "theme";
  const PREFS = ["dark", "system", "light"];
  const lightMQ = matchMedia("(prefers-color-scheme: light)");

  const storedPref = () => {
    try {
      const v = localStorage.getItem(KEY);
      return PREFS.includes(v) ? v : "dark";
    } catch {
      return "dark";
    }
  };

  const resolve = (pref) =>
    pref === "system" ? (lightMQ.matches ? "light" : "dark") : pref;

  // Every toggle on the page (header, mobile menu, styleguide) shares one
  // state: applyTheme updates aria-checked and the roving tabindex in all.
  const toggles = document.querySelectorAll("[data-theme-toggle]");
  const radios = () =>
    document.querySelectorAll('[data-theme-toggle] [role="radio"]');

  function applyTheme(pref) {
    root.dataset.theme = resolve(pref);
    root.dataset.themePref = pref;

    for (const btn of radios()) {
      const on = btn.dataset.value === pref;
      btn.setAttribute("aria-checked", String(on));
      btn.tabIndex = on ? 0 : -1;
    }

    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.content = getComputedStyle(root).getPropertyValue("--bg").trim();

    document.dispatchEvent(new CustomEvent("themechange", { detail: { pref, theme: root.dataset.theme } }));
  }

  function setPref(pref) {
    root.classList.add("theme-fade"); // cross-fade only after a user toggle
    try {
      localStorage.setItem(KEY, pref);
    } catch {}
    applyTheme(pref);
  }

  for (const group of toggles) {
    group.addEventListener("click", (e) => {
      const btn = e.target.closest('[role="radio"]');
      if (btn) setPref(btn.dataset.value);
    });

    // Roving tabindex: arrows move focus and selection, Home/End jump.
    group.addEventListener("keydown", (e) => {
      const items = [...group.querySelectorAll('[role="radio"]')];
      const i = items.indexOf(document.activeElement);
      if (i < 0) return;
      let next;
      if (e.key === "ArrowRight" || e.key === "ArrowDown") next = (i + 1) % items.length;
      else if (e.key === "ArrowLeft" || e.key === "ArrowUp") next = (i - 1 + items.length) % items.length;
      else if (e.key === "Home") next = 0;
      else if (e.key === "End") next = items.length - 1;
      else return;
      e.preventDefault();
      items[next].focus();
      setPref(items[next].dataset.value);
    });
  }

  lightMQ.addEventListener("change", () => {
    if (root.dataset.themePref === "system") applyTheme("system");
  });

  applyTheme(storedPref());

  /* ---- Header: scrolled state ------------------------------------------
     A 1px sentinel sits at the very top of the page; once it leaves the
     viewport the header turns frosted. No scroll listener. */
  const header = document.querySelector("[data-site-header]");
  const sentinel = document.querySelector("[data-scroll-sentinel]");
  if (header && sentinel) {
    if ("IntersectionObserver" in window) {
      new IntersectionObserver(([entry]) => {
        header.classList.toggle("site-header--scrolled", !entry.isIntersecting);
      }).observe(sentinel);
    } else {
      header.classList.add("site-header--scrolled"); // solid, never transparent
    }
  }

  /* ---- Mobile menu ------------------------------------------------------
     A native <dialog> opened with showModal(): focus trap, Esc and an inert
     page come for free. Closes on a link click (then the link navigates),
     Esc, a backdrop click or a resize past the 1024px nav breakpoint; focus
     then returns to the menu button. */
  const menu = document.querySelector("[data-menu]");
  const menuBtn = document.querySelector("[data-menu-open]");
  if (menu && menuBtn && typeof menu.showModal === "function") {
    const links = [...menu.querySelectorAll(".menu__link")];

    // Mark the section currently under the header (Phase 6 scrollspy will
    // do this live for the desktop nav too).
    const markCurrent = () => {
      const line = header ? header.getBoundingClientRect().bottom + innerHeight / 3 : innerHeight / 3;
      let current = null;
      for (const link of links) {
        const target = document.getElementById(link.hash.slice(1));
        if (target && target.getBoundingClientRect().top <= line) current = link;
      }
      for (const link of links) {
        if (link === current) link.setAttribute("aria-current", "true");
        else link.removeAttribute("aria-current");
      }
    };

    menuBtn.addEventListener("click", () => {
      markCurrent();
      menu.showModal();
      menuBtn.setAttribute("aria-expanded", "true");
    });
    menu.addEventListener("close", () => {
      menuBtn.setAttribute("aria-expanded", "false");
      menuBtn.focus();
    });
    menu.querySelector("[data-menu-close]")?.addEventListener("click", () => menu.close());
    menu.addEventListener("click", (e) => {
      if (e.target.closest(".menu__link")) {
        menu.close(); // the link's default navigation then runs
        return;
      }
      // A click on ::backdrop targets the dialog itself, outside its box.
      if (e.target !== menu) return;
      const r = menu.getBoundingClientRect();
      if (e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom) menu.close();
    });
    matchMedia("(min-width: 1024px)").addEventListener("change", (e) => {
      if (e.matches && menu.open) menu.close();
    });
  }

  /* ---- Career graph (experience) ----------------------------------------
     Legend filter, live duration of the current role, the fallback for
     browsers without scroll-driven animations, and pausing the pulse. */
  const career = document.querySelector(".career");
  if (career) {
    const reduceMotion = matchMedia("(prefers-reduced-motion: reduce)").matches;

    // Legend: highlight one kind; pressing it again (or "All") resets.
    // Dims, never hides (09-motion.css / 06-sections.css).
    const legend = document.querySelectorAll(".legend__item");
    for (const btn of legend) {
      btn.addEventListener("click", () => {
        const again = btn.getAttribute("aria-pressed") === "true";
        const kind = again ? "all" : btn.dataset.filter;
        career.dataset.filter = kind;
        career.dataset.lane = again ? "all" : btn.dataset.lane;
        for (const b of legend) b.setAttribute("aria-pressed", String(b.dataset.filter === kind));
        for (const row of career.children) {
          row.classList.toggle("is-quiet", kind !== "all" && !row.classList.contains(`career__row--${kind}`));
        }
      });
    }

    // The current role's duration is stale on a static site: recompute it
    // (inclusive months, as Go does).
    for (const el of career.querySelectorAll("[data-start]")) {
      const [y, m] = el.dataset.start.split("-").map(Number);
      const now = new Date();
      const months = (now.getFullYear() - y) * 12 + (now.getMonth() + 1 - m) + 1;
      const yrs = Math.floor(months / 12), mos = months % 12;
      const part = (n, unit) => n && `${n} ${unit}${n > 1 ? "s" : ""}`;
      if (months > 0) el.textContent = [part(yrs, "yr"), part(mos, "mo")].filter(Boolean).join(" ");
    }

    // Fallback: no scroll-driven animations → light each row once.
    if (!reduceMotion && !CSS.supports("animation-timeline: view()") && "IntersectionObserver" in window) {
      const io = new IntersectionObserver((entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            e.target.classList.add("is-lit");
            io.unobserve(e.target);
          }
        }
      }, { rootMargin: "0px 0px -35% 0px" });
      for (const row of career.children) io.observe(row);
    }

    // The one looping animation: pause it while off-screen.
    const pulse = career.querySelector(".node--current");
    if (pulse && !reduceMotion && "IntersectionObserver" in window) {
      new IntersectionObserver(([e]) => pulse.classList.toggle("is-paused", !e.isIntersecting)).observe(pulse);
    }
  }

  /* ---- Skills: "used in" panels ------------------------------------------
     Rendered open (no-JS readers see the evidence inline); here they close
     and each skill button toggles its panel, one open per row. */
  for (const row of document.querySelectorAll("[data-skill-row]")) {
    const btns = [...row.querySelectorAll("button[aria-controls]")];
    const show = (btn, on) => {
      btn.setAttribute("aria-expanded", String(on));
      document.getElementById(btn.getAttribute("aria-controls")).hidden = !on;
    };
    for (const btn of btns) {
      show(btn, false);
      btn.addEventListener("click", () => {
        const open = btn.getAttribute("aria-expanded") === "true";
        for (const b of btns) show(b, false);
        if (!open) show(btn, true);
      });
    }
  }

  /* ---- Contact: copy email ----------------------------------------------
     Clipboard first; if that's blocked, select the address and say how to
     copy it. Confirmed visibly and to screen readers (role="status"). */
  const copy = document.querySelector("[data-copy]");
  if (copy) {
    const label = copy.querySelector(".copy__label");
    const status = document.querySelector("[data-copy-status]");
    const d = copy.dataset;
    let timer;
    copy.addEventListener("click", async () => {
      let ok = false;
      try {
        await navigator.clipboard.writeText(d.copy);
        ok = true;
      } catch {
        const range = document.createRange();
        range.selectNodeContents(document.querySelector("[data-copy-source]"));
        getSelection().removeAllRanges();
        getSelection().addRange(range);
      }
      copy.classList.toggle("is-done", ok);
      label.textContent = ok ? d.done : d.fallback;
      if (status) status.textContent = ok ? d.statusDone : d.statusFallback;
      clearTimeout(timer);
      timer = setTimeout(() => {
        copy.classList.remove("is-done");
        label.textContent = d.label;
        if (status) status.textContent = "";
      }, 2200);
    });
  }

  /* ---- Contact: live local time (the build's time is the no-JS value) --- */
  const clock = document.querySelector("time[data-tz]");
  if (clock && window.Intl) {
    const fmt = new Intl.DateTimeFormat("en-GB", { timeZone: clock.dataset.tz, hour: "2-digit", minute: "2-digit" });
    const tick = () => {
      const now = new Date();
      clock.textContent = fmt.format(now);
      clock.dateTime = now.toISOString();
    };
    tick();
    setInterval(tick, 30000);
  }

  for (const el of document.querySelectorAll("[data-year]")) el.textContent = new Date().getFullYear();

  /* ---- Reveal ---------------------------------------------------------- */
  for (const group of document.querySelectorAll("[data-reveal-group]")) {
    [...group.children].forEach((child, i) => child.style.setProperty("--i", i));
  }

  const targets = document.querySelectorAll("[data-reveal], [data-reveal-group]");
  const reduced = matchMedia("(prefers-reduced-motion: reduce)").matches;

  if (reduced || !("IntersectionObserver" in window)) {
    for (const el of targets) el.classList.add("is-visible");
  } else {
    const io = new IntersectionObserver(
      (entries, observer) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue;
          entry.target.classList.add("is-visible");
          observer.unobserve(entry.target);
        }
      },
      { threshold: 0.15, rootMargin: "0px 0px -10% 0px" }
    );
    for (const el of targets) io.observe(el);
  }
})();

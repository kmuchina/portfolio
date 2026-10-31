/* main.js — deferred progressive enhancement: the page is complete without
   it. One function per feature, run by the loop at the end: each returns
   early when its elements aren't on the page, and one that throws never
   stops the rest. Features talk through DOM events on document:
   "themechange" (theme), "sectionchange" (scrollspy → nav indicator) and
   "arrive" (deep links → career, which resets the legend filter).
   Reduced motion is handled in CSS (the pulse, the nav indicator, the
   highlight all follow the media query live); JS reads
   reduceMotion.matches at the moment it scrolls. */
(() => {
  "use strict";

  const root = document.documentElement;
  const reduceMotion = matchMedia("(prefers-reduced-motion: reduce)");
  const hasIO = "IntersectionObserver" in window;
  const emit = (name, detail) => document.dispatchEvent(new CustomEvent(name, { detail }));

  // The page's one polite live region (#sr-status, layout.html). Emptied
  // first, so the same message twice is announced twice.
  const status = document.querySelector("[data-status]");
  let said;
  const announce = (msg) => {
    if (!status || !msg) return;
    status.textContent = "";
    clearTimeout(said);
    said = setTimeout(() => (status.textContent = msg), 50);
  };

  /* ---- Theme: dark | system | light (default dark) ---------------------
     Every toggle (header, menu, styleguide) shares one state; arrows move
     focus and selection (roving tabindex), Home/End jump. theme-init.js
     already set data-theme before first paint. */
  function theme({ announce }) {
    const KEY = "theme";
    const lightMQ = matchMedia("(prefers-color-scheme: light)");
    const meta = document.querySelector('meta[name="theme-color"]');
    const radios = [...document.querySelectorAll('[data-theme-toggle] [role="radio"]')];
    let pref = "dark";
    try {
      const v = localStorage.getItem(KEY);
      if (v === "system" || v === "light") pref = v;
    } catch {}

    const apply = () => {
      root.dataset.theme = pref === "system" ? (lightMQ.matches ? "light" : "dark") : pref;
      root.dataset.themePref = pref;
      for (const btn of radios) {
        const on = btn.dataset.value === pref;
        btn.setAttribute("aria-checked", String(on));
        btn.tabIndex = on ? 0 : -1;
      }
      if (meta) meta.content = getComputedStyle(root).getPropertyValue("--bg").trim();
      emit("themechange", { pref, theme: root.dataset.theme });
    };

    // "Dark theme" / "Light theme" / "System theme, currently dark".
    const say = () => {
      const d = status?.dataset;
      if (!d) return;
      const name = (d[`${root.dataset.theme}Name`] || root.dataset.theme).toLowerCase();
      announce(pref === "system" ? d.system.replace("{theme}", name) : d[pref]);
    };

    const set = (next) => {
      root.classList.add("theme-fade"); // cross-fade only after a user toggle
      pref = next;
      try {
        localStorage.setItem(KEY, pref);
      } catch {}
      apply();
      say();
    };

    for (const group of document.querySelectorAll("[data-theme-toggle]")) {
      group.addEventListener("click", (e) => {
        const btn = e.target.closest('[role="radio"]');
        if (btn) set(btn.dataset.value);
      });
      group.addEventListener("keydown", (e) => {
        const items = [...group.querySelectorAll('[role="radio"]')];
        const i = items.indexOf(document.activeElement);
        const n = items.length;
        const next = { ArrowRight: i + 1, ArrowDown: i + 1, ArrowLeft: i - 1 + n, ArrowUp: i - 1 + n, Home: 0, End: n - 1 }[e.key];
        if (i < 0 || next === undefined) return;
        e.preventDefault();
        items[next % n].focus();
        set(items[next % n].dataset.value);
      });
    }

    lightMQ.addEventListener("change", () => pref === "system" && apply());
    apply();
  }

  /* ---- Header: frosted once the 1px top sentinel scrolls away ---------- */
  function header() {
    const el = document.querySelector("[data-site-header]");
    const sentinel = document.querySelector("[data-scroll-sentinel]");
    if (!el || !sentinel) return;
    if (!hasIO) return el.classList.add("site-header--scrolled"); // solid, never transparent
    new IntersectionObserver(([e]) => el.classList.toggle("site-header--scrolled", !e.isIntersecting)).observe(sentinel);
  }

  /* ---- Mobile menu: a native <dialog> (focus trap, Esc, inert page) ----
     Closes on a link click (before the link navigates, so the page is
     scrollable again when the jump happens), a backdrop click or a resize
     past the 1024px nav breakpoint; focus returns to the menu button.
     The current link is kept live by scrollspy. */
  function menu() {
    const dialog = document.querySelector("[data-menu]");
    const btn = document.querySelector("[data-menu-open]");
    if (!dialog || !btn || typeof dialog.showModal !== "function") return;

    btn.addEventListener("click", () => {
      dialog.showModal();
      btn.setAttribute("aria-expanded", "true");
    });
    dialog.addEventListener("close", () => {
      btn.setAttribute("aria-expanded", "false");
      btn.focus({ preventScroll: true });
    });
    dialog.querySelector("[data-menu-close]")?.addEventListener("click", () => dialog.close());
    dialog.addEventListener("click", (e) => {
      if (e.target.closest(".menu__link")) return dialog.close();
      // A click on ::backdrop targets the dialog itself, outside its box.
      if (e.target !== dialog) return;
      const r = dialog.getBoundingClientRect();
      if (e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom) dialog.close();
    });
    matchMedia("(min-width: 1024px)").addEventListener("change", (e) => e.matches && dialog.open && dialog.close());
  }

  /* ---- Scrollspy: aria-current on the nav and menu links ----------------
     One IntersectionObserver, no scroll listener. A section is current
     while it crosses a thin band just above the middle of the viewport;
     the one that most recently entered the band wins. The hero has no
     link, so over it nothing is current. At the very bottom of the page
     Contact (the last section) is current even if it never reached the
     band (checked in the observer callback and on scrollend). Never
     touches the URL. */
  function scrollspy() {
    const links = [...document.querySelectorAll(".nav__link, .menu__link")];
    const sections = [...document.querySelectorAll("main > section[id]")];
    if (!hasIO || !links.length || !sections.length) return;
    const last = sections.at(-1);
    const inBand = new Set();
    let latest = null;
    let current;

    const update = () => {
      const atBottom = scrollY + innerHeight >= root.scrollHeight - 4;
      const id = (atBottom ? last : latest)?.id ?? null;
      if (id === current) return;
      current = id;
      let match = null;
      for (const a of links) {
        const on = id !== null && a.hash === `#${id}`;
        if (on) {
          a.setAttribute("aria-current", "true");
          match ??= a;
        } else a.removeAttribute("aria-current");
      }
      emit("sectionchange", { id: match ? id : null });
    };

    const spy = new IntersectionObserver((entries) => {
      for (const e of entries) {
        if (e.isIntersecting) {
          inBand.add(e.target);
          latest = e.target;
        } else inBand.delete(e.target);
      }
      if (latest && !inBand.has(latest) && inBand.size) latest = [...inBand].at(-1);
      update();
    }, { rootMargin: "-40% 0px -55% 0px" });
    for (const s of sections) spy.observe(s);
    addEventListener("scrollend", update);
  }

  /* ---- Nav indicator: one bar that slides to the current link ----------
     ≥ 1024px only (below, the nav is hidden and the bar stays off). Placed
     from the link's offsetLeft/offsetWidth minus its padding; re-placed
     when the nav resizes and once the fonts have loaded. */
  function navIndicator() {
    const bar = document.querySelector("[data-nav-indicator]");
    const nav = bar?.parentElement;
    if (!bar) return;
    let link = null;

    const place = () => {
      if (!link || !link.offsetWidth) return bar.classList.remove("is-on");
      const pad = parseFloat(getComputedStyle(link).paddingInlineStart) || 0;
      const appearing = !bar.classList.contains("is-on");
      if (appearing) bar.classList.add("is-placing"); // appear in place, don't slide in
      bar.style.setProperty("--nav-x", `${link.offsetLeft + pad}px`);
      bar.style.setProperty("--nav-w", String((link.offsetWidth - 2 * pad) / 100));
      if (appearing) {
        bar.getBoundingClientRect(); // commit the new position before transitions return
        bar.classList.remove("is-placing");
      }
      bar.classList.add("is-on");
    };

    document.addEventListener("sectionchange", (e) => {
      const id = e.detail.id;
      link = id ? [...nav.querySelectorAll(".nav__link")].find((a) => a.hash === `#${id}`) ?? null : null;
      place();
    });
    if ("ResizeObserver" in window) new ResizeObserver(place).observe(nav);
    document.fonts?.ready.then(place);
  }

  /* ---- Deep links: landing, highlight, focus ------------------------------
     Works for in-page clicks (nav, skill "used in" links, the skip link)
     and for hashes arriving from outside (a shared #job-… link, the URL
     bar, back/forward). Once the jump settles, focus moves to the
     target's heading (tabindex -1, preventScroll) and a job or project
     gets a one-time highlight (.is-arriving on the :target). The URL is
     the browser's business: a click updates the hash; nothing else does,
     except Back to top, which clears it without a history entry. An
     unknown hash does nothing. */
  function deepLinks({ reduceMotion }) {
    const HOLD = { job: 1600, project: 1200 };
    const byHash = (hash) => {
      try {
        return hash.length > 1 ? document.getElementById(decodeURIComponent(hash.slice(1))) : null;
      } catch {
        return null;
      }
    };
    const kindOf = (el) => (el.matches("article.job") ? "job" : el.matches("article.project") ? "project" : null);

    // The heading a reader continues from: the section's #<id>-title (the
    // hero's <h1>), an article's first heading; <main> itself.
    // quiet: the arrival came with the page load, when nobody knows yet
    // whether the reader uses a keyboard (browsers then treat script focus
    // as :focus-visible), so the heading's ring waits until it's left.
    const focusHeading = (el, quiet = false) => {
      const h = el.matches("main") ? el : el.querySelector(`#${CSS.escape(el.id)}-title, h1, h2, h3`) ?? el;
      if (!h.hasAttribute("tabindex") && !h.matches("a[href], button")) h.tabIndex = -1;
      if (quiet) {
        h.dataset.quietFocus = "";
        h.addEventListener("blur", () => delete h.dataset.quietFocus, { once: true });
      }
      h.focus({ preventScroll: true });
    };

    let lit = null;
    let unlight;
    const highlight = (el) => {
      const kind = kindOf(el);
      if (!kind) return;
      lit?.classList.remove("is-arriving");
      void el.offsetWidth; // restart the ring on a repeat visit
      el.classList.add("is-arriving");
      lit = el;
      clearTimeout(unlight);
      unlight = setTimeout(() => el.classList.remove("is-arriving"), HOLD[kind]);
    };

    // Run fn once the jump has settled: on scrollend, at once if nothing
    // moved, and at the latest after 1.5s (no scrollend support).
    const settle = (fn) => {
      let done = false;
      const go = () => {
        if (done) return;
        done = true;
        removeEventListener("scrollend", go);
        fn();
      };
      const y = scrollY;
      addEventListener("scrollend", go);
      setTimeout(() => scrollY === y && go(), 120);
      setTimeout(go, 1500);
    };

    let clicked = false;
    const arrive = (el) => {
      emit("arrive", { target: el }); // career: reset the legend if the job is dimmed
      settle(() => {
        clicked = false;
        focusHeading(el);
        highlight(el);
      });
    };

    document.addEventListener("click", (e) => {
      if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
      const a = e.target.closest("a[href*='#']");
      if (!a || a.origin !== location.origin || a.pathname !== location.pathname) return;
      const el = byHash(a.hash);
      if (!el) return;
      // Back to top: scroll up, drop the hash without a history entry, and
      // continue from the hero's <h1> (or <main> on other pages).
      if (a.matches(".to-top")) {
        e.preventDefault();
        history.replaceState(history.state, "", location.pathname + location.search);
        scrollTo({ top: 0, behavior: reduceMotion.matches ? "instant" : "smooth" });
        return settle(() => focusHeading(el));
      }
      clicked = a.hash !== location.hash; // a hashchange will follow: don't handle it twice
      arrive(el);
    });

    addEventListener("hashchange", () => {
      if (clicked) return void (clicked = false);
      const el = byHash(location.hash);
      if (el) arrive(el);
    });

    // Loaded with a hash: late fonts can shift the layout after the
    // browser's own jump, so land again once they're in, unless the reader
    // has already scrolled.
    const first = byHash(location.hash);
    if (first) {
      let moved = false;
      const mark = () => (moved = true);
      for (const t of ["wheel", "touchstart", "keydown", "pointerdown"]) addEventListener(t, mark, { once: true, passive: true });
      (document.fonts?.ready ?? Promise.resolve()).then(() => {
        if (moved) return;
        emit("arrive", { target: first });
        first.scrollIntoView({ block: "start", behavior: "instant" });
        focusHeading(first, true);
        highlight(first);
      });
    }
  }

  /* ---- Career graph -------------------------------------------------------
     Legend filter (announced), the current role's live duration, the
     fallback for browsers without scroll-driven animations, and pausing
     the pulse off-screen. Reduced motion is pure CSS, so live changes to
     it need nothing here. */
  function career({ announce }) {
    const list = document.querySelector(".career");
    if (!list) return;

    // Highlight one kind; pressing it again (or "All") resets. Dims, never hides.
    const legend = [...document.querySelectorAll(".legend__item")];
    const all = legend.find((b) => b.dataset.filter === "all");
    const filter = (btn) => {
      const kind = btn.dataset.filter;
      list.dataset.filter = kind;
      list.dataset.lane = btn.dataset.lane;
      for (const b of legend) b.setAttribute("aria-pressed", String(b === btn));
      for (const row of list.children) {
        row.classList.toggle("is-quiet", kind !== "all" && !row.classList.contains(`career__row--${kind}`));
      }
      announce(btn.dataset.announce);
    };
    for (const btn of legend) {
      btn.addEventListener("click", () => filter(btn.getAttribute("aria-pressed") === "true" && all ? all : btn));
    }
    // A link to a dimmed job resets the filter first.
    document.addEventListener("arrive", (e) => {
      if (all && e.detail.target.closest(".career__row.is-quiet")) filter(all);
    });

    // The current role's duration is stale on a static site: recompute it
    // (inclusive months, as Go does).
    for (const el of list.querySelectorAll("[data-start]")) {
      const [y, m] = el.dataset.start.split("-").map(Number);
      const now = new Date();
      const months = (now.getFullYear() - y) * 12 + (now.getMonth() + 1 - m) + 1;
      const part = (n, unit) => n && `${n} ${unit}${n > 1 ? "s" : ""}`;
      if (months > 0) el.textContent = [part(Math.floor(months / 12), "yr"), part(months % 12, "mo")].filter(Boolean).join(" ");
    }

    if (!hasIO) return;

    // No scroll-driven animations: light each row once as it enters.
    if (!CSS.supports("animation-timeline: view()")) {
      const lightUp = new IntersectionObserver((entries) => {
        for (const e of entries) {
          if (!e.isIntersecting) continue;
          e.target.classList.add("is-lit");
          lightUp.unobserve(e.target);
        }
      }, { rootMargin: "0px 0px -35% 0px" });
      for (const row of list.children) lightUp.observe(row);
    }

    // The one looping animation: paused while off-screen.
    const pulse = list.querySelector(".node--current");
    if (pulse) new IntersectionObserver(([e]) => pulse.classList.toggle("is-paused", !e.isIntersecting)).observe(pulse);
  }

  /* ---- Skills: "used in" panels -------------------------------------------
     Rendered open (no-JS readers see the evidence inline); here they close
     and each skill button toggles its panel, one open per row. Esc closes
     the open panel and returns focus to its button. */
  function skills() {
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
      row.addEventListener("keydown", (e) => {
        const open = btns.find((b) => b.getAttribute("aria-expanded") === "true");
        if (e.key !== "Escape" || !open) return;
        e.preventDefault();
        show(open, false);
        open.focus();
      });
    }
  }

  /* ---- Contact: copy email --------------------------------------------------
     Clipboard where available (https); otherwise, or if it's refused,
     select the address and say how to copy it. Shown on the button for
     2.2s and announced through the live region. */
  function copyEmail({ announce }) {
    const btn = document.querySelector("[data-copy]");
    if (!btn) return;
    const label = btn.querySelector(".copy__label");
    const source = document.querySelector("[data-copy-source]");
    const d = btn.dataset;
    let timer;
    btn.addEventListener("click", async () => {
      let ok = false;
      try {
        if (navigator.clipboard?.writeText) {
          await navigator.clipboard.writeText(d.copy);
          ok = true;
        }
      } catch {}
      if (!ok && source) {
        const range = document.createRange();
        range.selectNodeContents(source);
        getSelection()?.removeAllRanges();
        getSelection()?.addRange(range);
      }
      btn.classList.toggle("is-done", ok);
      label.textContent = ok ? d.done : d.fallback;
      announce(ok ? d.statusDone : d.statusFallback);
      clearTimeout(timer);
      timer = setTimeout(() => {
        btn.classList.remove("is-done");
        label.textContent = d.label;
      }, 2200);
    });
  }

  /* ---- Contact: live local time (the build's time is the no-JS value) ----
     The site's only interval; it stops while the tab is hidden. */
  function clock() {
    const el = document.querySelector("time[data-tz]");
    if (!el || !window.Intl) return;
    const fmt = new Intl.DateTimeFormat("en-GB", { timeZone: el.dataset.tz, hour: "2-digit", minute: "2-digit" });
    let timer;
    const tick = () => {
      const now = new Date();
      el.textContent = fmt.format(now);
      el.dateTime = now.toISOString();
    };
    const run = () => {
      clearInterval(timer);
      if (document.hidden) return;
      tick();
      timer = setInterval(tick, 30000);
    };
    document.addEventListener("visibilitychange", run);
    run();
  }

  function footerYear() {
    for (const el of document.querySelectorAll("[data-year]")) el.textContent = new Date().getFullYear();
  }

  const features = {
    theme, header, menu, scrollspy, navIndicator, deepLinks,
    career, skills, copyEmail, clock, footerYear,
  };
  for (const [name, init] of Object.entries(features)) {
    try {
      init({ reduceMotion, announce });
    } catch (err) {
      console.error(`[${name}]`, err);
    }
  }
})();

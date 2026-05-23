// Debounced live-search dropdown for the sidebar search box.
(function () {
  const box = document.getElementById("search");
  if (!box) return;
  let timer = null;
  let dropdown = null;
  box.addEventListener("input", () => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => query(box.value), 200);
  });
  async function query(q) {
    if (!q) { hide(); return; }
    const res = await fetch("/__search?q=" + encodeURIComponent(q));
    const data = await res.json();
    show(data.results || []);
  }
  function show(results) {
    if (!dropdown) {
      dropdown = document.createElement("div");
      dropdown.style.cssText = "position:absolute;background:var(--bg);border:1px solid var(--border);padding:0.5rem;max-width:300px;z-index:100";
      box.parentNode.appendChild(dropdown);
    }
    dropdown.innerHTML = results.slice(0, 10).map(r =>
      `<a href="${r.url}" style="display:block;padding:0.3rem 0"><strong>${escape(r.title)}</strong><br><small>${r.snippet}</small></a>`
    ).join("");
  }
  function hide() { if (dropdown) dropdown.innerHTML = ""; }
  function escape(s) { return String(s).replace(/[<>&]/g, c => ({"<":"&lt;",">":"&gt;","&":"&amp;"}[c])); }
})();

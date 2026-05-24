// Text-to-speech for the article body. Uses the browser-native
// window.speechSynthesis API — no network, no API key. Voice quality
// varies by OS/browser; defaults to the platform's first English voice.
(function () {
  if (!("speechSynthesis" in window)) {
    var bar = document.querySelector(".toolbar");
    if (bar) bar.style.display = "none";
    return;
  }
  var synth = window.speechSynthesis;
  var readBtn = document.getElementById("tts-read");
  var stopBtn = document.getElementById("tts-stop");
  var rateSel = document.getElementById("tts-rate");
  if (!readBtn || !stopBtn || !rateSel) return;

  // Collect spoken text from the article body, skipping pre/code blocks
  // and the toolbar itself. Returns a single string with paragraph
  // boundaries collapsed to ". ".
  function collectText() {
    var main = document.querySelector("main");
    if (!main) return "";
    var clone = main.cloneNode(true);
    clone.querySelectorAll(".toolbar, pre, code, .folder-contents").forEach(function (n) {
      n.remove();
    });
    var raw = clone.innerText || clone.textContent || "";
    return raw.replace(/\s+/g, " ").trim();
  }

  function setReading(reading) {
    readBtn.textContent = reading ? "⏸ Pause" : "▶ Read";
    readBtn.dataset.state = reading ? "playing" : "idle";
  }

  function start() {
    var text = collectText();
    if (!text) return;
    synth.cancel();
    var u = new SpeechSynthesisUtterance(text);
    u.rate = parseFloat(rateSel.value) || 1.0;
    u.onend = function () { setReading(false); };
    u.onerror = function () { setReading(false); };
    synth.speak(u);
    setReading(true);
  }

  readBtn.addEventListener("click", function () {
    if (readBtn.dataset.state === "playing") {
      synth.pause();
      setReading(false);
      readBtn.textContent = "▶ Resume";
      readBtn.dataset.state = "paused";
    } else if (readBtn.dataset.state === "paused") {
      synth.resume();
      setReading(true);
    } else {
      start();
    }
  });

  stopBtn.addEventListener("click", function () {
    synth.cancel();
    setReading(false);
  });

  rateSel.addEventListener("change", function () {
    if (readBtn.dataset.state === "playing" || readBtn.dataset.state === "paused") {
      // SpeechSynthesisUtterance rate is immutable mid-speech in most
      // browsers — restart with the new rate.
      synth.cancel();
      start();
    }
  });

  // Stop TTS on page navigation (covers SSE-driven live reload too).
  window.addEventListener("beforeunload", function () { synth.cancel(); });
})();

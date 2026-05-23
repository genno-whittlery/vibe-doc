(function () {
  const src = new EventSource("/__events");
  src.onmessage = (e) => {
    if (e.data === "reload") {
      location.reload();
    }
  };
  src.onerror = () => {
    setTimeout(() => location.reload(), 1000);
  };
})();

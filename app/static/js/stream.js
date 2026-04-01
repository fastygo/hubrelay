(function () {
  function ready(fn) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", fn);
      return;
    }
    fn();
  }

  ready(function () {
    var form = document.getElementById("ask-form");
    var output = document.getElementById("stream-output");
    var result = document.getElementById("stream-result");
    var status = document.getElementById("stream-status");
    var prompt = document.getElementById("prompt-input");
    var model = document.getElementById("model-input");

    if (!form || !output || !result || !status || !prompt) {
      return;
    }

    var activeSource = null;

    function closeSource() {
      if (activeSource) {
        activeSource.close();
        activeSource = null;
      }
    }

    function setStatus(value) {
      status.textContent = value;
    }

    form.addEventListener("submit", function (event) {
      var submitter = event.submitter;
      if (submitter && submitter.value === "sync") {
        closeSource();
        setStatus("sync");
        return;
      }

      event.preventDefault();

      var promptValue = prompt.value.trim();
      if (!promptValue) {
        setStatus("prompt required");
        return;
      }

      closeSource();
      output.textContent = "";
      result.classList.add("hidden");
      result.textContent = "";
      setStatus("connecting");

      var params = new URLSearchParams();
      params.set("prompt", promptValue);
      if (model && model.value.trim()) {
        params.set("model", model.value.trim());
      }

      activeSource = new EventSource("/ask/stream?" + params.toString());

      activeSource.addEventListener("chunk", function (message) {
        setStatus("streaming");
        try {
          var payload = JSON.parse(message.data);
          output.textContent += payload.delta || "";
        } catch (error) {
          output.textContent += message.data;
        }
      });

      activeSource.addEventListener("done", function (message) {
        closeSource();
        setStatus("done");
        result.classList.remove("hidden");
        result.textContent = message.data;
      });

      activeSource.addEventListener("error", function (message) {
        closeSource();
        setStatus("error");
        result.classList.remove("hidden");
        result.textContent = message.data || "stream failed";
      });
    });
  });
})();

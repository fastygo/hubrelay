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
    var panel = document.getElementById("stream-panel");
    var output = document.getElementById("stream-output");
    var result = document.getElementById("stream-result");
    var status = document.getElementById("stream-status");
    var prompt = document.getElementById("prompt-input");
    var model = document.getElementById("model-input");
    var context = document.getElementById("course-context-input");

    if (!form || !panel || !output || !result || !status || !prompt) {
      return;
    }

    var activeSource = null;
    var copy = {
      idle: panel.dataset.statusIdle || "idle",
      sync: panel.dataset.statusSync || "sync",
      promptRequired: panel.dataset.statusPromptRequired || "prompt required",
      connecting: panel.dataset.statusConnecting || "connecting",
      streaming: panel.dataset.statusStreaming || "streaming",
      done: panel.dataset.statusDone || "done",
      error: panel.dataset.statusError || "error",
      defaultError: panel.dataset.defaultError || "stream failed"
    };

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
        setStatus(copy.sync);
        return;
      }

      event.preventDefault();

      var promptValue = prompt.value.trim();
      if (!promptValue) {
        setStatus(copy.promptRequired);
        return;
      }

      closeSource();
      output.textContent = "";
      result.classList.add("hidden");
      result.textContent = "";
      setStatus(copy.connecting);

      var params = new URLSearchParams();
      params.set("prompt", promptValue);
      if (model && model.value.trim()) {
        params.set("model", model.value.trim());
      }
      if (context && context.value.trim()) {
        params.set("context", context.value.trim());
      }

      var streamEndpoint = form.dataset.streamEndpoint || "/ask/stream";
      var separator = streamEndpoint.indexOf("?") === -1 ? "?" : "&";
      activeSource = new EventSource(streamEndpoint + separator + params.toString());

      activeSource.addEventListener("chunk", function (message) {
        setStatus(copy.streaming);
        try {
          var payload = JSON.parse(message.data);
          output.textContent += payload.delta || "";
        } catch (error) {
          output.textContent += message.data;
        }
      });

      activeSource.addEventListener("done", function (message) {
        closeSource();
        setStatus(copy.done);
        result.classList.remove("hidden");
        result.textContent = message.data;
      });

      activeSource.addEventListener("error", function (message) {
        closeSource();
        setStatus(copy.error);
        result.classList.remove("hidden");
        result.textContent = message.data || copy.defaultError;
      });
    });
  });
})();

package registry

import (
	"fmt"
	"strings"

	"github.com/functionfly/functionfly/internal/storage/registry"
)

// EmbedOptions holds parsed query-string options for embed script generation.
type EmbedOptions struct {
	Namespace string // Global variable name (default: "ff")
	Autoload  bool   // Auto-initialize on DOMContentLoaded
	UI        bool   // Inject default UI widget
	Theme     string // "light", "dark", "auto"
}

// generateEmbedScript produces a self-contained IIFE JavaScript embed script
// for the given function.  The script exposes a global object (default: `ff`)
// with run(), form(), on(), and widget() methods.
func generateEmbedScript(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion, version string, opts EmbedOptions) string {
	baseURL := getBaseURL()

	// Resolve the version string shown in the script
	resolvedVersion := version
	if resolvedVersion == "" {
		if fnVersion != nil {
			resolvedVersion = fnVersion.Version
		} else {
			resolvedVersion = "latest"
		}
	}

	// Determine cache-control value: immutable for pinned versions
	cacheControl := "public, max-age=300"
	if version != "" {
		cacheControl = "public, max-age=31536000, immutable"
	}
	_ = cacheControl // used in the HTTP handler, not in the script itself

	autoloadStr := "true"
	if !opts.Autoload {
		autoloadStr = "false"
	}
	uiStr := "false"
	if opts.UI {
		uiStr = "true"
	}

	description := ""
	if fn.Description.Valid {
		description = fn.Description.String
	}

	return fmt.Sprintf(`/**
 * FunctionFly Embed — %s/%s@%s
 * %s
 *
 * Usage:
 *   <script src="https://functionfly.com/embed/%s/%s.js"></script>
 *   <script>
 *     const result = await %s.run({ key: "value" });
 *   </script>
 */
(function (global, config) {
  "use strict";

  var API_BASE  = %q;
  var AUTHOR    = %q;
  var NAME      = %q;
  var VERSION   = %q;
  var NAMESPACE = %q;

  // ── Event system ──────────────────────────────────────────────────────────
  var _handlers = {};

  function on(event, handler) {
    if (!_handlers[event]) { _handlers[event] = []; }
    _handlers[event].push(handler);
  }

  function emit(event, data) {
    var list = _handlers[event] || [];
    for (var i = 0; i < list.length; i++) { list[i](data); }
  }

  // ── Core execution ────────────────────────────────────────────────────────
  function run(input, options) {
    options = options || {};

    var url = API_BASE + "/" + AUTHOR + "/" + NAME;
    if (options.version && options.version !== "latest") {
      url += "@" + options.version;
    } else if (VERSION && VERSION !== "latest") {
      url += "@" + VERSION;
    }

    var headers = {
      "Content-Type": "application/json",
      "User-Agent":   "FunctionFly-Embed/1.0",
    };

    // Support data-api-key on the <script> tag
    var scriptEl = document.currentScript ||
      (function () {
        var scripts = document.getElementsByTagName("script");
        return scripts[scripts.length - 1];
      })();
    var apiKey = scriptEl && scriptEl.getAttribute("data-api-key");
    if (apiKey) { headers["Authorization"] = "Bearer " + apiKey; }
    if (options.apiKey) { headers["Authorization"] = "Bearer " + options.apiKey; }

    emit("execute:start", input);

    var promise = fetch(url, {
      method:  "POST",
      headers: headers,
      body:    JSON.stringify(input),
    })
      .then(function (resp) {
        if (!resp.ok) {
          return resp.json().catch(function () { return {}; }).then(function (errData) {
            var msg = (errData.error && errData.error.message) ||
              ("HTTP " + resp.status + ": " + resp.statusText);
            var err = new Error(msg);
            err.code       = (errData.error && errData.error.code) || "EXECUTION_FAILED";
            err.statusCode = resp.status;
            throw err;
          });
        }
        return resp.json();
      })
      .then(function (result) {
        if (!result.ok) {
          var err = new Error(
            (result.error && result.error.message) || "Function execution failed"
          );
          err.code = (result.error && result.error.code) || "EXECUTION_FAILED";
          throw err;
        }
        emit("execute:success", result);
        if (options.onSuccess) { options.onSuccess(result.data); }
        return result;
      })
      .catch(function (err) {
        emit("execute:error", err);
        if (options.onError) { options.onError(err); }
        throw err;
      });

    return promise;
  }

  // ── Form binding ──────────────────────────────────────────────────────────
  function form(formEl, options) {
    options = options || {};
    formEl.addEventListener("submit", function (e) {
      e.preventDefault();
      var data = {};
      var elements = formEl.elements;
      for (var i = 0; i < elements.length; i++) {
        var el = elements[i];
        if (el.name) { data[el.name] = el.value; }
      }
      run(data, options);
    });
  }

  // ── UI widget ─────────────────────────────────────────────────────────────
  function widget(container, options) {
    options = options || {};
    var title       = options.title       || (AUTHOR + "/" + NAME);
    var placeholder = options.placeholder || "Enter input (JSON)...";
    var buttonText  = options.buttonText  || "Run";
    var theme       = options.theme       || config.theme || "auto";

    var html = [
      '<div class="ff-widget" data-theme="' + theme + '" style="font-family:sans-serif;border:1px solid #ddd;border-radius:8px;padding:16px;max-width:480px">',
      '  <h3 style="margin:0 0 12px;font-size:14px;font-weight:600">' + title + '</h3>',
      '  <textarea class="ff-input" placeholder="' + placeholder + '" rows="4"',
      '    style="width:100%;box-sizing:border-box;padding:8px;border:1px solid #ccc;border-radius:4px;font-size:13px;resize:vertical"></textarea>',
      '  <button class="ff-btn" style="margin-top:8px;padding:8px 16px;background:#0070f3;color:#fff;border:none;border-radius:4px;cursor:pointer;font-size:13px">' + buttonText + '</button>',
      '  <pre class="ff-output" style="margin-top:12px;padding:8px;background:#f5f5f5;border-radius:4px;font-size:12px;overflow:auto;display:none"></pre>',
      '</div>',
    ].join("\n");

    container.innerHTML = html;

    var inputEl  = container.querySelector(".ff-input");
    var btnEl    = container.querySelector(".ff-btn");
    var outputEl = container.querySelector(".ff-output");

    btnEl.addEventListener("click", function () {
      var raw = inputEl.value.trim();
      var input;
      try {
        input = raw ? JSON.parse(raw) : {};
      } catch (_) {
        outputEl.style.display = "block";
        outputEl.textContent   = "Invalid JSON input";
        return;
      }

      btnEl.disabled    = true;
      btnEl.textContent = "Running…";

      run(input, {
        onSuccess: function (data) {
          outputEl.style.display = "block";
          outputEl.textContent   = JSON.stringify(data, null, 2);
          btnEl.disabled    = false;
          btnEl.textContent = buttonText;
          if (options.onSuccess) { options.onSuccess(data); }
        },
        onError: function (err) {
          outputEl.style.display = "block";
          outputEl.textContent   = "Error: " + err.message;
          btnEl.disabled    = false;
          btnEl.textContent = buttonText;
          if (options.onError) { options.onError(err); }
        },
      });
    });
  }

  // ── Public API ────────────────────────────────────────────────────────────
  var api = {
    run:     run,
    form:    form,
    on:      on,
    widget:  widget,
    version: VERSION,
  };

  global[NAMESPACE] = api;

  // Auto-initialize
  if (config.autoload) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", function () { emit("ready", api); });
    } else {
      emit("ready", api);
    }
  }

})(window, { autoload: %s, ui: %s, theme: %q });
`,
		fn.Author, fn.Name, resolvedVersion,
		description,
		fn.Author, fn.Name,
		opts.Namespace,
		baseURL,
		fn.Author,
		fn.Name,
		resolvedVersion,
		opts.Namespace,
		autoloadStr,
		uiStr,
		opts.Theme,
	)
}

// parseEmbedOptions extracts embed options from the HTTP query string.
func parseEmbedOptions(namespace, autoload, ui, theme string) EmbedOptions {
	opts := EmbedOptions{
		Namespace: "ff",
		Autoload:  true,
		UI:        false,
		Theme:     "auto",
	}

	if namespace != "" {
		// Sanitize: only allow valid JS identifier characters
		safe := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$' {
				return r
			}
			return -1
		}, namespace)
		if safe != "" {
			opts.Namespace = safe
		}
	}

	if autoload == "false" || autoload == "0" {
		opts.Autoload = false
	}

	if ui == "true" || ui == "1" {
		opts.UI = true
	}

	if theme == "light" || theme == "dark" {
		opts.Theme = theme
	}

	return opts
}

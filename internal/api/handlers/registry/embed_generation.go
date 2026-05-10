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

	publicURL := getPublicSiteURL()

	// SECURITY: Sanitize values used in JavaScript block comment.
	// If author, name, or description contain "*/", it would break out of the
	// block comment and allow arbitrary JS injection.
	safeAuthor := strings.ReplaceAll(fn.Author, "*/", "* /")
	safeName := strings.ReplaceAll(fn.Name, "*/", "* /")
	safeDesc := strings.ReplaceAll(description, "*/", "* /")
	safeVersion := strings.ReplaceAll(resolvedVersion, "*/", "* /")

	return fmt.Sprintf(`// FunctionFly Embed — %s/%s@%s
// %s
//
// Usage:
//   <script src="%s/embed/%s/%s.js"></script>
//   <script>
//     const result = await %s.run({ key: "value" });
//   </script>
//
// Security: data-api-key is visible in page source. This is acceptable for
// public embeds with rate-limited keys. For sensitive operations, use
// server-side execution via the FunctionFly REST API instead of embeds.
(function (global, config) {
  "use strict";

  var API_BASE  = %q;
  var AUTHOR    = %q;
  var NAME      = %q;
  var VERSION   = %q;
  var NAMESPACE = %q;

  // ── Security warning ──────────────────────────────────────────────────────
  // Warn if API key is exposed on a non-HTTPS page
  (function () {
    var scriptEl = document.currentScript ||
      (function () {
        var scripts = document.getElementsByTagName("script");
        return scripts[scripts.length - 1];
      })();
    var apiKey = scriptEl && scriptEl.getAttribute("data-api-key");
    if (apiKey && window.location.protocol !== "https:") {
      console.warn("[FunctionFly] data-api-key is exposed on a non-HTTPS page. " +
        "Use HTTPS to prevent key interception, or use server-side execution.");
    }
  })();

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
      "Content-Type":   "application/json",
      "User-Agent":     "FunctionFly-Embed/1.0",
      // Phase 3: send the page origin so the server can track embed analytics
      "X-Embed-Origin": window.location.origin,
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
    var handler = function (e) {
      e.preventDefault();
      // Use FormData API to serialize form data
      var formData = new FormData(formEl);
      var data = {};
      formData.forEach(function (value, key) {
        data[key] = value;
      });
      // Support checkbox arrays and other special cases
      var elements = formEl.elements;
      for (var i = 0; i < elements.length; i++) {
        var el = elements[i];
        if (el.type === 'checkbox' || el.type === 'radio') {
          if (el.name && !el.checked) { delete data[el.name]; }
          if (el.name && el.checked && !data[el.name]) { data[el.name] = el.value || true; }
        }
      }
      run(data, {
        onSuccess: options.onSuccess,
        onError: options.onError,
      });
    };
    formEl.addEventListener('submit', handler);
    // Return cleanup function
    return function () {
      formEl.removeEventListener('submit', handler);
    };
  }

  // ── Theme detection ───────────────────────────────────────────────────────
  function getTheme(preferredTheme) {
    if (preferredTheme === 'light' || preferredTheme === 'dark') {
      return preferredTheme;
    }
    // Auto-detect based on system preference
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return 'dark';
    }
    return 'light';
  }

  // ── UI widget ─────────────────────────────────────────────────────────────
  function widget(container, options) {
    options = options || {};
    var title       = options.title       || (AUTHOR + "/" + NAME);
    var placeholder = options.placeholder || "Enter input (JSON)...";
    var buttonText  = options.buttonText  || "Run";
    var theme       = getTheme(options.theme || config.theme || "auto");

    // Determine colors based on theme
    var isDark = theme === 'dark';
    var bgColor       = isDark ? '#1e1e1e' : '#ffffff';
    var borderColor   = isDark ? '#444444' : '#dddddd';
    var textColor     = isDark ? '#e0e0e0' : '#333333';
    var inputBg       = isDark ? '#2d2d2d' : '#f9f9f9';
    var outputBg      = isDark ? '#2a2a2a' : '#f5f5f5';
    var btnBg         = '#0070f3';
    var btnHover      = '#005bb5';
    var errorColor    = isDark ? '#ff6b6b' : '#dc3545';

    // Build self-contained widget HTML with inline CSS
    var html = [
      '<div class="ff-widget" style="',
        'font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;',
        'border: 1px solid ' + borderColor + ';',
        'border-radius: 8px;',
        'padding: 16px;',
        'max-width: 480px;',
        'background: ' + bgColor + ';',
        'color: ' + textColor + ';',
        'box-sizing: border-box;'
      ).replace(/\s+/g, ' ').trim(),
      '">',
      '  <h3 style="margin: 0 0 12px; font-size: 14px; font-weight: 600; color: ' + textColor + ';">' + escHtml(title) + '</h3>',
      '  <textarea class="ff-input" placeholder="' + escHtml(placeholder) + '" rows="4"',
      '    style="width: 100%%; box-sizing: border-box; padding: 8px; border: 1px solid ' + borderColor + '; border-radius: 4px; font-size: 13px; resize: vertical; background: ' + inputBg + '; color: ' + textColor + ';"></textarea>',
      '  <button class="ff-btn" style="margin-top: 8px; padding: 8px 16px; background: ' + btnBg + '; color: #fff; border: none; border-radius: 4px; cursor: pointer; font-size: 13px; font-weight: 500;">' + escHtml(buttonText) + '</button>',
      '  <pre class="ff-output" style="margin-top: 12px; padding: 8px; background: ' + outputBg + '; border-radius: 4px; font-size: 12px; overflow: auto; display: none; color: ' + textColor + ';"></pre>',
      '</div>',
    ].join('');

    container.innerHTML = html;

    var inputEl  = container.querySelector('.ff-input');
    var btnEl    = container.querySelector('.ff-btn');
    var outputEl = container.querySelector('.ff-output');

    // Add button hover effect
    btnEl.addEventListener('mouseenter', function () { btnEl.style.background = btnHover; });
    btnEl.addEventListener('mouseleave', function () { btnEl.style.background = btnBg; });

    btnEl.addEventListener('click', function () {
      var raw = inputEl.value.trim();
      var input;
      try {
        input = raw ? JSON.parse(raw) : {};
      } catch (_) {
        outputEl.style.display = 'block';
        outputEl.style.color = errorColor;
        outputEl.textContent   = 'Invalid JSON input';
        return;
      }

      // Show loading state
      btnEl.disabled    = true;
      btnEl.textContent = 'Running…';
      outputEl.style.display = 'none';

      run(input, {
        onSuccess: function (result) {
          var data = result.data || result;
          outputEl.style.display = 'block';
          outputEl.style.color = textColor;
          outputEl.textContent   = JSON.stringify(data, null, 2);
          btnEl.disabled    = false;
          btnEl.textContent = buttonText;
          if (options.onSuccess) { options.onSuccess(data); }
        },
        onError: function (err) {
          outputEl.style.display = 'block';
          outputEl.style.color = errorColor;
          outputEl.textContent   = 'Error: ' + (err.message || 'Execution failed');
          btnEl.disabled    = false;
          btnEl.textContent = buttonText;
          if (options.onError) { options.onError(err); }
        },
      });
    });

    // Return cleanup function
    return function () {
      container.innerHTML = '';
    };
  }

  // HTML escape utility to prevent XSS
  function escHtml(str) {
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;')
      .replace(/`+"`"+`/g, '&#96;');
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
		safeAuthor, safeName, safeVersion,
		safeDesc,
		publicURL, safeAuthor, safeName,
		opts.Namespace,
		baseURL,
		safeAuthor,
		safeName,
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

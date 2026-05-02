/**
 * FunctionFly Unified Theme System - Vanilla JS Version
 * For use in Astro pages and vanilla JS contexts.
 *
 * Exposes global `ffTheme` object:
 *   ffTheme.init()         - Initialize theme
 *   ffTheme.get()          - Get current theme state
 *   ffTheme.set(mode)      - Set theme mode
 *   ffTheme.toggle()       - Toggle between light/dark
 *   ffTheme.subscribe(fn)  - Listen for changes
 *
 * Storage: 'ff-user-theme' localStorage key (unified across all apps)
 * Broadcasts changes via 'ff-theme-change' custom events
 */
(function(global) {
  'use strict';

  var STORAGE_KEY = 'ff-user-theme';
  var LEGACY_KEYS = ['theme-storage', 'ff-docs-theme'];

  var _cachedState = null;
  var _listeners = [];

  function getSystemTheme() {
    if (typeof window === 'undefined') return 'dark';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  function resolveTheme(mode) {
    if (mode === 'system') return getSystemTheme();
    return mode;
  }

  function readStorage() {
    if (typeof window === 'undefined') return null;
    try {
      var stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        var parsed = JSON.parse(stored);
        if (parsed.mode === 'light' || parsed.mode === 'dark' || parsed.mode === 'system') {
          return parsed;
        }
      }
    } catch (e) {}
    return null;
  }

  function writeStorage(preference) {
    if (typeof window === 'undefined') return;
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(preference));
    } catch (e) {}
  }

  function migrateLegacy() {
    if (typeof window === 'undefined') return null;
    for (var i = 0; i < LEGACY_KEYS.length; i++) {
      var key = LEGACY_KEYS[i];
      try {
        var legacy = localStorage.getItem(key);
        if (!legacy) continue;

        if (key === 'theme-storage') {
          try {
            var parsed = JSON.parse(legacy);
            if (parsed.state && parsed.state.theme) {
              var mode = parsed.state.theme;
              if (mode === 'light' || mode === 'dark' || mode === 'system') {
                writeStorage({ mode: mode });
                localStorage.removeItem(key);
                return { mode: mode };
              }
            }
          } catch (e) {}
        } else if (key === 'ff-docs-theme') {
          if (legacy === 'light' || legacy === 'dark') {
            writeStorage({ mode: legacy });
            localStorage.removeItem(key);
            return { mode: legacy };
          }
        }
      } catch (e) {}
    }
    return null;
  }

  function applyToDOM(resolved) {
    if (typeof document === 'undefined') return;
    document.documentElement.setAttribute('data-theme', resolved);
  }

  function broadcast(state) {
    if (typeof window === 'undefined') return;
    window.dispatchEvent(new CustomEvent('ff-theme-change', { detail: state }));
  }

  function get() {
    if (_cachedState) return _cachedState;
    if (typeof window === 'undefined') {
      return { mode: 'system', resolved: 'dark' };
    }
    var migrated = migrateLegacy();
    var preference = migrated || readStorage();
    var mode = preference ? preference.mode : 'system';
    var resolved = resolveTheme(mode);
    _cachedState = { mode: mode, resolved: resolved };
    return _cachedState;
  }

  function set(mode) {
    if (typeof window === 'undefined') return;
    writeStorage({ mode: mode });
    var resolved = resolveTheme(mode);
    _cachedState = { mode: mode, resolved: resolved };
    applyToDOM(resolved);
    broadcast(_cachedState);
  }

  function toggle() {
    var current = get();
    var next = current.resolved === 'dark' ? 'light' : 'dark';
    set(next);
  }

  function subscribe(callback) {
    _listeners.push(callback);
    return function unsubscribe() {
      for (var i = 0; i < _listeners.length; i++) {
        if (_listeners[i] === callback) {
          _listeners.splice(i, 1);
          return;
        }
      }
    };
  }

  function notifyListeners(state) {
    for (var i = 0; i < _listeners.length; i++) {
      try {
        _listeners[i](state);
      } catch (e) {}
    }
  }

  function init() {
    if (typeof window === 'undefined') return;
    var state = get();
    applyToDOM(state.resolved);

    window.addEventListener('ff-theme-change', function(e) {
      _cachedState = e.detail;
      applyToDOM(e.detail.resolved);
      notifyListeners(e.detail);
    });
  }

  var ffTheme = {
    init: init,
    get: get,
    set: set,
    toggle: toggle,
    subscribe: subscribe
  };

  global.ffTheme = ffTheme;

})(typeof window !== 'undefined' ? window : this);
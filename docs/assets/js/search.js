/**
 * Search Module for Parallels DevOps Documentation
 * Provides a modal search interface with keyboard shortcuts (Ctrl/Cmd+K)
 * Uses Simple-Jekyll-Search loaded from CDN
 */

(function () {
  'use strict';

  // ---- State ----
  var searchModal = null;
  var searchInput = null;
  var searchResults = null;
  var activeResultIndex = -1;
  var searchData = [];

  // ---- DOM References ----
  function getElements() {
    searchModal = document.getElementById('search-overlay');
    searchInput = document.getElementById('search-input');
    searchResults = document.getElementById('search-results');
  }

  // ---- Open / Close ----
  function openSearch() {
    if (!searchModal) return;
    searchModal.classList.add('is-active');
    searchModal.style.display = 'flex';
    // Force reflow for transition
    setTimeout(function () {
      searchModal.style.opacity = '1';
    }, 10);
    if (searchInput) {
      setTimeout(function () { searchInput.focus(); }, 100);
    }
    document.body.style.overflow = 'hidden';
  }

  function closeSearch() {
    if (!searchModal) return;
    searchModal.style.opacity = '0';
    setTimeout(function () {
      searchModal.classList.remove('is-active');
      searchModal.style.display = 'none';
    }, 200);
    document.body.style.overflow = '';
    if (searchInput) {
      searchInput.value = '';
    }
    activeResultIndex = -1;
  }

  // ---- Highlight matching text ----
  function highlightMatch(text, query) {
    if (!query || !text) return escapeHtml(text);
    var escaped = escapeHtml(text);
    var queryLower = query.toLowerCase();
    var textLower = escaped.toLowerCase();
    var idx = textLower.indexOf(queryLower);
    if (idx === -1) return escaped;
    return escaped.substring(0, idx) +
      '<mark>' + escaped.substring(idx, idx + query.length) + '</mark>' +
      escaped.substring(idx + query.length);
  }

  function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  // ---- Render Results ----
  function renderResults(results, query) {
    if (!searchResults) return;

    if (results.length === 0) {
      searchResults.innerHTML =
        '<div class="search-results-empty">' +
        '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>' +
        '<p>No results found for "' + escapeHtml(query) + '"</p>' +
        '</div>';
      return;
    }

    // Group results by category
    var grouped = {};
    results.forEach(function (result) {
      var cat = result.category || 'Documentation';
      if (!grouped[cat]) grouped[cat] = [];
      grouped[cat].push(result);
    });

    var html = '';
    Object.keys(grouped).forEach(function (cat) {
      html += '<div class="search-result-item-separator">' + escapeHtml(cat) + '</div>';
      grouped[cat].forEach(function (result) {
        var excerpt = result.content || '';
        if (excerpt.length > 150) excerpt = excerpt.substring(0, 150) + '...';
        html += '<a href="' + result.url + '" class="search-result-item" data-url="' + result.url + '">' +
          '<div class="search-result-item-title">' + highlightMatch(result.title, query) + '</div>' +
          '<div class="search-result-item-excerpt">' + highlightMatch(excerpt, query) + '</div>' +
          '</a>';
      });
    });

    searchResults.innerHTML = html;
    activeResultIndex = -1;

    // Attach click handlers
    searchResults.querySelectorAll('.search-result-item').forEach(function (item) {
      item.addEventListener('click', function () {
        // Navigate happens naturally via <a> tag
      });
    });
  }

  // ---- Keyboard Navigation ----
  function handleKeyDown(e) {
    var items = searchResults ? searchResults.querySelectorAll('.search-result-item') : [];
    if (!items.length) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      activeResultIndex = Math.min(activeResultIndex + 1, items.length - 1);
      updateActiveResult(items);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      activeResultIndex = Math.max(activeResultIndex - 1, 0);
      updateActiveResult(items);
    } else if (e.key === 'Enter' && activeResultIndex >= 0) {
      e.preventDefault();
      items[activeResultIndex].click();
    }
  }

  function updateActiveResult(items) {
    items.forEach(function (item, i) {
      item.classList.toggle('is-active', i === activeResultIndex);
      if (i === activeResultIndex) {
        item.scrollIntoView({ block: 'nearest' });
      }
    });
  }

  // ---- Detect baseurl dynamically ----
  function detectBaseurl() {
    // Strategy 1: Look for a <link rel="stylesheet"> that contains site CSS
    var cssLink = document.querySelector('link[rel="stylesheet"]');
    if (cssLink) {
      var href = cssLink.getAttribute('href');
      if (href && href.indexOf('/assets/') !== -1) {
        // Extract baseurl from CSS path, e.g. /prl-devops-service/assets/css/app.css -> /prl-devops-service
        return href.substring(0, href.indexOf('/assets/'));
      }
    }

    // Strategy 2: Look at window.location.pathname and strip known doc paths
    var loc = window.location.pathname;
    // Remove trailing slash
    loc = loc.replace(/\/+$/, '');

    // If path starts with something like /prl-devops-service/docs/...
    var match = loc.match(/^\/([^/]+)(?:\/(docs|features|solutions|ui|blog))?\//);
    if (match && match[1]) {
      // Check if this looks like a baseurl (not "docs" itself)
      var candidate = '/' + match[1];
      if (match[2]) {
        // We found /baseurl/docs/ pattern
        return candidate;
      }
    }

    // Strategy 3: Check if we're at a root-like path under a subdirectory
    // e.g., /prl-devops-service/ or just /
    var parts = loc.split('/').filter(Boolean);
    if (parts.length === 1) {
      // Single segment like /prl-devops-service
      return '/' + parts[0];
    }

    // Fallback: no baseurl
    return '';
  }

  // ---- Load Search Data ----
  function loadSearchData(callback) {
    var xhr = new XMLHttpRequest();
    // Detect baseurl from page context.
    // Strategy: find any link that points into the site (e.g. a nav link or CSS),
    // then strip known doc paths to isolate the baseurl.
    // Fallback: try common patterns.
    var baseurl = detectBaseurl();
    var searchUrl = baseurl + '/search.json';
    xhr.open('GET', searchUrl, true);
    xhr.setRequestHeader('Accept', 'application/json');
    xhr.onload = function () {
      if (xhr.status === 200) {
        try {
          searchData = JSON.parse(xhr.responseText);
          if (callback) callback(null, searchData);
        } catch (e) {
          console.error('[Search] Failed to parse search.json:', e);
        }
      } else {
        console.warn('[Search] Could not load search.json (HTTP ' + xhr.status + '). Search will be empty.');
      }
    };
    xhr.onerror = function () {
      console.warn('[Search] Network error loading search.json');
    };
    xhr.send();
  }

  // ---- Search Function ----
  function performSearch(query) {
    if (!query || query.trim().length === 0) {
      if (searchResults) searchResults.innerHTML = '';
      return;
    }

    var lowerQuery = query.toLowerCase();
    var results = searchData.filter(function (item) {
      var title = (item.title || '').toLowerCase();
      var content = (item.content || '').toLowerCase();
      var category = (item.category || '').toLowerCase();
      return title.indexOf(lowerQuery) !== -1 ||
             content.indexOf(lowerQuery) !== -1 ||
             category.indexOf(lowerQuery) !== -1;
    });

    // Sort: title matches first, then content
    results.sort(function (a, b) {
      var aTitle = (a.title || '').toLowerCase().indexOf(lowerQuery);
      var bTitle = (b.title || '').toLowerCase().indexOf(lowerQuery);
      if (aTitle !== -1 && bTitle === -1) return -1;
      if (aTitle === -1 && bTitle !== -1) return 1;
      return 0;
    });

    // Limit results
    if (results.length > 15) results = results.slice(0, 15);

    renderResults(results, query);
  }

  // ---- Initialize ----
  function init() {
    getElements();
    if (!searchInput || !searchModal) {
      console.warn('[Search] Search elements not found in DOM');
      return;
    }

    // Load search data
    loadSearchData(function () {
      // Search input handler with debounce
      var debounceTimer = null;
      searchInput.addEventListener('input', function () {
        clearTimeout(debounceTimer);
        var query = searchInput.value;
        debounceTimer = setTimeout(function () {
          performSearch(query);
        }, 150);
      });

      // Close on result click
      searchInput.addEventListener('blur', function () {
        // Delay close to allow click on result
        setTimeout(function () {
          if (!searchInput.matches(':focus')) {
            closeSearch();
          }
        }, 200);
      });
    });

    // Close button
    var closeBtn = searchModal.querySelector('.search-modal-close');
    if (closeBtn) {
      closeBtn.addEventListener('click', closeSearch);
    }

    // Overlay click to close
    searchModal.addEventListener('click', function (e) {
      if (e.target === searchModal) {
        closeSearch();
      }
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', function (e) {
      // Ctrl/Cmd + K to open
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        if (searchModal && searchModal.classList.contains('is-active')) {
          closeSearch();
        } else {
          openSearch();
        }
        return;
      }

      // Escape to close
      if (e.key === 'Escape' && searchModal && searchModal.classList.contains('is-active')) {
        e.preventDefault();
        closeSearch();
        return;
      }

      // Arrow navigation when modal is open
      if (searchModal && searchModal.classList.contains('is-active')) {
        handleKeyDown(e);
      }
    });
  }

  // Expose open/close globally for header trigger
  window.openSearchModal = openSearch;
  window.closeSearchModal = closeSearch;

  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();

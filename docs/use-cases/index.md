---
layout: default
title: Use Cases
permalink: /docs/use-cases/
---

<!-- Use Cases Index Page -->
<div class="use-cases-index">
  <div class="uci-header">
    <h1>Use Cases</h1>
    <p class="uci-description">
      Interactive guided experiences for learning our DevOps services.
    </p>
    <div class="uci-controls">
      <input type="text" id="uci-search" placeholder="Search use cases..." class="uci-search-input">
      <div class="uci-filter-tags" id="uci-filter-tags"></div>
    </div>
  </div>

  <div class="uci-grid" id="uci-grid">
    {% assign use_cases = site.pages | where_exp: "item", "item.path contains 'use-cases/' and item.path != 'use-cases/index.md'" | sort: 'order' %}
    {% for uc in use_cases %}
      {% assign uc_key = uc.uce_data | default: uc.id %}
      {% assign uc_data = site.data[uc_key] | default: uc %}
      {% assign depends_list = uc_data.depends_on | join: ',' %}
      {% assign scenario = uc_data.scenario | default: uc_data.introduction.scenario %}
      {% assign scenario_md = uc_data.markdown_scenario | default: uc_data.introduction.markdown_scenario %}
    {% assign scenario_display = '' %}
    {% assign scenario_plain = '' %}
    {% if scenario_md %}
      {% assign scenario_display = scenario_md | markdownify | strip_newlines %}
      {% assign scenario_plain = scenario_md | markdownify | strip_newlines | replace: '<p>', '' | replace: '</p>', '' | replace: '<br>', ' ' | replace: '<strong>', '' | replace: '</strong>', '' | replace: '<em>', '' | replace: '</em>', '' | replace: '<b>', '' | replace: '</b>', '' | replace: '<i>', '' | replace: '</i>', '' | strip %}
    {% elsif scenario %}
      {% assign scenario_display = scenario %}
      {% assign scenario_plain = scenario %}
    {% endif %}
    {% if uc_data.hidden %}{% continue %}{% endif %}
    <article class="uci-card"
      data-title="{{ uc_data.title | escape }}"
      data-scenario="{{ scenario_plain | escape | strip }}"
      data-category="{{ uc_data.category | default: 'Uncategorized' | slugify }}"
      data-group="{{ uc_data.group | default: 'General' | slugify }}"
      data-level="{{ uc_data.level | default: '' }}"
      data-uce-id="{{ uc_data.id | default: '' }}"
      data-tags="{{ uc_data.tags | join: ',' }}"
      data-unlocks="{{ uc_data.unlocks | size }}"
      data-depends-on="{{ depends_list }}">
      <div class="uci-card-inner">
        <div class="uci-card-meta">
          {% if uc_data.level %}
          <span class="uci-badge uci-badge--{{ uc_data.level }}">{{ uc_data.level | capitalize }}</span>
          {% endif %}
          {% if uc_data.duration %}
          <span class="uci-meta-item">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
            </svg>
            {{ uc_data.duration }}
          </span>
          {% endif %}
        </div>
        <h2 class="uci-card-title">
          <a href="{{ site.baseurl }}{{ uc.url }}">{{ uc_data.title | default: 'Untitled' }}</a>
        </h2>
        {% if scenario_display %}
        <div class="uci-card-desc">{{ scenario_plain | truncatewords: 30 }}</div>
        {% endif %}
        {% if uc_data.tags %}
        <div class="uci-card-tags">
          {% for tag in uc_data.tags %}
          <span class="uci-tag">{{ tag }}</span>
          {% endfor %}
        </div>
        {% endif %}
        {% if depends_list != '' and depends_list != nil %}
        <div class="uci-card-depends" id="uci-depends-{{ forloop.index }}">
          <span class="uci-dep-label">Requires:</span>
          {% assign dep_array = depends_list | split: "," %}
          {% for dep in dep_array %}
          <span class="uci-dep-tag" data-dep-slug="{{ dep | strip }}">Loading...</span>
          {% endfor %}
        </div>
        {% endif %}
      </div>
    </article>
    {% endfor %}
  </div>

  <div class="uci-empty" id="uci-empty" style="display:none;">
    <p>No use cases match your search.</p>
  </div>
</div>

<script>
(function() {
  var cards = document.querySelectorAll('.uci-card');
  var searchInput = document.getElementById('uci-search');
  var filterTags = document.getElementById('uci-filter-tags');
  var emptyMsg = document.getElementById('uci-empty');

  var allTags = {};
  cards.forEach(function(card) {
    var tags = (card.dataset.tags || '').split(',');
    tags.forEach(function(tag) {
      tag = tag.trim();
      if (tag) allTags[tag] = true;
    });
  });

  Object.keys(allTags).slice(0, 15).forEach(function(tag) {
    var btn = document.createElement('button');
    btn.className = 'uci-filter-btn';
    btn.textContent = tag;
    btn.setAttribute('data-tag', tag);
    btn.addEventListener('click', function() {
      btn.classList.toggle('uci-filter-btn--active');
      filterCards();
    });
    filterTags.appendChild(btn);
  });

  function normalize(str) {
    return (str || '').toLowerCase().replace(/[^a-z0-9]+/g, ' ').replace(/^\s+|\s+$/g, '');
  }

  function fuzzyMatch(text, query) {
    if (!query) return { match: true, score: 0 };
    var normText = normalize(text);
    var normQuery = normalize(query);
    var tokens = normQuery.split(/\s+/).filter(Boolean);
    if (tokens.length === 0) return { match: true, score: 0 };

    var totalScore = 0;
    var matchedTokens = 0;

    tokens.forEach(function(token) {
      // Exact substring match (highest score)
      if (normText.indexOf(token) !== -1) {
        matchedTokens++;
        totalScore += 10;
        return;
      }

      // Word-boundary match
      var escaped = token.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      try {
        var wordRe = new RegExp('\\b' + escaped + '\\b', 'i');
        if (wordRe.test(normText)) {
          matchedTokens++;
          totalScore += 8;
          return;
        }
      } catch(e) { /* skip */ }

      // Levenshtein distance for typo tolerance
      var maxDist = token.length <= 5 ? 2 : 3;
      if (levenshteinDistance(normText, token) <= maxDist) {
        matchedTokens++;
        totalScore += 4;
        return;
      }

      // Substring match within a single word
      var found = false;
      var words = normText.split(/\s+/);
      for (var w = 0; w < words.length; w++) {
        if (levenshteinDistance(words[w], token) <= Math.max(1, Math.floor(token.length / 3))) {
          matchedTokens++;
          totalScore += 3;
          found = true;
          break;
        }
      }
      if (!found) {
        for (var w2 = 0; w2 < words.length; w2++) {
          if (words[w2].indexOf(token) !== -1) {
            matchedTokens++;
            totalScore += 1;
            found = true;
            break;
          }
        }
      }
    });

    var threshold = Math.ceil(tokens.length / 2);
    return { match: matchedTokens >= threshold, score: totalScore };
  }

  function levenshteinDistance(a, b) {
    var la = a.length;
    var lb = b.length;
    var maxLen = Math.max(la, lb);
    if (maxLen > 20) return 999;
    if (la === 0) return lb;
    if (lb === 0) return la;

    var prevRow = [];
    for (var i = 0; i <= la; i++) prevRow[i] = i;

    for (var j = 0; j < lb; j++) {
      var currRow = [j + 1];
      for (var i = 0; i < la; i++) {
        var cost = a[i] === b[j] ? 0 : 1;
        var insertVal = currRow[i] + 1;
        var deleteVal = prevRow[i + 1] + 1;
        var replaceVal = prevRow[i] + cost;
        currRow[i + 1] = Math.min(insertVal, deleteVal, replaceVal);
      }
      prevRow = currRow;
    }
    return prevRow[la];
  }

  function filterCards() {
    var query = (searchInput.value || '').toLowerCase().trim();
    var activeTags = [];
    document.querySelectorAll('.uci-filter-btn--active').forEach(function(btn) {
      activeTags.push(btn.dataset.tag);
    });

    var visible = 0;
    var scoredCards = [];

    cards.forEach(function(card) {
      var title = card.dataset.title || '';
      var scenario = card.dataset.scenario || '';
      var tags = (card.dataset.tags || '').toLowerCase();
      var category = (card.dataset.category || '').toLowerCase();

      var matchesTags = activeTags.length === 0 || activeTags.some(function(t) { return tags.includes(t); });

      if (!matchesTags) {
        card.style.display = 'none';
        return;
      }

      if (!query) {
        card.style.display = '';
        visible++;
        return;
      }

      var titleResult = fuzzyMatch(title, query);
      var scenarioResult = fuzzyMatch(scenario, query);
      var tagsResult = fuzzyMatch(tags, query);

      var totalScore = 0;
      var anyMatch = false;

      if (titleResult.match) { totalScore += titleResult.score * 3; anyMatch = true; }
      if (scenarioResult.match) { totalScore += scenarioResult.score * 1.5; anyMatch = true; }
      if (tagsResult.match) { totalScore += tagsResult.score * 2; anyMatch = true; }

      if (anyMatch) {
        card.style.display = '';
        visible++;
        scoredCards.push({ card: card, score: totalScore });
      } else {
        card.style.display = 'none';
      }
    });

    if (visible > 0 && query) {
      scoredCards.sort(function(a, b) { return b.score - a.score; });
      var grid = document.getElementById('uci-grid');
      scoredCards.forEach(function(item) {
        if (item.card.style.display !== 'none') {
          grid.appendChild(item.card);
        }
      });
    }

    emptyMsg.style.display = visible === 0 ? '' : 'none';
  }

  cards.forEach(function(card) {
    card.style.cursor = 'pointer';
    card.addEventListener('click', function(e) {
      if (e.target.tagName === 'A') return;
      var link = card.querySelector('.uci-card-title a');
      if (link) window.location.href = link.getAttribute('href');
    });
  });

  var STORAGE_PREFIX = 'uce_progress:';
  cards.forEach(function(card) {
    var uceId = (card.dataset.uceId || '').trim();
    if (!uceId) return;
    try {
      var raw = localStorage.getItem(STORAGE_PREFIX + uceId);
      if (!raw) return;
      var state = JSON.parse(raw);
      if (!state) return;

      var badge = null;

      if (state.current_step === '__complete__') {
        card.classList.add('uci-card--completed');
        badge = document.createElement('span');
        badge.className = 'uci-badge uci-badge--completed';
        badge.innerHTML = '<svg viewBox="0 0 24 24" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg> Done';
      }
      else if (state.completed_steps && state.completed_steps.length > 0) {
        badge = document.createElement('span');
        badge.className = 'uci-badge uci-badge--progress';
        badge.textContent = state.completed_steps.length + ' steps done';
      }

      if (badge) {
        var meta = card.querySelector('.uci-card-meta');
        if (meta) {
          var badgeWrapper = document.createElement('div');
          badgeWrapper.className = 'uci-badge-wrapper';
          badgeWrapper.appendChild(badge);
          meta.appendChild(badgeWrapper);
        }
      }
    } catch (e) {
      // Ignore corrupted localStorage data
    }
  });

  var slugToUceId = {};
  cards.forEach(function(c) {
    var slug = c.dataset.title ? c.querySelector('.uci-card-title a') : null;
    var uceId = (c.dataset.uceId || '').trim();
    if (!slug || !uceId) return;
    var href = slug.getAttribute('href');
    if (href) {
      var parts = href.split('/').filter(Boolean);
      var pageSlug = parts[parts.length - 1];
      if (pageSlug) slugToUceId[pageSlug] = uceId;
    }
  });

  var slugToTitle = {};
  cards.forEach(function(c) {
    var uceId = (c.dataset.uceId || '').trim();
    var title = (c.dataset.title || '').trim();
    if (!uceId || !title) return;
    slugToTitle[uceId] = title;
  });

  var MAX_TAG_LENGTH = 20;

  function escapeHtml(str) {
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  var depCards = document.querySelectorAll('[data-depends-on]');
  depCards.forEach(function(card) {
    var depsStr = (card.dataset.dependsOn || '').trim();
    if (!depsStr) return;

    var allDone = true;

    var depTags = card.querySelectorAll('.uci-dep-tag');
    depTags.forEach(function(tag) {
      var depSlug = tag.dataset.depSlug;
      var lookupKey = slugToUceId[depSlug] || depSlug;
      var fullTitle = slugToTitle[lookupKey] || depSlug.replace(/-/g, ' ');
      var displayTitle = fullTitle.length > MAX_TAG_LENGTH
        ? fullTitle.slice(0, MAX_TAG_LENGTH).trimEnd() + '\u2026'
        : fullTitle;

      try {
        var raw = localStorage.getItem(STORAGE_PREFIX + lookupKey);
        if (!raw) {
          tag.innerHTML = '<i class="fa-solid fa-lock"></i> ' + escapeHtml(displayTitle);
          tag.className = 'uci-dep-tag uci-dep-tag--locked';
          tag.title = 'Complete "' + escapeHtml(fullTitle) + '" first to unlock this use case';
          allDone = false;
          return;
        }
        var state = JSON.parse(raw);
        if (state && state.current_step === '__complete__') {
          tag.innerHTML = '<i class="fa-solid fa-circle-check"></i> ' + escapeHtml(displayTitle);
          tag.className = 'uci-dep-tag uci-dep-tag--done';
          tag.title = escapeHtml(fullTitle) + ' \u2014 completed';
        } else {
          tag.innerHTML = '<i class="fa-solid fa-clock"></i> ' + escapeHtml(displayTitle);
          tag.className = 'uci-dep-tag uci-dep-tag--pending';
          tag.title = escapeHtml(fullTitle) + ' \u2014 in progress';
          allDone = false;
        }
      } catch (e) {
        tag.textContent = displayTitle;
        tag.className = 'uci-dep-tag uci-dep-tag--locked';
        allDone = false;
      }
    });

    if (!allDone) {
      card.classList.add('uci-card--partial');
    }
  });

  searchInput.addEventListener('input', filterCards);
})();
</script>

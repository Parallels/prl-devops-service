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
    <article class="uci-card"
      data-title="{{ uc.title | escape }}"
      data-category="{{ uc.category | default: 'Uncategorized' | slugify }}"
      data-group="{{ uc.group | default: 'General' | slugify }}"
      data-level="{{ uc.level | default: '' }}"
      data-uce-id="{{ uc.uce_data | default: '' }}"
      data-tags="{{ uc.tags | join: ',' }}"
      data-unlocks="{{ uc.unlocks | size }}">
      <div class="uci-card-inner">
        <div class="uci-card-meta">
          {% if uc.level %}
          <span class="uci-badge uci-badge--{{ uc.level }}">{{ uc.level | capitalize }}</span>
          {% endif %}
          {% if uc.duration %}
          <span class="uci-meta-item">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
            </svg>
            {{ uc.duration }}
          </span>
          {% endif %}
        </div>
        <h2 class="uci-card-title">
          <a href="{{ site.baseurl }}{{ uc.url }}">{{ uc.title | default: 'Untitled' }}</a>
        </h2>
        {% if uc.scenario %}
        <p class="uci-card-desc">{{ uc.scenario | truncatewords: 30 }}</p>
        {% endif %}
        {% if uc.tags %}
        <div class="uci-card-tags">
          {% for tag in uc.tags %}
          <span class="uci-tag">{{ tag }}</span>
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

  function filterCards() {
    var query = (searchInput.value || '').toLowerCase().trim();
    var activeTags = [];
    document.querySelectorAll('.uci-filter-btn--active').forEach(function(btn) {
      activeTags.push(btn.dataset.tag);
    });

    var visible = 0;
    cards.forEach(function(card) {
      var title = (card.dataset.title || '').toLowerCase();
      var tags = (card.dataset.tags || '').toLowerCase();
      var category = (card.dataset.category || '').toLowerCase();

      var matchesQuery = !query || title.includes(query) || tags.includes(query) || category.includes(query);
      var matchesTags = activeTags.length === 0 || activeTags.some(function(t) { return tags.includes(t); });

      if (matchesQuery && matchesTags) {
        card.style.display = '';
        visible++;
      } else {
        card.style.display = 'none';
      }
    });

    emptyMsg.style.display = visible === 0 ? '' : 'none';
  }

  // Click on entire card navigates to the card's link
  cards.forEach(function(card) {
    card.style.cursor = 'pointer';
    card.addEventListener('click', function(e) {
      // Don't double-navigate if clicking a nested link
      if (e.target.tagName === 'A') return;
      var link = card.querySelector('.uci-card-title a');
      if (link) window.location.href = link.getAttribute('href');
    });
  });

  // Mark cards for use cases that are completed or in progress
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

      // Fully completed — green checkmark badge
      if (state.current_step === '__complete__') {
        card.classList.add('uci-card--completed');
        badge = document.createElement('span');
        badge.className = 'uci-badge uci-badge--completed';
        badge.innerHTML = '<svg viewBox="0 0 24 24" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg> Done';
      }
      // Partially completed — show progress
      else if (state.completed_steps && state.completed_steps.length > 0) {
        badge = document.createElement('span');
        badge.className = 'uci-badge uci-badge--progress';
        badge.textContent = state.completed_steps.length + ' steps done';
      }

      if (badge) {
        var meta = card.querySelector('.uci-card-meta');
        if (meta) meta.appendChild(badge);
      }
    } catch (e) {
      // Ignore corrupted localStorage data
    }
  });

  searchInput.addEventListener('input', filterCards);
})();
</script>

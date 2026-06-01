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
    {% assign use_cases = site.pages | where_exp: "item", "item.path contains 'use-cases/' and item.path != 'index.md'" | sort: 'title' %}
    {% for uc in use_cases %}
    <article class="uci-card"
      data-title="{{ uc.title | escape }}"
      data-category="{{ uc.category | default: 'Uncategorized' | slugify }}"
      data-group="{{ uc.group | default: 'General' | slugify }}"
      data-level="{{ uc.level | default: '' }}"
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

  searchInput.addEventListener('input', filterCards);
})();
</script>

<style>
.use-cases-index {
  max-width: 1024px;
  margin: 0 auto;
  padding: 32px 24px;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.uci-header {
  margin-bottom: 32px;
}

.uci-header h1 {
  font-size: 28px;
  font-weight: 800;
  color: #303236;
  margin: 0 0 8px;
}

.uci-description {
  font-size: 15px;
  color: #7b7d85;
  margin: 0 0 20px;
}

.uci-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.uci-search-input {
  padding: 8px 16px;
  border: 1px solid #e2e2e2;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  width: 280px;
  max-width: 100%;
  transition: border-color 0.2s;
}

.uci-search-input:focus {
  outline: none;
  border-color: #3d73d8;
}

.uci-filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.uci-filter-btn {
  padding: 4px 10px;
  border: 1px solid #e2e2e2;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  background: #fff;
  color: #7b7d85;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.uci-filter-btn--active {
  background: #3d73d8;
  color: #fff;
  border-color: #3d73d8;
}

.uci-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.uci-card {
  background: #fff;
  border: 1px solid #e2e2e2;
  border-radius: 12px;
  overflow: hidden;
  transition: box-shadow 0.2s, border-color 0.2s;
}

.uci-card:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
  border-color: #3d73d8;
}

.uci-card-inner {
  padding: 20px;
}

.uci-card-meta {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.uci-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.uci-badge--beginner {
  background: rgba(34,197,94,0.1);
  color: #22c55e;
}

.uci-badge--intermediate {
  background: rgba(59,130,246,0.1);
  color: #3b82f6;
}

.uci-badge--advanced {
  background: rgba(239,68,68,0.1);
  color: #ef4444;
}

.uci-meta-item {
  font-size: 12px;
  color: #7b7d85;
  display: flex;
  align-items: center;
  gap: 4px;
}

.uci-card-title {
  font-size: 16px;
  font-weight: 700;
  color: #303236;
  margin: 0 0 8px;
}

.uci-card-title a {
  color: inherit;
  text-decoration: none;
}

.uci-card-title a:hover {
  color: #3d73d8;
}

.uci-card-desc {
  font-size: 14px;
  color: #7b7d85;
  line-height: 1.5;
  margin: 0 0 12px;
}

.uci-card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.uci-tag {
  padding: 2px 8px;
  background: #f6f6f6;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #7b7d85;
}

.uci-empty {
  text-align: center;
  padding: 48px 24px;
  color: #7b7d85;
  font-size: 15px;
}

@media screen and (max-width: 600px) {
  .uci-grid {
    grid-template-columns: 1fr;
  }
  .uci-search-input {
    width: 100%;
  }
}
</style>

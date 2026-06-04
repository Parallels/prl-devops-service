/**
 * uce-engine.js — Render Engine for the Use Case Engine
 *
 * Parses step definitions, manages step sequencing, renders narrative +
 * section_header steps, handles prev/next navigation, and drives the progress bar.
 * Depends on: uce-state.js
 *
 * Usage: Loaded via script tag after DOM is ready. Reads data from
 *   window.__UCE_DATA__ (set by Jekyll layout) and renders into #uce-flow.
 */
(function () {
  'use strict';

  /* ── Constants ──────────────────────────────────────────────────── */
  var STEPS_CONTAINER_ID = 'uce-flow-container';

  /* ── State ──────────────────────────────────────────────────────── */
  var useCaseId = '';
  var steps = [];
  var sideQuests = [];
  var whatsNext = null;
  var visibleSteps = [];
  var state = null;
  var currentIndex = 0;

  /* ── DOM refs (lazily cached) ──────────────────────────────────── */
  var dom = {};

  function cacheDom() {
    dom.container = document.getElementById(STEPS_CONTAINER_ID);
    dom.containerParent = dom.container ? dom.container.parentElement : null;
    dom.progressFill = document.getElementById('uce-progress-fill');
    dom.topbarCount = document.getElementById('uce-topbar-count');
    dom.topbarProgress = document.getElementById('uce-topbar-progress');
  }

  /* ── Initialization ─────────────────────────────────────────────── */

  function init() {
    // Read data from data attribute (avoids jsonify issues with multi-line strings)
    var dataEl = document.getElementById('uce-flow-container');
    var rawData = dataEl ? dataEl.getAttribute('data-uce-data') : null;
    if (!rawData) {
      console.error('[UCE] No use case data found. Ensure the layout sets data-uce-data.');
      return;
    }

    var data;
    try {
      data = JSON.parse(rawData);
    } catch (e) {
      console.error('[UCE] Failed to parse use case data:', e);
      return;
    }

    useCaseId = data.id || '';
    steps = data.steps || [];
    sideQuests = data.side_quests || [];
    whatsNext = data.whats_next || null;

    if (!useCaseId || steps.length === 0) {
      console.warn('[UCE] Use case has no ID or steps.');
      return;
    }

    cacheDom();

    // Hide loading indicator
    if (dom.loading) {
      dom.loading.style.display = 'none';
    }

    // Resolve localStorage state
    state = UCEState.getState(useCaseId);

    // Filter to visible steps based on branch choices
    visibleSteps = UCEState.getVisibleSteps(useCaseId, steps);

    // Determine current index (last completed step, or first step)
    currentIndex = findCurrentIndex();

    // Set initial active_step and current_step to match current position
    var initState = UCEState.getState(useCaseId);
    if (!initState.active_step) {
      // When use case is complete, leave active_step null so updateStepStates doesn't mark anything as --current
      if (initState.current_step !== '__complete__') {
        initState.active_step = visibleSteps[currentIndex].id;
      }
    }
    if (!initState.current_step) {
      initState.current_step = visibleSteps[currentIndex].id;
    }
    initState.updated_at = new Date().toISOString();
    UCEState.save();

    // Render the flow
    renderFlow();
    updateStepStates();
    updateProgress();

    // Scroll to the current step so the user lands on the right position after reload
    scrollToStep(currentIndex);

    /* ── Resolve expressions on already-completed steps (reload scenario) ── */
    if (window.ExpressionResolver) {
      window.ExpressionResolver.refreshAllSteps();
    }

    /* ── Preload whats_next recommendation data ── */
    preloadWhatsNextData(dom.container);

    /* ── Render whats_next section if use case is already complete ── */
    if (whatsNext && state.current_step === '__complete__') {
      renderWhatsNextSection(whatsNext);
      // Scroll to the whats_next section so the user sees the completion card
      setTimeout(function() {
        var wnSection = document.getElementById('uce-whats-next');
        if (wnSection) {
          wnSection.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }, 200);
    }

    // Bind navigation events
    bindNavigation();
    bindPanelEvents();

    // Check for return-to-origin (side quest completion redirect)
    checkReturnToOrigin();

    // Update breadcrumb with parent link from side quest navigation
    updateBreadcrumbFromSession();

    // Emit ready event
    emitEvent('engine-ready', { useCaseId: useCaseId, totalSteps: visibleSteps.length });

    // Hide loading spinner
    var loadingEl = document.getElementById('uce-loading');
    if (loadingEl) loadingEl.style.display = 'none';
  }

  /**
   * Find the index of the current step.
   * Priority: current_step (where user left off) > last completed > first step.
   * @param {Array} stepsArr — optional step array (defaults to visibleSteps)
   */
  function findCurrentIndex(stepsArr) {
    var arr = stepsArr || visibleSteps;

    // Priority 1: saved active_step (where the user last was viewing)
    // This ensures Prev/Next navigate relative to where the user left off,
    // not relative to the current_step (completion marker) which may be ahead.
    var initState = UCEState.getState(useCaseId);
    if (initState.active_step) {
      for (var ai = 0; ai < arr.length; ai++) {
        if (arr[ai].id === initState.active_step) {
          return ai;
        }
      }
    }

    // Priority 2: saved current_step that is NOT completed
    if (initState.current_step && initState.current_step !== '__complete__') {
      for (var ci = 0; ci < arr.length; ci++) {
        if (arr[ci].id === initState.current_step && !UCEState.isStepCompleted(useCaseId, arr[ci].id)) {
          return ci;
        }
      }
    }

    // Fallback: last completed step
    for (var i = arr.length - 1; i >= 0; i--) {
      if (UCEState.isStepCompleted(useCaseId, arr[i].id)) {
        return i;
      }
    }
    return 0;
  }

  /* ── Rendering ──────────────────────────────────────────────────── */

  function renderFlow() {
    if (!dom.container) return;

    // Clear existing content (keep data attributes)
    dom.container.innerHTML = '';

    for (var i = 0; i < visibleSteps.length; i++) {
      var step = visibleSteps[i];
      var el = renderStep(step, i);
      if (el) {
        dom.container.appendChild(el);
      }
    }
  }

  /**
   * Render the whats_next completion section and append after the flow container.
   * Only called when the use case is fully complete.
   */
  function renderWhatsNextSection(wn) {
    if (!wn || !dom.container) return;

    // Remove existing section if re-rendering
    var existing = document.getElementById('uce-whats-next');
    if (existing) existing.remove();

    var section = document.createElement('div');
    section.id = 'uce-whats-next';
    section.className = 'uce-whatsnext-section';

    /* ── Congratulations header card ─────────────────────────────── */
    var headerCard = document.createElement('div');
    headerCard.className = 'uce-whatsnext-header';

    // Resolve body — supports markdown_body (parsed via marked) or plain body
    var resolvedBody = resolveStepBody(wn);

    // var partyHornSvg = '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M5.5713 14.5L9.46583 18.4141M18.9996 3.60975C17.4044 3.59505 16.6658 4.33233 16.4236 5.07743C16.2103 5.73354 16.4052 7.07735 15.896 8.0727C15.4091 9.02443 14.1204 9.5617 12.6571 9.60697M20 7.6104L20.01 7.61049M19 15.96L19.01 15.9601M7.00001 3.94926L7.01001 3.94936M19 11.1094C17.5 11.1094 16.5 11.6094 15.5949 12.5447M10.2377 7.18796C11 6.10991 11.5 5.10991 11.0082 3.52734M3.53577 20.4645L7.0713 9.85791L14.1424 16.929L3.53577 20.4645Z" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"></path> </g></svg>';
    var partyHornSvg = '<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><g clip-path="url(#clip0_1214_4417)"><path fill-rule="evenodd" clip-rule="evenodd" d="M13 11C14.9509 11 16.2542 10.3277 17.1047 9.19375C17.8601 8.18654 18.1833 6.89179 18.4476 5.83292L18.4701 5.74254C18.7639 4.5674 19.0067 3.65769 19.4953 3.00625C19.9105 2.45271 20.5759 2 22 2C22.5523 2 23 1.55228 23 1C23 0.447715 22.5523 0 22 0C20.0491 0 18.7458 0.672288 17.8953 1.80625C17.1399 2.81346 16.8167 4.10821 16.5524 5.16708L16.5299 5.25746C16.2361 6.4326 15.9933 7.34231 15.5047 7.99375C15.0895 8.54729 14.4241 9 13 9C12.4477 9 12 9.44771 12 10C12 10.5523 12.4477 11 13 11ZM9.60168 0.201244C9.15997 -0.130281 8.53314 -0.0409558 8.20161 0.400757C7.87009 0.842469 7.95941 1.4693 8.40112 1.80083C8.82124 2.11615 9.17514 2.51113 9.44262 2.96321C9.71009 3.4153 9.8859 3.91565 9.96 4.43568C10.0341 4.95572 10.0051 5.48526 9.87452 5.99406C9.74397 6.50287 9.51449 6.98098 9.19917 7.4011C8.86765 7.84282 8.95697 8.46965 9.39868 8.80118C9.8404 9.1327 10.4672 9.04338 10.7988 8.60166C11.2717 7.97148 11.616 7.25431 11.8118 6.4911C12.0076 5.72789 12.0512 4.93358 11.94 4.15353C11.8288 3.37348 11.5651 2.62296 11.1639 1.94483C10.7627 1.26669 10.2319 0.674222 9.60168 0.201244ZM7 5C7 5.55228 6.55228 6 6 6C5.44772 6 5 5.55228 5 5C5 4.44772 5.44772 4 6 4C6.55228 4 7 4.44772 7 5ZM22 9C22.5523 9 23 8.55228 23 8C23 7.44772 22.5523 7 22 7C21.4477 7 21 7.44772 21 8C21 8.55228 21.4477 9 22 9ZM20 18C20 18.5523 19.5523 19 19 19C18.4477 19 18 18.5523 18 18C18 17.4477 18.4477 17 19 17C19.5523 17 20 17.4477 20 18ZM19.5663 14.0403C19.0464 13.9659 18.5168 13.9947 18.0079 14.125C17.4991 14.2553 17.0208 14.4845 16.6005 14.7996C16.1587 15.1309 15.5319 15.0413 15.2006 14.5994C14.8693 14.1575 14.9589 13.5307 15.4008 13.1994C16.0312 12.7268 16.7486 12.3829 17.5119 12.1875C18.2752 11.992 19.0695 11.9489 19.8495 12.0604C20.6295 12.172 21.3799 12.4361 22.0578 12.8377C22.7358 13.2392 23.3279 13.7704 23.8006 14.4008C24.1319 14.8427 24.0423 15.4695 23.6004 15.8008C23.1585 16.1321 22.5317 16.0424 22.2004 15.6005C21.8853 15.1802 21.4905 14.8261 21.0386 14.5584C20.5866 14.2907 20.0863 14.1147 19.5663 14.0403ZM6.70714 9.29291C6.46777 9.05355 6.12357 8.95153 5.79244 9.0218C5.4613 9.09207 5.1882 9.32509 5.06668 9.64104L3.67459 13.2605L10.7396 20.3255L14.359 18.9334C14.675 18.8118 14.908 18.5387 14.9782 18.2076C15.0485 17.8765 14.9465 17.5323 14.7071 17.2929L6.70714 9.29291ZM8.6968 21.1111L2.88891 15.3032L0.0666833 22.641C-0.0751812 23.0099 0.0134815 23.4277 0.292922 23.7071C0.572363 23.9866 0.99016 24.0752 1.35901 23.9334L8.6968 21.1111Z" fill="currentColor"></path></g></svg>';

    headerCard.innerHTML =
      '<div class="uce-whatsnext-icon">' + partyHornSvg + '</div>' +
      '<div class="uce-whatsnext-content">' +
        '<h3 class="uce-whatsnext-title">Congratulations!</h3>' +
        '<p class="uce-whatsnext-message">' + resolvedBody + '</p>' +
      '</div>';
    section.appendChild(headerCard);

    /* ── Recommendation grid ─────────────────────────────────────── */
    if (!wn.recommendations || wn.recommendations.length === 0) {
      dom.container.parentNode.insertBefore(section, dom.container.nextSibling);
      return;
    }

    var heading = document.createElement('h4');
    heading.className = 'uce-whatsnext-heading';
    heading.textContent = "What's Next";
    section.appendChild(heading);

    var grid = document.createElement('div');
    grid.className = 'uce-whatsnext-grid';

    var visibleCount = 0;
    for (var i = 0; i < wn.recommendations.length; i++) {
      var rec = wn.recommendations[i];
      if (!rec.use_case) continue;

      /* Skip hidden use cases */
      if (window.__UCE_HIDDEN_MAP__ && window.__UCE_HIDDEN_MAP__[rec.use_case]) {
        continue;
      }

      visibleCount++;

      var card = document.createElement('a');
      card.className = 'uce-whatsnext-card';
      card.href = (window.__UCE_BASEURL__) + '/use-cases/' + rec.use_case + '/';
      card.setAttribute('rel', 'noopener');

      var iconHtml = '';
      if (rec.icon) {
        iconHtml = '<i class="fa-solid ' + escapeHtml(rec.icon) + '"></i>';
      }

      card.innerHTML =
        '<div class="uce-whatsnext-card-head">' +
          (iconHtml ? '<div class="uce-whatsnext-card-icon">' + iconHtml + '</div>' : '') +
          '<span class="uce-whatsnext-card-title uce-skeleton"></span>' +
        '</div>' +
        '<p class="uce-whatsnext-card-desc"><span class="uce-whatsnext-card-text uce-skeleton"></span></p>';

      grid.appendChild(card);
    }

    /* Hide the whole section if all recommendations are hidden */
    if (visibleCount === 0) {
      return;
    }

    section.appendChild(grid);
    dom.container.parentNode.insertBefore(section, dom.container.nextSibling);

    /* ── Populate recommendation data ── */
    preloadWhatsNextData(section);

    /* ── Resolve any [[expressions]] in the whats_next section ── */
    if (window.ExpressionResolver) {
      window.ExpressionResolver.resolveNode(section);
    }
  }

  /**
   * Render a single step element.
   */
  function renderStep(step, index) {
    var el = document.createElement('div');
    el.className = 'uce-flow-step';
    el.setAttribute('data-uce-step', step.id || '');
    el.setAttribute('data-uce-step-index', String(index));
    el.setAttribute('data-uce-kind', step.kind || 'narrative');

    // Status class — a step is current (first incomplete) or completed (passed via Next)
    var isCompleted = UCEState.isStepCompleted(useCaseId, step.id);
    var isCurrent = (index === currentIndex);

    // If the last step is completed, treat it as completed — not current
    var isLastStep = (index === visibleSteps.length - 1);
    if (isLastStep && isCompleted) {
      isCurrent = false;
      isCompleted = true;
    }

    // Current step takes precedence — never mark it completed
    if (isCurrent) {
      el.classList.add('uce-flow-step--current');
    } else if (isCompleted) {
      el.classList.add('uce-flow-step--completed');
    }

    // Step header (number + title)
    var header = document.createElement('div');
    header.className = 'uce-step-header';

    var numberWrapper = document.createElement('div');
    numberWrapper.className = 'uce-step-number-wrap';

    var numberEl = document.createElement('span');
    numberEl.className = 'uce-step-number';
    if (isCompleted) {
      numberEl.innerHTML = '<i class="fa-solid fa-check"></i>';
    } else if (isCurrent && step.icon) {
      numberEl.className += ' uce-step-number--icon';
      numberEl.innerHTML = '<i class="fa-solid ' + step.icon + '"></i>';
    } else {
      numberEl.textContent = String(index + 1);
    }

    numberWrapper.appendChild(numberEl);

    var titleEl = document.createElement('span');
    titleEl.className = 'uce-step-title';
    titleEl.textContent = step.title || '';

    header.appendChild(numberWrapper);
    header.appendChild(titleEl);

    // Duration badge
    if (step.duration) {
      var durBadge = document.createElement('span');
      durBadge.className = 'uce-step-duration';
      durBadge.textContent = step.duration;
      header.appendChild(durBadge);
    }

    // Current badge placeholder — persistent element that updateStepStates fills/empties
    var currentBadgePlaceholder = document.createElement('span');
    currentBadgePlaceholder.className = 'uce-current-placeholder';
    currentBadgePlaceholder.style.display = 'none';
    header.appendChild(currentBadgePlaceholder);

    // Chevron toggle — only for completed steps that are NOT current
    if (isCompleted && !isCurrent) {
      var chevronBtn = document.createElement('button');
      chevronBtn.className = 'uce-step-toggle';
      chevronBtn.setAttribute('aria-label', 'Toggle step content');
      chevronBtn.innerHTML = '<svg class="uce-step-chevron" viewBox="0 0 20 20" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 7l5 5 5-5"/></svg>';
      header.appendChild(chevronBtn);
    }

    // Nav buttons — prev hidden on first step
    var navRow = document.createElement('div');
    navRow.className = 'uce-step-nav';

    var stepId = step.id || '';

    if (index > 0) {
      var prevBtn = document.createElement('button');
      prevBtn.className = 'uce-btn uce-btn--outline uce-step-nav-btn';
      prevBtn.setAttribute('data-uce-action', 'prev');
      prevBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg> Prev';
      navRow.appendChild(prevBtn);
    }

    var nextBtn = document.createElement('button');
    nextBtn.className = 'uce-btn uce-btn--primary uce-step-nav-btn';
    nextBtn.setAttribute('data-uce-action', 'next');
    var isLastStep = (index === visibleSteps.length - 1);
    var allCompleted = visibleSteps.every(function(s) { return UCEState.isStepCompleted(useCaseId, s.id); });
    if (allCompleted) {
      // Hide nav row entirely — use case is complete
      navRow.style.display = 'none';
    } else {
      nextBtn.innerHTML = (isLastStep ? 'Finish' : 'Next') + ' <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>';
    }

    // Lock Next button on interactive steps until user acts
    var interactiveKinds = ['choice', 'quiz'];
    if (interactiveKinds.indexOf(step.kind) !== -1) {
      var state = UCEState.getState(useCaseId);
      var isDone = false;
      if (step.kind === 'choice' && state.branches_chosen[stepId]) {
        isDone = true;
      }
      if (step.kind === 'quiz' && state.quiz_answers[stepId]) {
        isDone = true;
      }
      if (!isDone) {
        nextBtn.classList.add('uce-step-nav-btn--locked');
        nextBtn.setAttribute('data-uce-locked', 'true');
      }
    }

    navRow.appendChild(nextBtn);
    header.appendChild(navRow);

    el.appendChild(header);

    // Body based on step kind
    var bodyEl = null;

    switch (step.kind) {
      case 'section_header':
        bodyEl = renderSectionHeader(step);
        break;
      case 'narrative':
        bodyEl = renderNarrative(step);
        break;
      case 'choice':
        bodyEl = renderChoiceStep(step);
        break;
      case 'checkpoint':
        bodyEl = renderCheckpointStep(step);
        break;
      case 'quiz':
        bodyEl = renderQuizStep(step);
        break;
      case 'code_editor':
        bodyEl = renderCodeEditorStep(step);
        break;
      case 'markdown':
        bodyEl = renderMarkdownStep(step);
        break;
      default:
        bodyEl = renderNarrative(step);
    }

    if (bodyEl) {
      // Resources section — always rendered as the last child inside the body,
      // so it hides when the step collapses.
      var resEl = renderResources(step.resources);
      if (resEl) {
        bodyEl.appendChild(resEl);
      }
      el.appendChild(bodyEl);
    }

    /* ── Resolve any [[expressions]] in the rendered step ─────────── */
    if (window.ExpressionResolver) {
      window.ExpressionResolver.resolveNode(el);
    }

    return el;
  }

  /** Render a section header step. */
  function renderSectionHeader(step) {
    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--section';

    var divider = document.createElement('hr');
    divider.className = 'uce-divider';

    var label = document.createElement('div');
    label.className = 'uce-section-label';
    if (step.icon) {
      label.setAttribute('data-uce-icon', step.icon);
    }
    label.textContent = step.title || '';

    div.appendChild(divider);
    div.appendChild(label);
    return div;
  }

  /** Render a markdown step — body is raw Markdown parsed via marked.js. */
  function renderMarkdownStep(step) {
    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--markdown';

    var html = '';
    // markdown_body takes precedence — it's raw Markdown to be parsed
    var mdSource = step.markdown_body || step.body || '';
    if (mdSource) {
      var _marked = window.marked || (typeof globalThis !== 'undefined' && globalThis.marked);
      if (_marked && typeof _marked.parse === 'function') {
        try {
          html = _marked.parse(mdSource);
        } catch (e) {
          console.warn('[UCE] marked.parse failed for step ' + step.id + ':', e);
          html = '<p>' + escapeHtml(mdSource) + '</p>';
        }
      } else {
        console.warn('[UCE] marked.js not available — rendering markdown step body as plain text');
        html = '<p>' + escapeHtml(mdSource) + '</p>';
      }
    }

    div.innerHTML = html;
    return div;
  }

  /**
   * Resolve step body content for narrative rendering.
   * If markdown_body is present, parse it via marked.js and return HTML.
   * Otherwise return body as-is (already HTML).
   * If both are present, only use markdown_body.
   * @param {Object} obj — step or panel config with markdown_body/body
   * @returns {string} HTML string to render
   */
  function resolveStepBody(obj) {
    if (obj.markdown_body) {
      var _marked = window.marked || (typeof globalThis !== 'undefined' && globalThis.marked);
      if (_marked && typeof _marked.parse === 'function') {
        try {
          return _marked.parse(obj.markdown_body);
        } catch (e) {
          console.warn('[UCE] marked.parse failed for markdown_body in ' + (obj.id || 'step') + ':', e);
          return '<p>' + escapeHtml(obj.markdown_body) + '</p>';
        }
      } else {
        console.warn('[UCE] marked.js not available — rendering markdown_body as plain text');
        return '<p>' + escapeHtml(obj.markdown_body) + '</p>';
      }
    }
    return obj.body || '';
  }

  /* ── Panel Event Delegation (Phase 3-01) ───────────────────────── */

  function bindPanelEvents() {
    if (!dom.container) return;

    // Toggle expand/collapse for completed steps
    dom.container.addEventListener('click', function(e) {
      var toggleBtn = e.target.closest('.uce-step-toggle');
      if (toggleBtn) {
        e.preventDefault();
        e.stopPropagation();
        var stepEl = toggleBtn.closest('.uce-flow-step');
        // Only allow toggle on completed steps, not the active one
        if (!stepEl || !stepEl.classList.contains('uce-flow-step--completed')) return;
        var chevron = stepEl.querySelector('.uce-step-chevron');
        var bodyEl = stepEl.querySelector('.uce-step-body');
        if (stepEl && bodyEl) {
          var isExpanded = stepEl.classList.contains('uce-flow-step--expanded');
          stepEl.classList.toggle('uce-flow-step--expanded');
          bodyEl.style.display = isExpanded ? 'none' : 'block';
          if (chevron) {
            chevron.style.transform = isExpanded ? '' : 'rotate(180deg)';
          }
        }
        return;
      }

      // Copy-to-clipboard for terminal and api_call panels
      var copyBtn = e.target.closest('.mac-window-copy');
      if (copyBtn) {
        e.preventDefault();
        var panelEl = copyBtn.closest('.uce-panel-terminal, .uce-panel-api');
        if (panelEl) {
          var codeEl = panelEl.querySelector('.mac-window-body code');
          if (codeEl) {
            var text = codeEl.textContent;
            if (navigator.clipboard && navigator.clipboard.writeText) {
              navigator.clipboard.writeText(text).then(function() {
                showCopyTooltip(copyBtn);
              }).catch(function() {
                fallbackCopy(text, copyBtn);
              });
            } else {
              fallbackCopy(text, copyBtn);
            }
          }
        }
      }

      // Reset branch choice
      var resetBtn = e.target.closest('.uce-branch-reset');
      if (resetBtn) {
        // Block resets when use case is complete
        if (state.current_step === '__complete__') return;
        e.preventDefault();
        e.stopPropagation();
        var choiceStepId = resetBtn.getAttribute('data-uce-step');
        if (!choiceStepId) return;
        // Remove the choice
        var st = UCEState.getState(useCaseId);
        delete st.branches_chosen[choiceStepId];
        st.updated_at = new Date().toISOString();
        UCEState.save();
        // Re-render flow with updated branching
        visibleSteps = UCEState.getVisibleSteps(useCaseId, steps);
        currentIndex = findCurrentIndex();
        renderFlow();
        updateStepStates();
        updateProgress();
        emitEvent('branch-reset', { stepId: choiceStepId });
        return;
      }
    });
  }

  function showCopyTooltip(btn) {
    var orig = btn.innerHTML;
    btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="#22C55E" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg> Copied!';
    btn.style.color = '#22C55E';
    setTimeout(function() {
      btn.innerHTML = orig;
      btn.style.color = '';
    }, 2000);
  }

  function fallbackCopy(text, btn) {
    var ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.left = '-9999px';
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand('copy'); showCopyTooltip(btn); } catch(err) { /* noop */ }
    document.body.removeChild(ta);
  }

  /** Render a narrative step. */
  function renderNarrative(step) {
    var div = document.createElement('div');
    div.className = 'uce-step-body';

    // Layout mode: 'columns' (side-by-side) | 'rows' (stacked) | 'auto' (infer from presence of panel)
    var layout = step.layout || 'auto';

    /* ── Panel resolution: panels (plural) wins over panel (singular) ── */
    var panelsConfig = step.panels;
    var singlePanelConfig = step.panel;

    var visiblePanels = [];

    if (panelsConfig && Array.isArray(panelsConfig) && panelsConfig.length > 0) {
      /* ── plural panels: filter by if-condition ── */
      /* Build a full stepMap from module-level steps so _resolveRef
         can look up referenced steps (choices, quizzes, etc.) */
      var stepMap = {};
      for (var si = 0; si < steps.length; si++) {
        stepMap[steps[si].id] = steps[si];
      }
      for (var pi = 0; pi < panelsConfig.length; pi++) {
        var pCfg = panelsConfig[pi];
        var panelVisible = true;
        if (pCfg.if) {
          panelVisible = UCEState._evaluateCondition(pCfg.if, stepMap, UCEState.getState(useCaseId));
        }
        if (panelVisible) {
          visiblePanels.push(pCfg);
        }
      }
    } else if (singlePanelConfig) {
      /* ── singular panel (backward compat) ── */
      visiblePanels.push(singlePanelConfig);
    }

    var hasPanel = visiblePanels.length > 0;
    var stepHasBody = !!(step.body || step.markdown_body);

    // Resolve effective layout
    var effectiveLayout = layout;
    if (layout === 'auto') {
      effectiveLayout = hasPanel ? 'columns' : 'rows';
    }

    // Grid wrapper
    var gridCls = 'uce-narrative-grid';
    if (effectiveLayout === 'rows') {
      gridCls += ' uce-narrative-grid--rows';
    } else if (effectiveLayout === 'columns' && !hasPanel) {
      gridCls += ' uce-narrative-grid--single';
    }
    var grid = document.createElement('div');
    grid.className = gridCls;

    /* ── Decide: single panel (normal) vs multiple panels (tabbed) ── */
    var isTabbed = visiblePanels.length > 1;

    /* ── Single-panel path: left column + right column (existing behavior) ── */
    if (!isTabbed) {
      var pCfg = visiblePanels.length > 0 ? visiblePanels[0] : null;

      var bodyHtml = resolveStepBody(step);
      // Panel-level markdown_body takes precedence over step-level bodyHtml
      var pBody = pCfg ? (resolveStepBody(pCfg) || bodyHtml || '') : (bodyHtml || '');
      var pSqIds = (pCfg && pCfg.side_quests && pCfg.side_quests.length > 0)
        ? pCfg.side_quests : (step.side_quests || []);
      var hasLeft = !!pBody || pSqIds.length > 0;

      if (hasLeft) {
        var leftCol = document.createElement('div');
        leftCol.className = 'uce-narrative-left';

        if (pBody) {
          var bodyEl = document.createElement('div');
          bodyEl.className = 'uce-narrative-body';
          if (pCfg && pCfg.kind === 'api_call') {
            var panelTitle = pCfg.title || step.title || '';
            var stepTag = step.tag || '';
            var tagHtml = stepTag
              ? '<span class="uce-step-tag">' + escapeHtml(stepTag) + '</span>'
              : '';
            bodyEl.innerHTML =
              '<div class="uce-api-call-header">' +
              tagHtml +
              '<h3 class="uce-api-call-title">' + escapeHtml(panelTitle) + '</h3>' +
              '</div>' +
              pBody;
          } else {
            bodyEl.innerHTML = pBody;
          }
          leftCol.appendChild(bodyEl);
        }

        if (pSqIds.length > 0) {
          var sqTriggers = renderSideQuestTriggers(pSqIds, step.id);
          for (var k = 0; k < sqTriggers.length; k++) {
            leftCol.appendChild(sqTriggers[k]);
          }
        }

        grid.appendChild(leftCol);
      }

      /* ── Right column — panel (if present) ── */
      if (hasPanel) {
        var panelEl = renderPanel(pCfg);
        if (panelEl) {
          var rightCol = document.createElement('div');
          rightCol.className = 'uce-narrative-right';

          if (effectiveLayout === 'columns') {
            var expandBtn = document.createElement('button');
            expandBtn.className = 'uce-narrative-expand';
            expandBtn.setAttribute('aria-label', 'Expand panel');
            expandBtn.innerHTML = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>';
            expandBtn.addEventListener('click', function () {
              var isExpanded = grid.classList.toggle('uce-narrative-grid--expanded');
              expandBtn.classList.toggle('uce-narrative-expand--active', isExpanded);
              expandBtn.setAttribute('aria-label', isExpanded ? 'Collapse panel' : 'Expand panel');
            });
            rightCol.appendChild(expandBtn);
          }

          rightCol.appendChild(panelEl);
          grid.appendChild(rightCol);
        }
      }
    }

    /* ── Tabbed-panel path: one left column + tab bar + one right column ── */
    if (isTabbed) {
      renderNarrativeTabbed(grid, step, visiblePanels, effectiveLayout);
    }

    /* ── Step-level checklist — always at the bottom ── */
    if (step.checklist && step.checklist.length > 0) {
      var checklistUl = document.createElement('ul');
      checklistUl.className = 'uce-checklist';
      for (var i = 0; i < step.checklist.length; i++) {
        var li = document.createElement('li');
        li.className = 'uce-checklist-item';
        var md = '';
        if (typeof marked !== 'undefined') {
          try {
            md = marked.parse(step.checklist[i]).replace(/^<p>|<\/p>\n?$/g, '').trim();
          } catch(e) { /* fall through to plain text */ }
        }
        if (!md) md = step.checklist[i];
        li.innerHTML = '<i class="fa-solid fa-circle-check uce-check-icon"></i>' + md;
        checklistUl.appendChild(li);
      }
      grid.appendChild(checklistUl);
    }

    div.appendChild(grid);

    return div;
  }

  /* ── Tab helpers ──────────────────────────────────────────── */

  /** Get the display label for a panel (name > id > kind). */
  function getPanelLabel(pCfg) {
    if (pCfg.name) return pCfg.name;
    if (pCfg.id) return pCfg.id;
    return pCfg.kind || '';
  }

  /** Render the tabbed-panel narrative (multiple visible panels). */
  function renderNarrativeTabbed(grid, step, visiblePanels, effectiveLayout) {
    /* ── Left column: step-level body + side_quests ── */
    var stepBodyHtml = resolveStepBody(step);
    var hasLeft = !!stepBodyHtml || (step.side_quests && step.side_quests.length > 0);
    if (hasLeft) {
      var leftCol = document.createElement('div');
      leftCol.className = 'uce-narrative-left';

      if (stepBodyHtml) {
        var bodyEl = document.createElement('div');
        bodyEl.className = 'uce-narrative-body';
        bodyEl.innerHTML = stepBodyHtml;
        leftCol.appendChild(bodyEl);
      }

      if (step.side_quests && step.side_quests.length > 0) {
        var sqTriggers = renderSideQuestTriggers(step.side_quests, step.id);
        for (var k = 0; k < sqTriggers.length; k++) {
          leftCol.appendChild(sqTriggers[k]);
        }
      }

      grid.appendChild(leftCol);
    }

    /* ── Right column: tab bar + active panel container ── */
    var rightCol = document.createElement('div');
    rightCol.className = 'uce-narrative-right';

    /* Tab bar wrapper — flex-grow pushes expand button to the right */
    var tabBarWrap = document.createElement('div');
    tabBarWrap.className = 'uce-panel-tabs-wrapper';

    /* Tab bar */
    var tabBar = document.createElement('div');
    tabBar.className = 'uce-panel-tabs';

    for (var ti = 0; ti < visiblePanels.length; ti++) {
      (function (idx) {
        var tabBtn = document.createElement('button');
        tabBtn.className = 'uce-panel-tab' + (idx === 0 ? ' uce-panel-tab--active' : '');
        tabBtn.setAttribute('data-panel-index', idx);
        tabBtn.textContent = getPanelLabel(visiblePanels[idx]);
        tabBtn.addEventListener('click', function () {
          setActiveTab(tabBar, visiblePanels[idx], idx);
        });
        tabBar.appendChild(tabBtn);
      })(ti);
    }

    tabBarWrap.appendChild(tabBar);
    rightCol.appendChild(tabBarWrap);

    /* Active panel container */
    var activePanelWrap = document.createElement('div');
    activePanelWrap.className = 'uce-panel-active';
    rightCol.appendChild(activePanelWrap);

    grid.appendChild(rightCol);

    /* ── Expand toggle inside tab bar wrapper ── */
    if (effectiveLayout === 'columns') {
      var expandBtn = document.createElement('button');
      expandBtn.className = 'uce-narrative-expand';
      expandBtn.setAttribute('aria-label', 'Expand panel');
      expandBtn.innerHTML = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>';
      expandBtn.addEventListener('click', function () {
        var isExpanded = grid.classList.toggle('uce-narrative-grid--expanded');
        expandBtn.classList.toggle('uce-narrative-expand--active', isExpanded);
        expandBtn.setAttribute('aria-label', isExpanded ? 'Collapse panel' : 'Expand panel');
      });
      tabBarWrap.appendChild(expandBtn);
    }

    /* Render initial tab */
    if (visiblePanels.length > 0) {
      setActiveTab(tabBar, visiblePanels[0], 0);
    }
  }

  /** Activate a tab and render its panel inside the active container. */
  function setActiveTab(tabBar, pCfg, index) {
    var tabs = tabBar.querySelectorAll('.uce-panel-tab');
    for (var t = 0; t < tabs.length; t++) {
      tabs[t].classList.remove('uce-panel-tab--active');
    }
    tabs[index].classList.add('uce-panel-tab--active');

    /* Find the active panel container — sibling of the tab bar wrapper */
    var wrap = tabBar.parentElement.parentElement.querySelector('.uce-panel-active');
    if (!wrap) return;
    wrap.innerHTML = '';

    var panelEl = renderPanel(pCfg);
    if (panelEl) {
      wrap.appendChild(panelEl);
    }
  }

  /** Generate a curl command from api_call params. */
  function generateCurlCommand(cfg) {
    var parts = ['curl'];
    if (cfg.method) parts.push('-X', cfg.method.toUpperCase());
    if (cfg.url) parts.push('"' + cfg.url + '"');
    if (cfg.headers) {
      var h = cfg.headers;
      if (typeof h === 'object') {
        for (var key in h) {
          if (Object.prototype.hasOwnProperty.call(h, key)) {
            parts.push('-H', '"' + key + ': ' + h[key] + '"');
          }
        }
      } else if (typeof h === 'string') {
        parts.push('-H', '"' + h + '"');
      }
    }
    if (cfg.body) parts.push('-d', '"' + cfg.body.replace(/"/g, '\\"') + '"');
    return parts.join(' ');
  }

  /** Syntax-highlight a JSON string. */
  function syntaxHighlightJSON(json) {
    if (!json) return '';
    // Tokenize JSON
    return json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/("(?:[^"\\]|\\.)*")\s*:/g, '<span class="json-key">$1</span>:')
      .replace(/:\s*("(?:[^"\\]|\\.)*")/g, ': <span class="json-string">$1</span>')
      .replace(/:\s*(\d+\.?\d*)/g, ': <span class="json-number">$1</span>')
      .replace(/:\s*(true|false)/gi, ': <span class="json-bool">$1</span>')
      .replace(/:\s*(null)/gi, ': <span class="json-null">$1</span>');
  }

  /** Render a panel element. */
  function renderPanel(panel) {
    var div = document.createElement('div');
    div.className = 'uce-panel';
    div.setAttribute('data-uce-panel', panel.kind || '');

    switch (panel.kind) {
      case 'terminal':
        div.classList.add('uce-panel-terminal');
        var termCmd = panel.cmd || '';
        var termOutput = panel.output || '';
        var cmdLang = panel.cmd_language || 'bash';
        var outLang = panel.output_language || '';
        var showCopy = panel.copy_button !== false;
        // Syntax-highlight cmd (default: bash) and output (default: plain)
        var termCmdHtml = (function() {
          if (typeof Prism !== 'undefined' && Prism.languages[cmdLang]) {
            return Prism.highlight(escapeHtml(termCmd), Prism.languages[cmdLang], 'language-' + cmdLang);
          }
          return escapeHtml(termCmd);
        })();
        var termOutputHtml = (function() {
          if (!termOutput) return '';
          if (outLang && typeof Prism !== 'undefined' && Prism.languages[outLang]) {
            return Prism.highlight(escapeHtml(termOutput), Prism.languages[outLang], 'language-' + outLang);
          }
          // Default output: plain text (no highlighting)
          return escapeHtml(termOutput);
        })();
        // Build macOS-style terminal windows
        var cmdWindow =
          '<div class="mac-window">' +
            '<div class="mac-window-header">' +
              '<span class="mac-dot mac-dot--red"></span>' +
              '<span class="mac-dot mac-dot--yellow"></span>' +
              '<span class="mac-dot mac-dot--green"></span>' +
              '<span class="mac-window-title">Terminal</span>' +
              (showCopy ? '<button class="mac-window-copy" data-uce-action="copy" aria-label="Copy command"><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>' : '') +
            '</div>' +
            '<pre class="mac-window-body"><code class="language-' + (cmdLang || 'bash') + '">' + termCmdHtml + '</code></pre>' +
          '</div>';
        var outputWindow = termOutputHtml
          ? '<div class="mac-window">' +
              '<div class="mac-window-header">' +
                '<span class="mac-dot mac-dot--red"></span>' +
                '<span class="mac-dot mac-dot--yellow"></span>' +
                '<span class="mac-dot mac-dot--green"></span>' +
                '<span class="mac-window-title">Output</span>' +
              '</div>' +
              '<pre class="mac-window-body"><code class="language-' + (outLang || 'plaintext') + '">' + termOutputHtml + '</code></pre>' +
            '</div>'
          : '';
        div.innerHTML = cmdWindow + outputWindow;
        break;
      case 'walkthrough':
        var wtKey = panel.items_ref || '';
        var twData = null;
        if (wtKey && window.__UCE_WALKTHROUGH_DATA__ && window.__UCE_WALKTHROUGH_DATA__[wtKey]) {
          twData = window.__UCE_WALKTHROUGH_DATA__[wtKey];
        }
        if (!twData) {
          // No data — render nothing
          return null;
        }
        div.classList.add('uce-panel-walkthrough');
        // Build the walkthrough HTML
        var twHtml = buildWalkthroughHtml(twData, panel);
        div.innerHTML = twHtml;
        // Initialize the walkthrough JS
        initWalkthrough(div, twData);
        return div;
      case 'api_call':
        div.classList.add('uce-panel-api');
        var cfg = {
          url: panel.url || '',
          method: panel.method || 'GET',
          headers: panel.headers || null,
          body: panel.body || null
        };
        var curlCmd = generateCurlCommand(cfg);
        // Syntax-highlight on plain text FIRST, then escape
        var curlHtml = curlCmd
          .replace(/curl/g, '[CMD]curl[/CMD]')
          .replace(/(-X)\s+(\w+)/g, '[FLAG]$1[/FLAG] [METHOD]$2[/METHOD]')
          .replace(/(-H)\s+"([^"]+)"/g, '[FLAG]$1[/FLAG] "[HEADER]$2[/HEADER]"')
          .replace(/(-d)\s+"([^"]+)"/g, '[FLAG]$1[/FLAG] "[BODY]$2[/BODY]"')
          .replace(/"([^"]+)"/g, '[URL]$1[/URL]');
        curlHtml = escapeHtml(curlHtml)
          .replace(/\[CMD\]/g, '<span class="cmd-cmd">')
          .replace(/\[\/CMD\]/g, '</span>')
          .replace(/\[FLAG\]/g, '<span class="cmd-flag">')
          .replace(/\[\/FLAG\]/g, '</span>')
          .replace(/\[METHOD\]/g, '<span class="cmd-method">')
          .replace(/\[\/METHOD\]/g, '</span>')
          .replace(/\[HEADER\]/g, '<span class="cmd-header">')
          .replace(/\[\/HEADER\]/g, '</span>')
          .replace(/\[BODY\]/g, '<span class="cmd-body">')
          .replace(/\[\/BODY\]/g, '</span>')
          .replace(/\[URL\]/g, '<span class="cmd-url">')
          .replace(/\[\/URL\]/g, '</span>');
        var responseBody = panel.expect_body || '';
        var responseHtml = '';
        if (responseBody) {
          responseHtml = syntaxHighlightJSON(responseBody);
        } else {
          responseHtml = '<span class="json-null">{}</span>';
        }
        div.innerHTML =
          '<div class="mac-window">' +
            '<div class="mac-window-header">' +
              '<span class="mac-dot mac-dot--red"></span>' +
              '<span class="mac-dot mac-dot--yellow"></span>' +
              '<span class="mac-dot mac-dot--green"></span>' +
              '<span class="mac-window-title">Terminal</span>' +
              '<button class="mac-window-copy" data-uce-action="copy" aria-label="Copy command"><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg></button>' +
            '</div>' +
            '<pre class="mac-window-body"><code>' + curlHtml + '</code></pre>' +
          '</div>' +
          '<div class="mac-window mac-window--response">' +
            '<div class="mac-window-header mac-window-header--response">' +
              '<span class="mac-dot mac-dot--red"></span>' +
              '<span class="mac-dot mac-dot--yellow"></span>' +
              '<span class="mac-dot mac-dot--green"></span>' +
              '<span class="mac-window-title mac-window-title--response">RESPONSE BODY</span>' +
            '</div>' +
            '<pre class="mac-window-body"><code>' + responseHtml + '</code></pre>' +
          '</div>';
        break;
      default:
        div.innerHTML = '<div class="uce-panel-placeholder">Unknown panel type: ' + escapeHtml(panel.kind) + '</div>';
    }

    return div;
  }

  /** Build walkthrough HTML from data. */
  function buildWalkthroughHtml(tw, panel) {
    var mainImage = tw.main_image || '';
    var uid = panel.id ? '-uid-' + panel.id : '';
    var html = '<div class="tw-container" id="tw-container' + uid + '">' +
      '<div class="tw-stage">' +
        '<div class="tw-images">' +
          '<img class="tw-img-main" src="' + window.__UCE_BASEURL__ + '/assets/img/' + escapeHtml(mainImage) + '" alt="' + escapeHtml(tw.title) + ' \u2014 default view" />' +
        '</div>' +
        '<div class="tw-hotspot-layer"></div>' +
        '<div class="tw-scene-badge" id="tw-scene-badge' + uid + '">' +
          '<span class="tw-scene-badge-label"></span>' +
          '<button class="tw-scene-badge-close" id="tw-scene-badge-close' + uid + '" aria-label="Exit scene">\u00d7</button>' +
        '</div>' +
      '</div>' +
      '<div class="tw-info-panel" id="tw-info-panel' + uid + '">' +
        '<button class="tw-info-close" id="tw-info-close' + uid + '" aria-label="Close">\u2715</button>' +
        '<div class="tw-info-content">' +
          '<span class="tw-info-step-label"></span>' +
          '<h3 class="tw-info-title"></h3>' +
          '<p class="tw-info-desc"></p>' +
          '<div class="tw-nav">' +
            '<button class="tw-nav-btn tw-nav-prev" aria-label="Previous step">\u2190 Prev</button>' +
            '<div class="tw-dots"></div>' +
            '<button class="tw-nav-btn tw-nav-next" aria-label="Next step">Next \u2192</button>' +
          '</div>' +
        '</div>' +
      '</div>' +
    '</div>';
    return html;
  }

  /** Initialize walkthrough JS for a rendered panel. */
  function initWalkthrough(panelEl, twData) {
    (function() {
      var tw = twData;
      var items = tw.items || [];
      if (!items.length) return;

      // Build UID suffix from panel.id (set by caller)
      var uid = '';
      var panelIdEl = panelEl.querySelector('[id^="tw-container-uid-"]');
      if (panelIdEl) {
        var m = panelIdEl.id.match(/-uid-(.+)$/);
        if (m) uid = m[1];
      }
      var sel = function(selStr) {
        var suffix = uid ? '-' + uid : '';
        if (selStr.charAt(0) === '#') {
          return panelEl.querySelector(selStr + suffix);
        }
        if (selStr.charAt(0) === '.') {
          return panelEl.querySelector(selStr + (suffix ? '-' + suffix : ''));
        }
        return panelEl.querySelector('#' + selStr + suffix);
      };

      var container = sel('tw-container');
      if (!container) return;

      var stage = container.querySelector('.tw-stage');
      var imgMain = container.querySelector('.tw-img-main');
      var hotspotLayer = container.querySelector('.tw-hotspot-layer');
      var infoPanel = sel('tw-info-panel');
      var infoStepLabel = sel('.tw-info-step-label');
      var infoTitle = sel('.tw-info-title');
      var infoDesc = sel('.tw-info-desc');
      var dotsContainer = sel('.tw-dots');
      var btnPrev = sel('.tw-nav-prev');
      var btnNext = sel('.tw-nav-next');
      var sceneBadge = sel('tw-scene-badge');
      var sceneBadgeLabel = sel('.tw-scene-badge-label');
      var sceneBadgeClose = sel('tw-scene-badge-close');
      var infoClose = sel('tw-info-close');

      if (tw.show_background_effect === false) {
        stage.classList.add('tw-stage--no-effect');
      }

      var stack = [];
      var sceneDepth = 0;
      var currentHotspotEl = null;
      var TRANSITION_MS = 200;

      var rootDots = buildDots(items);
      stack.push({ items: items, currentIndex: 0, dots: rootDots });

      function buildDots(itms) {
        dotsContainer.innerHTML = '';
        var dots = [];
        itms.forEach(function(item, i) {
          var color = (item.hotspot && item.hotspot.color) || '#3B82F6';
          var dot = document.createElement('button');
          dot.className = 'tw-dot' + (i === 0 ? ' tw-dot--active' : '');
          dot.setAttribute('aria-label', 'Go to item ' + (i + 1));
          dot.style.background = color;
          dot.addEventListener('click', (function(idx) {
            return function() { goToItemInScene(idx); };
          })(i));
          dotsContainer.appendChild(dot);
          dots.push(dot);
        });
        return dots;
      }

      function pushScene(sceneItem) {
        var sceneSteps = sceneItem.steps || [];
        if (!sceneSteps.length) return;
        sceneDepth++;
        var sceneDots = buildDots(sceneSteps);
        stack.push({ items: sceneSteps, currentIndex: 0, dots: sceneDots, sceneTitle: sceneItem.title });
        sceneBadge.classList.add('tw-scene-badge--visible');
        sceneBadgeLabel.textContent = sceneItem.title;
        var sceneImage = sceneItem.image || tw.main_image;
        var newSrc = window.__UCE_BASEURL__ + '/assets/img/' + sceneImage;
        var tempImg = new Image();
        tempImg.src = newSrc;
        tempImg.onload = function() {
          imgMain.style.transition = 'opacity ' + TRANSITION_MS + 'ms ease';
          imgMain.style.opacity = '0';
          setTimeout(function() { imgMain.src = newSrc; imgMain.onload = function() { imgMain.style.opacity = '1'; }; }, TRANSITION_MS);
        };
        hotspotLayer.innerHTML = '';
        renderSceneItems(sceneSteps);
        initFirstItem();
      }

      function popScene() {
        if (stack.length <= 1) return;
        stack[stack.length - 1].dots.forEach(function(d) { if (d.parentNode) d.parentNode.removeChild(d); });
        stack.pop();
        sceneDepth--;
        if (sceneDepth === 0) { sceneBadge.classList.remove('tw-scene-badge--visible'); }
        else { sceneBadgeLabel.textContent = stack[stack.length - 1].sceneTitle; }
        var parentSrc = window.__UCE_BASEURL__ + '/assets/img/' + tw.main_image;
        var tempImg = new Image();
        tempImg.src = parentSrc;
        tempImg.onload = function() {
          imgMain.style.transition = 'opacity ' + TRANSITION_MS + 'ms ease';
          imgMain.style.opacity = '0';
          setTimeout(function() { imgMain.src = parentSrc; imgMain.onload = function() { imgMain.style.opacity = '1'; }; }, TRANSITION_MS);
        };
        hotspotLayer.innerHTML = '';
        buildDots(stack[stack.length - 1].items);
        renderSceneItems(stack[stack.length - 1].items);
        initFirstItem();
      }

      function renderSceneItems(itms) {
        itms.forEach(function(item, i) {
          if (!item.hotspot) return;
          var r = item.hotspot.radius || 30;
          var color = item.hotspot.color || '#3B82F6';
          var btn = document.createElement('button');
          btn.className = 'tw-hotspot';
          btn.setAttribute('data-id', item.id);
          btn.style.left = item.hotspot.x + '%';
          btn.style.top = item.hotspot.y + '%';
          btn.style.setProperty('--tw-radius', r + 'px');
          btn.style.setProperty('--tw-color', color);
          btn.style.setProperty('--tw-color-dark', darken(color, 20));
          btn.style.setProperty('--tw-color-glow', color + '4D');
          btn.style.setProperty('--tw-color-shadow', color + '66');
          btn.style.width = (r * 2) + 'px';
          btn.style.height = (r * 2) + 'px';
          btn.setAttribute('aria-label', item.title);
          btn.innerHTML = '<span class="tw-hotspot-pulse"></span><span class="tw-hotspot-dot"></span>';
          btn.addEventListener('click', (function(idx) {
            return function() { goToItemInScene(idx); };
          })(i));
          hotspotLayer.appendChild(btn);
        });
      }

      function showCurrentItemInScene(animate) {
        if (animate === undefined) animate = true;
        var entry = stack[stack.length - 1];
        var item = entry.items[entry.currentIndex];
        var color = (item.hotspot && item.hotspot.color) || '#3B82F6';
        infoStepLabel.textContent = 'Step ' + (entry.currentIndex + 1) + ' of ' + entry.items.length;
        infoTitle.innerHTML = item.title;
        infoDesc.innerHTML = item.description;
        infoPanel.classList.add('tw-info-panel--visible');
        if (item.show_background_effect === false) { infoPanel.classList.add('tw-info-panel--no-effect'); }
        else { infoPanel.classList.remove('tw-info-panel--no-effect'); }
        entry.dots.forEach(function(d, i) {
          d.classList.toggle('tw-dot--active', i === entry.currentIndex);
          d.style.background = (entry.items[i].hotspot && entry.items[i].hotspot.color) || '#3B82F6';
        });
        var hs = hotspotLayer.querySelectorAll('.tw-hotspot');
        hs.forEach(function(h, i) {
          var hsColor = (entry.items[i].hotspot && entry.items[i].hotspot.color) || '#3B82F6';
          h.style.setProperty('--tw-color', hsColor);
          h.style.setProperty('--tw-color-dark', darken(hsColor, 20));
          h.style.setProperty('--tw-color-glow', hsColor + '4D');
          h.style.setProperty('--tw-color-shadow', hsColor + '66');
          h.classList.toggle('tw-hotspot--active', i === entry.currentIndex);
        });
        currentHotspotEl = hs[entry.currentIndex] || null;
        imgMain.style.opacity = (item.hotspot && item.hotspot.x !== undefined) ? '0.6' : '1';
        positionPanel();
      }

      function initFirstItem() {
        var entry = stack[stack.length - 1];
        var item = entry.items[entry.currentIndex];
        infoTitle.innerHTML = item.title;
        infoDesc.innerHTML = item.description;
        dotsContainer.querySelectorAll('.tw-dot').forEach(function(d) { d.classList.remove('tw-dot--active'); });
        hotspotLayer.querySelectorAll('.tw-hotspot').forEach(function(h) { h.classList.remove('tw-hotspot--active'); });
        currentHotspotEl = null;
        imgMain.style.opacity = '1';
        infoPanel.classList.remove('tw-info-panel--visible');
      }

      function positionPanel() {
        if (!currentHotspotEl || !infoPanel) return;
        var stageRect = stage.getBoundingClientRect();
        var hotspotRect = currentHotspotEl.getBoundingClientRect();
        var panelRect = infoPanel.getBoundingClientRect();
        var hx = hotspotRect.left + hotspotRect.width / 2 - stageRect.left;
        var hy = hotspotRect.top + hotspotRect.height / 2 - stageRect.top;
        var panelW = panelRect.width || 300;
        var panelH = panelRect.height || 200;
        var placement = placeBelow(hx, hy, panelW, panelH);
        if (!placement) placement = placeAbove(hx, hy, panelW, panelH);
        if (!placement) placement = placeRight(hx, hy, panelW, panelH);
        if (!placement) placement = placeLeft(hx, hy, panelW, panelH);
        infoPanel.style.left = placement.x + 'px';
        infoPanel.style.top = placement.y + 'px';
      }

      function placeBelow(x, y, pw, ph) {
        var belowY = y + 30;
        var belowX = clamp(x - pw / 2, 10, stage.offsetWidth - pw - 10);
        return (belowY + ph <= stage.offsetHeight - 10) ? { x: belowX, y: belowY } : null;
      }

      function placeAbove(x, y, pw, ph) {
        var aboveY = y - ph - 30;
        var aboveX = clamp(x - pw / 2, 10, stage.offsetWidth - pw - 10);
        return (aboveY >= 10) ? { x: aboveX, y: aboveY } : null;
      }

      function placeRight(x, y, pw, ph) {
        var rightX = x + 30;
        var rightY = clamp(y - ph / 2, 10, stage.offsetHeight - ph - 10);
        return (rightX + pw <= stage.offsetWidth - 10) ? { x: rightX, y: rightY } : null;
      }

      function placeLeft(x, y, pw, ph) {
        var leftX = x - pw - 30;
        var leftY = clamp(y - ph / 2, 10, stage.offsetHeight - ph - 10);
        return (leftX >= 10) ? { x: leftX, y: leftY } : null;
      }

      function clamp(val, min, max) { return Math.min(Math.max(val, min), max); }

      function darken(hex, percent) {
        var num = parseInt(hex.replace('#', ''), 16);
        var r = Math.max(0, (num >> 16) - Math.round(2.55 * percent));
        var g = Math.max(0, ((num >> 8) & 0x00FF) - Math.round(2.55 * percent));
        var b = Math.max(0, (num & 0x0000FF) - Math.round(2.55 * percent));
        return '#' + (0x1000000 + r * 0x10000 + g * 0x100 + b).toString(16).slice(1);
      }

      function goToItemInScene(index) {
        var entry = stack[stack.length - 1];
        if (index < 0 || index >= entry.items.length) return;
        entry.currentIndex = index;
        var item = entry.items[index];
        if (item.type === 'scene' && item.steps) { pushScene(item); return; }
        showCurrentItemInScene(true);
      }

      function goToPrevItemInScene() {
        var entry = stack[stack.length - 1];
        if (entry.currentIndex > 0) { goToItemInScene(entry.currentIndex - 1); }
        else { goToItemInScene(entry.items.length - 1); }
      }

      function goToNextItemInScene() {
        var entry = stack[stack.length - 1];
        if (entry.currentIndex < entry.items.length - 1) { goToItemInScene(entry.currentIndex + 1); }
        else { goToItemInScene(0); }
      }

      btnPrev.addEventListener('click', goToPrevItemInScene);
      btnNext.addEventListener('click', goToNextItemInScene);
      sceneBadgeClose.addEventListener('click', popScene);
      infoClose.addEventListener('click', resetSelection);

      document.addEventListener('keydown', function(e) {
        if (!container || !isInViewport(container)) return;
        if (e.key === 'ArrowLeft') btnPrev.click();
        if (e.key === 'ArrowRight') btnNext.click();
        if (e.key === 'Escape' && sceneDepth > 0) sceneBadgeClose.click();
      });

      function isInViewport(el) {
        var rect = el.getBoundingClientRect();
        return rect.bottom > 0 && rect.top < window.innerHeight;
      }

      var resizeTimer;
      window.addEventListener('resize', function() {
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(positionPanel, 100);
      });

      renderSceneItems(items);
      initFirstItem();
      positionPanel();

      stage.addEventListener('click', function(e) {
        if (e.target === stage || e.target === imgMain) { resetSelection(); }
      });

      function resetSelection() {
        dotsContainer.querySelectorAll('.tw-dot').forEach(function(d) { d.classList.remove('tw-dot--active'); });
        hotspotLayer.querySelectorAll('.tw-hotspot').forEach(function(h) { h.classList.remove('tw-hotspot--active'); });
        currentHotspotEl = null;
        imgMain.style.opacity = '1';
        infoPanel.classList.remove('tw-info-panel--visible');
      }
    })();
  }

  function renderChoiceCards(step, stepId) {
    var cards = [];
    var branches = step.branches || {};
    var savedChoice = state.branches_chosen[stepId];
    var isLocked = !!savedChoice;

    for (var key in branches) {
      if (!branches.hasOwnProperty(key)) continue;
      var branch = branches[key];
      var isSelected = (savedChoice === key);
      var card = document.createElement('div');
      card.className = 'uce-branch-card';
      card.setAttribute('data-uce-branch', key);
      card.setAttribute('data-uce-step', stepId);

      if (isSelected) {
        card.classList.add('uce-branch-card--selected');
      }
      if (isLocked && !isSelected) {
        card.classList.add('uce-branch-card--disabled');
      }

      var iconHtml = '';
      if (branch.icon) {
        var iconType = branch.icon_type || 'fa-solid';
        iconHtml = '<span class="uce-branch-icon"><i class="' + iconType + ' ' + branch.icon + '"></i></span>';
      }

      var label = branch.label || key.charAt(0).toUpperCase() + key.slice(1);
      var desc = branch.description || '';

      // Build card content
      var html = iconHtml +
        '<span class="uce-branch-label">' + escapeHtml(label) + '</span>' +
        (desc ? '<span class="uce-branch-desc">' + escapeHtml(desc) + '</span>' : '');

      if (isSelected && state.current_step !== '__complete__') {
        // Selected AND still editable: wrap badge + reset in a flex row
        html += '<div class="uce-branch-actions">' +
          '<span class="uce-branch-badge"><span class="uce-badge uce-badge--selected">SELECTED</span></span>' +
          '<button class="uce-branch-reset" data-uce-step="' + escapeHtml(stepId) + '" aria-label="Reset choice">&times; Reset</button>' +
          '</div>';
      } else if (!isSelected && !isLocked) {
        // Not locked: show Select now button
        html += '<button class="uce-branch-select">Select now</button>';
      }
      // Locked and not selected: no button, just disabled

      card.innerHTML = html;

      cards.push(card);
    }

    return cards;
  }

  function handleBranchSelection(branchCard) {
    var stepId = branchCard.getAttribute('data-uce-step');
    var branchKey = branchCard.getAttribute('data-uce-branch');
    if (!stepId || !branchKey) return;

    // Save choice
    UCEState.chooseBranch(useCaseId, stepId, branchKey);
    // Also mark as completed so findCurrentIndex lands here and Next unlocks
    UCEState.completeStep(useCaseId, stepId);

    // Lock all cards in this step
    var allCards = dom.container.querySelectorAll('.uce-branch-card[data-uce-step="' + stepId + '"]');
    for (var i = 0; i < allCards.length; i++) {
      var c = allCards[i];
      var bk = c.getAttribute('data-uce-branch');
      if (bk === branchKey) {
        c.classList.add('uce-branch-card--selected');
        // Add SELECTED badge
        if (!c.querySelector('.uce-badge--selected')) {
          var badge = document.createElement('span');
          badge.className = 'uce-badge uce-badge--selected';
          badge.textContent = 'SELECTED';
          c.appendChild(badge);
        }
      } else {
        c.classList.add('uce-branch-card--disabled');
        var btn = c.querySelector('.uce-branch-select');
        if (btn) btn.style.display = 'none';
      }
    }

    // Re-filter visible steps
    visibleSteps = UCEState.getVisibleSteps(useCaseId, steps);
    renderFlow();
    updateProgress();

    /* ── Re-resolve expressions now that a branch has been chosen ─── */
    if (window.ExpressionResolver) {
      window.ExpressionResolver.refreshAllSteps();
    }

    // Scroll to new current step
    var newIdx = findCurrentIndex();
    currentIndex = newIdx;
    scrollToStep(newIdx);

    emitEvent('branch-chosen', { stepId: stepId, branchKey: branchKey });
  }

  /* ── Choice Step Renderer (Phase 2-01) ─────────────────────────── */

  function renderChoiceStep(step) {
    var stepId = step.id || '';
    var question = step.question || 'Choose an option:';
    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--choice';
    div.setAttribute('data-uce-step-id', stepId);

    // Question heading
    var questionEl = document.createElement('p');
    questionEl.className = 'uce-choice-question';
    if (step.icon) {
      var iconType = step.icon_type || 'fa-solid';
      questionEl.innerHTML = '<i class="' + iconType + ' ' + step.icon + '"></i> ' + question;
    } else {
      questionEl.textContent = question;
    }
    div.appendChild(questionEl);

    // Cards container
    var cardsContainer = document.createElement('div');
    cardsContainer.className = 'uce-branch-cards';
    cardsContainer.setAttribute('data-uce-choice-step', stepId);

    var cards = renderChoiceCards(step, stepId);
    for (var i = 0; i < cards.length; i++) {
      cardsContainer.appendChild(cards[i]);
    }

    div.appendChild(cardsContainer);

    // Attach click handlers to branch cards
    var cardButtons = div.querySelectorAll('.uce-branch-select');
    for (var j = 0; j < cardButtons.length; j++) {
      (function(btn) {
        btn.addEventListener('click', function(e) {
          e.preventDefault();
          e.stopPropagation();
          var card = btn.closest('.uce-branch-card');
          if (card) {
            handleBranchSelection(card);
          }
        });
      })(cardButtons[j]);
    }

    return div;
  }

  /* ── Checkpoint Step Renderer (Phase 2-02 — multi-task) ─────────── */

  function renderCheckpointStep(step) {
    var stepId = step.id || '';
    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--checkpoint';
    div.setAttribute('data-uce-step-id', stepId);

    // Body text
    if (step.body) {
      var bodyP = document.createElement('p');
      bodyP.className = 'uce-checkpoint-body';
      bodyP.innerHTML = step.body;
      div.appendChild(bodyP);
    }

    // Collect all API tasks — support both flat api config and tasks array
    var checkpoint = step.checkpoint || {};
    var apiConfig = checkpoint.api || {};
    var tasks = [];

    if (apiConfig.tasks && Array.isArray(apiConfig.tasks) && apiConfig.tasks.length > 0) {
      // Multi-task mode
      tasks = apiConfig.tasks.map(function(t, i) {
        return {
          id: t.id || ('task-' + i),
          title: t.title || ('Verify ' + (i + 1)),
          url: t.url || '',
          method: (t.method || 'GET').toUpperCase(),
          timeout_ms: t.timeout_ms || 30000,
          expect_status_code: t.expect_status_code,
          response_path: t.response_path || '',
          response_type: (t.response_type || 'text').toLowerCase(),
          on_failure: t.on_failure || checkpoint.on_failure || 'retry-or-skip'
        };
      });
    } else {
      // Legacy single-task mode
      tasks.push({
        id: 'task-0',
        title: 'Verify API Connectivity',
        url: apiConfig.url || '',
        method: (apiConfig.method || 'GET').toUpperCase(),
        timeout_ms: apiConfig.timeout_ms || 30000,
        expect_status_code: apiConfig.expect_status_code,
        response_path: apiConfig.response_path || '',
        response_type: (apiConfig.response_type || 'text').toLowerCase(),
        on_failure: checkpoint.on_failure || 'retry-or-skip'
      });
    }

    var onFail = checkpoint.on_failure || 'retry-or-skip';

    // Wrapper container
    var wrapper = document.createElement('div');
    wrapper.className = 'uce-checkpoint-wrapper';

    // Header: title + duration
    var header = document.createElement('div');
    header.className = 'uce-checkpoint-header';
    header.innerHTML =
      '<span class="uce-checkpoint-header-title"><svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 12l2 2 4-4"/><circle cx="12" cy="12" r="10"/></svg> Verification Tasks</span>' +
      (step.duration ? '<span class="uce-checkpoint-header-duration">' + escapeHtml(step.duration) + '</span>' : '');
    wrapper.appendChild(header);

    // Task cards
    var taskCardsDiv = document.createElement('div');
    taskCardsDiv.className = 'uce-checkpoint-tasks';

    var taskElements = [];
    var allVerifyBtns = [];
    var allRetryBtns = [];
    var allSkipBtns = [];

    // Raw output toggle — create BEFORE task loop so _rawPre is available
    var rawDetails = document.createElement('details');
    rawDetails.className = 'uce-checkpoint-raw-output';
    var rawSummary = document.createElement('summary');
    rawSummary.textContent = 'Show Raw Output';
    rawDetails.appendChild(rawSummary);
    var rawPre = document.createElement('pre');
    rawPre.className = 'uce-checkpoint-raw-pre';
    rawPre.textContent = '';
    rawDetails.appendChild(rawPre);

    tasks.forEach(function(task, idx) {
      var card = document.createElement('div');
      card.className = 'uce-checkpoint-task';
      card.setAttribute('data-uce-task-id', task.id);

      // Task header: title + status badge
      var taskHead = document.createElement('div');
      taskHead.className = 'uce-checkpoint-task-header';
      taskHead.innerHTML =
        '<span class="uce-checkpoint-task-title">' + escapeHtml(task.title) + '</span>' +
        '<span class="uce-checkpoint-task-status uce-checkpoint-status--pending">Pending</span>';
      card.appendChild(taskHead);

      // Task status/error area
      var taskStatus = document.createElement('div');
      taskStatus.className = 'uce-checkpoint-task-status-area';
      taskStatus.innerHTML = '<span class="uce-checkpoint-task-message">Ready to verify</span>';
      card.appendChild(taskStatus);

      // Task actions (verify/retry/skip per task)
      var taskActions = document.createElement('div');
      taskActions.className = 'uce-checkpoint-task-actions';

      var verifyBtn = document.createElement('button');
      verifyBtn.className = 'uce-btn uce-btn--primary uce-checkpoint-verify';
      verifyBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Verify';
      verifyBtn.setAttribute('data-uce-action', 'verify');
      verifyBtn.setAttribute('data-uce-task-id', task.id);
      taskActions.appendChild(verifyBtn);

      var retryBtn = document.createElement('button');
      retryBtn.className = 'uce-btn uce-btn--secondary uce-checkpoint-retry';
      retryBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg> Retry';
      retryBtn.style.display = 'none';
      retryBtn.setAttribute('data-uce-action', 'retry');
      retryBtn.setAttribute('data-uce-task-id', task.id);
      taskActions.appendChild(retryBtn);

      var skipBtn = document.createElement('button');
      skipBtn.className = 'uce-btn uce-btn--outline uce-checkpoint-skip';
      skipBtn.textContent = 'Skip';
      skipBtn.style.display = 'none';
      skipBtn.setAttribute('data-uce-action', 'skip');
      skipBtn.setAttribute('data-uce-task-id', task.id);
      taskActions.appendChild(skipBtn);

      card.appendChild(taskActions);
      taskCardsDiv.appendChild(card);

      // Store refs
      var te = { card: card, statusBadge: taskHead.querySelector('.uce-checkpoint-task-status'), statusArea: taskStatus, message: taskStatus.querySelector('.uce-checkpoint-task-message'), verifyBtn: verifyBtn, retryBtn: retryBtn, skipBtn: skipBtn, _rawPre: rawPre };
      taskElements.push(te);
      allVerifyBtns.push(verifyBtn);
      allRetryBtns.push(retryBtn);
      allSkipBtns.push(skipBtn);

      // Wire up per-task buttons
      (function(t, vb, rb, sb, te) {
        vb.addEventListener('click', function() {
          runCheckpoint(t.url, t.method, t.timeout_ms, t.expect_status_code, t.response_path, t.response_type, t.on_failure, stepId, te, vb, rb, sb);
        });
        rb.addEventListener('click', function() {
          rb.style.display = 'none';
          runCheckpoint(t.url, t.method, t.timeout_ms, t.expect_status_code, t.response_path, t.response_type, t.on_failure, stepId, te, vb, rb, sb);
        });
        sb.addEventListener('click', function() {
          setTaskStatus(te, 'skipped', 'Skipped');
          UCEState.completeStep(useCaseId, stepId);
          emitEvent('checkpoint-result', { success: false, skipped: true, stepId: stepId });
          setTimeout(function() {
            if (currentIndex < visibleSteps.length - 1) {
              currentIndex++;
              scrollToStep(currentIndex);
              updateProgress();
              emitEvent('step-change', { index: currentIndex, stepId: visibleSteps[currentIndex].id });
            }
          }, 800);
        });
      })(task, verifyBtn, retryBtn, skipBtn, te);
    });

    wrapper.appendChild(taskCardsDiv);

    // Shared actions: Retry All + Skip
    var sharedActions = document.createElement('div');
    sharedActions.className = 'uce-checkpoint-actions';

    var globalVerifyBtn = document.createElement('button');
    globalVerifyBtn.className = 'uce-btn uce-btn--primary uce-checkpoint-verify-all';
    globalVerifyBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Verify All';
    globalVerifyBtn.addEventListener('click', function() {
      allVerifyBtns.forEach(function(vb) { if (vb.style.display !== 'none') vb.click(); });
    });
    sharedActions.appendChild(globalVerifyBtn);

    var globalRetryBtn = document.createElement('button');
    globalRetryBtn.className = 'uce-btn uce-btn--secondary uce-checkpoint-retry-all';
    globalRetryBtn.textContent = 'Retry Verification';
    globalRetryBtn.style.display = 'none';
    globalRetryBtn.addEventListener('click', function() {
      allRetryBtns.forEach(function(rb) { if (rb.style.display !== 'none') rb.click(); });
    });
    sharedActions.appendChild(globalRetryBtn);

    var globalSkipBtn = document.createElement('button');
    globalSkipBtn.className = 'uce-btn uce-btn--outline uce-checkpoint-skip-all';
    globalSkipBtn.textContent = 'Skip';
    globalSkipBtn.style.display = 'none';
    globalSkipBtn.addEventListener('click', function() {
      allSkipBtns.forEach(function(sb) { if (sb.style.display !== 'none') sb.click(); });
    });
    sharedActions.appendChild(globalSkipBtn);

    wrapper.appendChild(sharedActions);
    wrapper.appendChild(rawDetails);

    // Store refs globally
    wrapper._taskElements = taskElements;
    wrapper._allVerifyBtns = allVerifyBtns;
    wrapper._allRetryBtns = allRetryBtns;
    wrapper._allSkipBtns = allSkipBtns;
    wrapper._rawPre = rawPre;

    div.appendChild(wrapper);
    return div;
  }

  /** Set status for a single task card. */
  function setTaskStatus(taskEl, state, message) {
    var badge = taskEl.statusBadge;
    var msg = taskEl.message;
    if (!badge || !msg) return;

    badge.className = 'uce-checkpoint-task-status uce-checkpoint-status--' + state;
    var labels = { success: 'Passed', error: 'Failed', verifying: 'Verifying…', skipped: 'Skipped' };
    badge.textContent = labels[state] || state;
    var icons = { success: '&#10003; ', error: '&#9888; ', verifying: '&#8635; ', skipped: '&#8212; ' };
    msg.innerHTML = (icons[state] || '') + escapeHtml(message);
  }

  /* ── Checkpoint API Runner (Phase 2-02) ────────────────────────── */

  /**
   * Run a checkpoint API verification.
   * Enforces localhost-only URLs.
   */
  function runCheckpoint(url, method, timeoutMs, expectStatus, responsePath, responseType, onFail, stepId, panel, verifyBtn, retryBtn, skipBtn) {
    // Validate URL
    if (!url) {
      setStatus(panel, 'error', 'No API URL configured');
      return;
    }

    // Localhost enforcement
    var isLocal = /^https?:\/\/(localhost|127\.0\.0\.1)/i.test(url);
    if (!isLocal) {
      setStatus(panel, 'error', 'Only localhost/127.0.0.1 URLs are allowed for checkpoint verification');
      return;
    }

    // Set loading state
    setPanelState(panel, 'verifying', verifyBtn);

    // Create abort controller for timeout
    var controller = new AbortController();
    var timer = setTimeout(function() {
      controller.abort();
      handleCheckpointResult(panel, false, new Error('Request timed out (' + timeoutMs + 'ms)'), onFail, stepId, null);
    }, timeoutMs);

    // Make the fetch request
    fetch(url, {
      method: method,
      signal: controller.signal,
      headers: {
        'Accept': 'application/json'
      }
    })
    .then(function(response) {
      clearTimeout(timer);

      // Capture headers for raw output
      var headersObj = {};
      response.headers.forEach(function(val, key) { headersObj[key] = val; });

      // Check status code if expected
      if (expectStatus !== undefined && response.status !== expectStatus) {
        handleCheckpointResult(panel, false, new Error('Expected status ' + expectStatus + ', got ' + response.status), onFail, stepId, {
          status: response.status,
          statusText: response.statusText,
          headers: headersObj
        });
        return;
      }

      // Evaluate response path if configured
      if (responsePath) {
        if (responseType === 'json') {
          return response.json().then(function(data) {
            var result = evaluateJsonPath(data, responsePath);
            if (result) {
              handleCheckpointResult(panel, true, null, onFail, stepId, {
                url: url,
                method: method,
                status: response.status,
                statusText: response.statusText,
                headers: headersObj,
                data: data
              });
            } else {
              handleCheckpointResult(panel, false, new Error('Response path "' + responsePath + '" did not match'), onFail, stepId, {
                url: url,
                method: method,
                status: response.status,
                statusText: response.statusText,
                headers: headersObj,
                data: data
              });
            }
          }).catch(function(err) {
            handleCheckpointResult(panel, false, err, onFail, stepId, {
              url: url,
              method: method,
              status: response.status,
              statusText: response.statusText,
              headers: headersObj,
              error: err.message
            });
          });
        } else {
          // text mode — compare whole response body as string
          return response.text().then(function(text) {
            var result = evaluateTextMatch(text, responsePath);
            if (result) {
              handleCheckpointResult(panel, true, null, onFail, stepId, {
                url: url,
                method: method,
                status: response.status,
                statusText: response.statusText,
                headers: headersObj,
                body: text
              });
            } else {
              handleCheckpointResult(panel, false, new Error('Response text "' + responsePath + '" did not match'), onFail, stepId, {
                url: url,
                method: method,
                status: response.status,
                statusText: response.statusText,
                headers: headersObj,
                body: text
              });
            }
          }).catch(function(err) {
            handleCheckpointResult(panel, false, err, onFail, stepId, {
              url: url,
              method: method,
              status: response.status,
              statusText: response.statusText,
              headers: headersObj,
              error: err.message
            });
          });
        }
      } else {
        // No path check — just status code is enough
        handleCheckpointResult(panel, true, null, onFail, stepId, {
          url: url,
          method: method,
          status: response.status,
          statusText: response.statusText,
          headers: headersObj
        });
      }
    })
    .catch(function(err) {
      clearTimeout(timer);
      // Network error — no response available
      handleCheckpointResult(panel, false, err, onFail, stepId, {
        url: url,
        method: method,
        error: err.message,
        errorType: 'network'
      });
    });
  }

  /**
   * Handle checkpoint result — update UI accordingly.
   */
  function handleCheckpointResult(panel, success, error, onFail, stepId, responseData) {
    if (success) {
      setTaskStatus(panel, 'success', 'Verified successfully!');
      // Hide verify button on success
      var verifyBtn = panel.verifyBtn;
      if (verifyBtn) verifyBtn.style.display = 'none';
      // Mark step complete
      UCEState.completeStep(useCaseId, stepId);

      // Auto-advance after 1s
      setTimeout(function() {
        if (currentIndex < visibleSteps.length - 1) {
          currentIndex++;
          scrollToStep(currentIndex);
          updateProgress();
          emitEvent('step-change', { index: currentIndex, stepId: visibleSteps[currentIndex].id });
        }
      }, 1000);
    } else {
      var errMsg = error ? (error.message || 'Verification failed') : 'Verification failed';
      setTaskStatus(panel, 'error', errMsg);

      // Hide verify button on failure
      var verifyBtn = panel.verifyBtn;
      if (verifyBtn) verifyBtn.style.display = 'none';

      // Show appropriate buttons based on on_fail behavior
      var retryBtn = panel.retryBtn;
      var skipBtn = panel.skipBtn;
      if (onFail === 'retry-or-skip' || onFail === 'retry-only') {
        if (retryBtn) retryBtn.style.display = '';
      }
      if (onFail === 'retry-or-skip' || onFail === 'skip-only') {
        if (skipBtn) skipBtn.style.display = '';
      }
      if (onFail === 'fail-step') {
        // Disable next by marking step as not complete
      }
    }

    emitEvent('checkpoint-result', { success: success, stepId: stepId, error: error ? error.message : null });

    // Populate raw output with beautified response data
    var rawPre = panel ? panel._rawPre : null;
    if (!rawPre) return;

    var block = [];

    // Result banner
    var cls = success ? 'uce-raw-pass' : 'uce-raw-fail';
    block.push('<span class="' + cls + '">' + (success ? '[PASS]' : '[FAIL]') + '</span> ' + (error ? escapeHtml(error.message) : 'Verification complete'));
    block.push('');

    // Response data
    if (responseData) {
      if (responseData.url) {
        block.push('<span class="uce-raw-url">&gt; ' + escapeHtml(responseData.method) + ' ' + escapeHtml(responseData.url) + '</span>');
      }
      if (responseData.status) {
        block.push('  <span class="uce-raw-status">Status:</span> ' + escapeHtml(String(responseData.status)) + ' ' + escapeHtml(responseData.statusText || ''));
      }
      if (responseData.headers) {
        block.push('  <span class="uce-raw-info">Headers:</span>');
        for (var h in responseData.headers) {
          if (responseData.headers.hasOwnProperty(h)) {
            block.push('    <span class="uce-raw-header-key">' + escapeHtml(h) + '</span>: <span class="uce-raw-header-val">' + escapeHtml(responseData.headers[h]) + '</span>');
          }
        }
      }
      if (responseData.data !== undefined && responseData.data !== null) {
        block.push('');
        block.push('  <span class="uce-raw-body-label">Response Body:</span>');
        block.push('  <span class="uce-raw-body">' + escapeHtml(JSON.stringify(responseData.data, null, 2)).replace(/\n/g, '\n  ') + '</span>');
      } else if (responseData.body !== undefined && responseData.body !== null) {
        block.push('');
        block.push('  <span class="uce-raw-body-label">Response Body:</span>');
        block.push('  <span class="uce-raw-body">' + escapeHtml(String(responseData.body)).replace(/\n/g, '\n  ') + '</span>');
      }
    }
    if (responseData && responseData.error) {
      block.push('');
      block.push('  <span class="uce-raw-error">' + (responseData.errorType === 'network' ? 'Network Error' : 'Parse Error') + ':</span> ' + escapeHtml(responseData.error));
    }

    // Accumulate results
    var existing = rawPre.innerHTML || '';
    if (existing) {
      rawPre.innerHTML = existing + '<hr style="border:none;border-top:1px solid rgba(0,0,0,0.08);margin:16px 0;">' + block.join('\n');
    } else {
      rawPre.innerHTML = block.join('\n');
    }
  }

  /**
   * Evaluate a simple JSON path expression.
   * Supports: $.field == "value" or $.field == 123
   */
  function evaluateJsonPath(data, expr) {
    if (!expr) return false;
    expr = expr.trim();

    // If data is a string, parse it first
    if (typeof data === 'string') {
      try { data = JSON.parse(data); } catch(e) { return false; }
    }

    // Regex: $.field == "value"
    var match = expr.match(/^(\$\.[\w.]+)\s*(===|==|!=|>=|<=|>|<)\s*(.+)$/);
    if (!match) return false;

    var pathStr = match[1];
    var operator = match[2];
    var expectedStr = match[3].trim();

    // Strip surrounding quotes
    if ((expectedStr.startsWith('"') && expectedStr.endsWith('"')) ||
        (expectedStr.startsWith("'") && expectedStr.endsWith("'"))) {
      expectedStr = expectedStr.slice(1, -1);
    }

    // Navigate path — skip empty first segment from $. 
    var segments = pathStr.replace(/^\$/, '').split('.').filter(function(s){return s;});
    var current = data;
    for (var i = 0; i < segments.length; i++) {
      if (current === null || current === undefined) return false;
      current = current[segments[i]];
    }

    // Compare
    if (operator === '==') {
      return String(current) === expectedStr;
    }
    if (operator === '===') {
      return current === expectedStr;
    }
    if (operator === '!=') {
      return String(current) !== expectedStr;
    }
    if (operator === '>=' || operator === '<=' || operator === '>' || operator === '<') {
      var num = Number(expectedStr);
      if (isNaN(num)) return false;
      if (operator === '>=') return Number(current) >= num;
      if (operator === '<=') return Number(current) <= num;
      if (operator === '>')  return Number(current) > num;
      if (operator === '<')  return Number(current) < num;
    }
    return false;
  }

  /**
   * Match response text against a response_path expression.
   * Supports: "OK" (contains), "status == OK" (contains key=value), or plain text.
   */
  function evaluateTextMatch(text, expr) {
    if (!expr) return false;
    expr = expr.trim();

    // If it looks like a JSON path expression (starts with $), try JSON parse first
    if (expr.startsWith('$')) {
      try {
        var obj = JSON.parse(text);
        return evaluateJsonPath(obj, expr);
      } catch(e) {
        // Not valid JSON, fall through to text matching
      }
    }

    // Plain text contains check — strip surrounding quotes if present
    var searchStr = expr;
    if ((searchStr.startsWith('"') && searchStr.endsWith('"')) ||
        (searchStr.startsWith("'") && searchStr.endsWith("'"))) {
      searchStr = searchStr.slice(1, -1);
    }

    return text.indexOf(searchStr) !== -1;
  }

  /**
   * Update checkpoint panel status UI.
   */
  function setStatus(panel, state, message) {
    var statusEl = panel.querySelector('.uce-checkpoint-status');
    if (!statusEl) return;

    statusEl.className = 'uce-checkpoint-status uce-checkpoint-status--' + state;

    var icon = '';
    if (state === 'success') icon = '\u2713 ';
    if (state === 'error') icon = '\u2717 ';
    if (state === 'verifying') icon = '\u21bb ';
    if (state === 'skipped') icon = '\u23ED ';

    statusEl.innerHTML = '<span class="uce-checkpoint-' + state + '">' + icon + escapeHtml(message) + '</span>';
  }

  /**
   * Set panel to verifying state.
   */
  function setPanelState(panel, state, verifyBtn) {
    if (state === 'verifying') {
      verifyBtn.disabled = true;
      verifyBtn.innerHTML = '<span class="uce-spinner"></span> Verifying...';
    }
  }

  /* ── Side Quest System (Phase 4-01) ─────────────────────────────── */

  /**
   * Render side quest trigger elements for a narrative step.
   * Returns an array of DOM elements (fly-out panels).
   */
  function renderSideQuestTriggers(sqIds, originStepId) {
    var elements = [];

    for (var i = 0; i < sqIds.length; i++) {
      var sqId = sqIds[i];
      var sq = findSideQuest(sqId);
      if (!sq) continue;

      /* Skip side quests whose linked use case is hidden */
      if (sq.links_to && !sq.links_to.includes('://') && window.__UCE_HIDDEN_MAP__ && window.__UCE_HIDDEN_MAP__[sq.links_to]) {
        continue;
      }

      var tag = UCEState.getSideQuestTag(useCaseId, sqId);
      var completed = UCEState.isSideQuestCompleted(useCaseId, sqId);
      var domId = 'uce-sq-' + sqId + '-' + originStepId;

      var flyout = document.createElement('div');
      flyout.className = 'uce-sidequest-flyout';
      flyout.setAttribute('data-uce-sq-id', sqId);
      flyout.setAttribute('data-uce-origin-step', originStepId);
      flyout.setAttribute('data-uce-done', completed ? 'true' : 'false');

      // Check if side quest is expanded by default
      var isExpanded = !!sq.expanded;

      // Flyout header
      var header = document.createElement('div');
      header.className = 'uce-sidequest-header';
      header.innerHTML =
        '<i class="fa-solid fa-scroll uce-sidequest-icon"></i>' +
        '<span class="uce-sidequest-title">Side Quest: ' + escapeHtml(sq.title || sqId) + '</span>' +
        '<button class="uce-sidequest-toggle" aria-label="Toggle side quest panel"><svg class="uce-sidequest-chevron" viewBox="0 0 20 20" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 7l5 5 5-5"/></svg></button>';
      flyout.appendChild(header);

      // Flyout body
      var body = document.createElement('div');
      body.className = 'uce-sidequest-body';
      body.style.display = isExpanded ? 'block' : 'none';
      if (isExpanded) {
        var chevron = header.querySelector('.uce-sidequest-chevron');
        if (chevron) {
          chevron.style.transform = 'rotate(180deg)';
        }
      }

      var desc = document.createElement('p');
      desc.className = 'uce-sidequest-desc';
      desc.textContent = sq.description || '';
      body.appendChild(desc);

      if (sq.requires && sq.requires.length > 0) {
        var req = document.createElement('p');
        req.className = 'uce-sidequest-req';
        req.innerHTML = '<strong>Requires:</strong> ' + sq.requires.join(', ');
        body.appendChild(req);
      }

      var dur = document.createElement('p');
      dur.className = 'uce-sidequest-dur';
      dur.textContent = 'Estimated: ' + (sq.duration || '~5 min');
      body.appendChild(dur);

      // Actions
      var actions = document.createElement('div');
      actions.className = 'uce-sidequest-actions';

      if (completed) {
        // DONE state
        actions.innerHTML =
          '<span class="uce-sidequest-done-label">&#10003; DONE: ' + escapeHtml(sq.title || sqId) + '</span>' +
          (sq.links_to
            ? '<button class="uce-btn uce-btn--outline uce-sidequest-view-again" data-uce-sq="' + sqId + '">[+ View again]</button>'
            : '');
      } else if (tag && tag.status === 'skipped') {
        // Skipped state — show "View again" button
        actions.innerHTML =
          '<button class="uce-btn uce-btn--secondary uce-sidequest-take-now" data-uce-sq="' + sqId + '"><i class="fa-solid fa-circle-question"></i> Take now!</button>' +
          '<button class="uce-btn uce-btn--outline uce-sidequest-dismiss" data-uce-sq="' + sqId + '">Dismiss</button>';
      } else {
        // Fresh state — show Take/Skip buttons
        var takeLabel = sq.links_to ? 'Take now!' : 'Start!';
        actions.innerHTML =
          '<button class="uce-btn uce-btn--primary uce-sidequest-take-now" data-uce-sq="' + sqId + '"><i class="fa-solid fa-circle-question"></i> ' + takeLabel + '</button>' +
          '<button class="uce-btn uce-btn--outline uce-sidequest-skip" data-uce-sq="' + sqId + '">Skip for now</button>';
      }

      body.appendChild(actions);
      flyout.appendChild(body);

      // Toggle button
      (function(fly) {
        var tog = fly.querySelector('.uce-sidequest-toggle');
        var chevron = fly.querySelector('.uce-sidequest-chevron');
        var bdy = fly.querySelector('.uce-sidequest-body');
        tog.addEventListener('click', function(e) {
          e.stopPropagation();
          var isOpen = bdy.style.display !== 'none';
          bdy.style.display = isOpen ? 'none' : 'block';
          if (chevron) {
            chevron.style.transform = isOpen ? '' : 'rotate(180deg)';
          }
        });
      })(flyout);

      // Take now button
      (function(btn, sqId) {
        if (!btn) return;
        btn.addEventListener('click', function(e) {
          e.stopPropagation();
          handleTakeNow(sqId, originStepId);
        });
      })(actions.querySelector('.uce-sidequest-take-now'), sqId);

      // Skip button
      (function(btn, sqId) {
        if (!btn) return;
        btn.addEventListener('click', function(e) {
          e.stopPropagation();
          handleSkipForNow(sqId, originStepId);
        });
      })(actions.querySelector('.uce-sidequest-skip'), sqId);

      // Dismiss button
      (function(btn, sqId) {
        if (!btn) return;
        btn.addEventListener('click', function(e) {
          e.stopPropagation();
          handleSkipForNow(sqId, originStepId);
        });
      })(actions.querySelector('.uce-sidequest-dismiss'), sqId);

      // View again button
      (function(btn, sqId) {
        if (!btn) return;
        btn.addEventListener('click', function(e) {
          e.stopPropagation();
          handleViewAgain(sqId);
        });
      })(actions.querySelector('.uce-sidequest-view-again'), sqId);

      elements.push(flyout);
    }

    return elements;
  }

  /** Find a side quest definition by ID. */
  function findSideQuest(sqId) {
    for (var i = 0; i < sideQuests.length; i++) {
      if (sideQuests[i].id === sqId) return sideQuests[i];
    }
    return null;
  }

  /**
   * Render a collapsible Resources section for a step.
   * Returns a DOM element or null if no valid resources.
   */
  function renderResources(resources) {
    if (!resources || resources.length === 0) return null;

    // Filter out entries without a title
    var valid = [];
    for (var i = 0; i < resources.length; i++) {
      if (resources[i].title && resources[i].url) {
        valid.push(resources[i]);
      }
    }
    if (valid.length === 0) return null;

    var flyout = document.createElement('div');
    flyout.className = 'uce-resources-flyout';

    // Flyout header
    var header = document.createElement('div');
    header.className = 'uce-resources-header';
    header.innerHTML =
      '<i class="fa-solid fa-globe uce-resources-icon"></i>' +
      '<span class="uce-resources-title">Resources</span>' +
      '<button class="uce-resources-toggle" aria-label="Toggle resources panel"><svg class="uce-resources-chevron" viewBox="0 0 20 20" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 7l5 5 5-5"/></svg></button>';
    flyout.appendChild(header);

    // Flyout body (collapsed by default)
    var body = document.createElement('div');
    body.className = 'uce-resources-body';
    body.style.display = 'none';

    for (var r = 0; r < valid.length; r++) {
      var item = valid[r];
      var row = document.createElement('div');
      row.className = 'uce-resource-item';
      row.innerHTML =
        '<span class="uce-resource-title">' + escapeHtml(item.title) + '</span>' +
        '<a class="uce-resource-btn" href="' + escapeHtml(item.url) + '" target="_blank" rel="noopener noreferrer">Go</a>';
      body.appendChild(row);
    }

    flyout.appendChild(body);

    // Toggle button
    (function(fly) {
      var tog = fly.querySelector('.uce-resources-toggle');
      var chevron = fly.querySelector('.uce-resources-chevron');
      var bdy = fly.querySelector('.uce-resources-body');
      tog.addEventListener('click', function(e) {
        e.stopPropagation();
        var isOpen = bdy.style.display !== 'none';
        bdy.style.display = isOpen ? 'none' : 'block';
        if (chevron) {
          chevron.style.transform = isOpen ? '' : 'rotate(180deg)';
        }
      });
    })(flyout);

    return flyout;
  }

  /**
   * Render a "What's Next" completion step — congratulations card + recommendation grid.
   */
  function renderWhatsNextStep(step) {
    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--whats-next';

    /* ── Congratulations header card ─────────────────────────────── */
    var headerCard = document.createElement('div');
    headerCard.className = 'uce-whatsnext-header';
    headerCard.innerHTML =
      '<div class="uce-whatsnext-icon"><i class="fa-solid fa-party-horn"></i></div>' +
      '<div class="uce-whatsnext-content">' +
        '<h3 class="uce-whatsnext-title">Congratulations!</h3>' +
        '<p class="uce-whatsnext-message">' + escapeHtml(step.body || '') + '</p>' +
      '</div>';
    div.appendChild(headerCard);

    /* ── Recommendation grid ─────────────────────────────────────── */
    if (!step.recommendations || step.recommendations.length === 0) {
      return div;
    }

    var heading = document.createElement('h4');
    heading.className = 'uce-whatsnext-heading';
    heading.textContent = "What's Next";
    div.appendChild(heading);

    var grid = document.createElement('div');
    grid.className = 'uce-whatsnext-grid';

    for (var i = 0; i < step.recommendations.length; i++) {
      var rec = step.recommendations[i];
      if (!rec.use_case) continue;

      var card = document.createElement('a');
      card.className = 'uce-whatsnext-card';
      card.href = (window.__UCE_BASEURL__) + '/use-cases/' + rec.use_case + '/';
      card.setAttribute('rel', 'noopener');

      var iconHtml = '';
      if (rec.icon) {
        iconHtml = '<i class="fa-solid ' + escapeHtml(rec.icon) + '"></i>';
      }

      card.innerHTML =
        '<div class="uce-whatsnext-card-head">' +
          (iconHtml ? '<div class="uce-whatsnext-card-icon">' + iconHtml + '</div>' : '') +
          '<span class="uce-whatsnext-card-title">Loading…</span>' +
        '</div>' +
        '<p class="uce-whatsnext-card-desc"><span class="uce-whatsnext-card-text"></span></p>' +
        '<svg class="uce-whatsnext-card-arrow" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>';

      grid.appendChild(card);
    }

    div.appendChild(grid);
    return div;
  }

  /**
   * Preload use case metadata for whats_next recommendations.
   * Called after all steps are rendered to fill in titles/descriptions.
   */
  function preloadWhatsNextData(container) {
    if (!container) return;
    var cards = container.querySelectorAll('.uce-whatsnext-card');
    for (var i = 0; i < cards.length; i++) {
      var card = cards[i];
      var href = card.getAttribute('href');
      if (!href) continue;

      // Extract slug from URL: /use-cases/{slug}/
      var parts = href.replace(/\/$/, '').split('/');
      var slug = parts[parts.length - 1];
      if (!slug) continue;

      // Look for matching use case data in window (set by layout)
      var ucData = null;
      if (window.__UCE_WHATS_NEXT_DATA__ && window.__UCE_WHATS_NEXT_DATA__[slug]) {
        ucData = window.__UCE_WHATS_NEXT_DATA__[slug];
      }

      if (ucData) {
        var titleEl = card.querySelector('.uce-whatsnext-card-title');
        var textEl = card.querySelector('.uce-whatsnext-card-text');
        if (titleEl) {
          titleEl.textContent = ucData.title || slug;
          titleEl.classList.remove('uce-skeleton');
        }
        if (textEl) {
          var desc = ucData.introduction && (ucData.introduction.markdown_scenario || ucData.introduction.scenario)
            ? (ucData.introduction.markdown_scenario || ucData.introduction.scenario)
            : '';
          // Render as markdown
          textEl.innerHTML = desc ? marked.parse(desc) : '';
          textEl.classList.remove('uce-skeleton');
        }
      }
    }
  }

  /**
   * Handle "Take now!" — navigate to linked use case with return_to.
   */
  function handleTakeNow(sqId, originStepId) {
    var sq = findSideQuest(sqId);
    if (!sq) return;

    // Tag as "taken"
    UCEState.tagSideQuest(useCaseId, sqId, 'taken');
    UCEState.save();

    if (sq.links_to) {
      // Store return info in sessionStorage (cross-page)
      var baseurl = (window.__UCE_BASEURL__) || '';
      sessionStorage.setItem('uce_return_to', JSON.stringify({
        from: window.__UCE_PAGE_SLUG__ || useCaseId,
        from_title: window.__USE_CASE_TITLE__ || '',
        from_data_key: useCaseId,
        step: originStepId,
        sq_id: sqId
      }));

      // Navigate to linked use case
      var targetPage = baseurl + '/use-cases/' + sq.links_to;
      window.location.href = targetPage;
    } else {
      // Inline side quest — mark as completed
      UCEState.completeSideQuest(useCaseId, sqId);
      emitEvent('side-quest-completed', { sqId: sqId });

      /* Resolve expressions now that the side quest is completed */
      if (window.ExpressionResolver) {
        window.ExpressionResolver.refreshAllSteps();
      }
    }
  }

  /**
   * Handle "Skip for now" — tag side quest as skipped.
   */
  function handleSkipForNow(sqId, originStepId) {
    UCEState.tagSideQuest(useCaseId, sqId, 'skipped');
    UCEState.save();
    emitEvent('side-quest-skipped', { sqId: sqId });
  }

  /**
   * Handle "View again" — increment view count.
   */
  function handleViewAgain(sqId) {
    var tag = UCEState.getSideQuestTag(useCaseId, sqId);
    if (tag) {
      tag.view_count = (tag.view_count || 0) + 1;
      UCEState.tagSideQuest(useCaseId, sqId, tag.status);
    }
    emitEvent('side-quest-viewed', { sqId: sqId });
  }

  /**
   * Check sessionStorage for return-to-origin after side quest completion.
   * Shows a banner and optionally navigates back.
   */
  function checkReturnToOrigin() {
    try {
      var ret = sessionStorage.getItem('uce_return_to');
      if (!ret) return;

      var info = JSON.parse(ret);

      // If we're on the same page we started from, we came back from a side quest
      var currentPageSlug = window.__UCE_PAGE_SLUG__ || useCaseId;
      if (info.from === currentPageSlug) {
        // Remove the session storage
        sessionStorage.removeItem('uce_return_to');
        showReturnBanner(info);
      }
    } catch (e) {
      try { sessionStorage.removeItem('uce_return_to'); } catch(ex) { /* noop */ }
    }
  }

  /**
   * Update breadcrumb with parent link from side quest navigation.
   * Reads sessionStorage for return info and injects a clickable parent link
   * before the current page title.
   */
  function updateBreadcrumbFromSession() {
    try {
      var ret = sessionStorage.getItem('uce_return_to');
      if (!ret) return;

      var info = JSON.parse(ret);
      if (!info.from || !info.from_title) return;

      var breadcrumb = document.querySelector('.uce-breadcrumb-nav ul');
      if (!breadcrumb) return;

      var parentLi = document.createElement('li');
      parentLi.innerHTML = '<a href="' + (window.__UCE_BASEURL__ || '') + '/use-cases/' + info.from + '">' + escapeHtml(info.from_title) + '</a>';

      // Insert before the first active step (use_case title)
      var activeItem = breadcrumb.querySelector('li.is-active');
      if (activeItem) {
        breadcrumb.insertBefore(parentLi, activeItem);
      } else {
        breadcrumb.appendChild(parentLi);
      }
    } catch (e) {
      // Ignore errors — breadcrumb is non-critical
    }
  }

  /**
   * Show a return-to-origin banner at the top of the flow.
   */
  function showReturnBanner(info) {
    if (!dom.container) return;

    var banner = document.createElement('div');
    banner.className = 'uce-return-banner';
    banner.setAttribute('data-uce-return-from', info.from || '');

    var sq = findSideQuest(info.sq_id || '');
    var sqTitle = sq ? sq.title : info.sq_id;

    banner.innerHTML =
      '<span class="uce-return-icon">&#8617;</span>' +
      '<span class="uce-return-text">Returning from: <strong>' + escapeHtml(sqTitle || 'Side Quest') + '</strong></span>' +
      '<button class="uce-btn uce-btn--primary uce-return-continue">Continue</button>';

    // Insert at top of container
    dom.container.insertBefore(banner, dom.container.firstChild);

    // Continue button scrolls to the origin step
    (function() {
      var btn = banner.querySelector('.uce-return-continue');
      btn.addEventListener('click', function() {
        var stepEl = dom.container.querySelector('[data-uce-step="' + (info.step || '') + '"]');
        if (stepEl) {
          stepEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      });
    })();

    emitEvent('return-banner-shown', { from: info.from, step: info.step });
  }

  function renderCheckpointPlaceholder(step) {
    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--checkpoint';
    if (step.body) {
      div.innerHTML = '<p>' + step.body + '</p>';
    }
    return div;
  }

  /* ── Quiz Step Renderer (Phase 3-02) ───────────────────────────── */

  function renderQuizStep(step) {
    var stepId = step.id || '';
    var question = step.question || '';
    var options = step.options || [];
    var feedbackCorrect = step.feedback_correct || 'Correct! Well done.';
    var feedbackIncorrect = step.feedback_incorrect || 'Not quite — try again.';
    var letters = ['A', 'B', 'C', 'D', 'E'];

    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--quiz';
    div.setAttribute('data-uce-step-id', stepId);

    // Question
    var qEl = document.createElement('p');
    qEl.className = 'uce-quiz-question';
    var qText = question || '';
    try {
      var _marked = window.marked || (typeof globalThis !== 'undefined' && globalThis.marked);
      if (_marked && typeof _marked.parse === 'function') {
        qEl.innerHTML = _marked.parse(qText);
      }
    } catch (e) {
      qEl.textContent = qText;
    }
    div.appendChild(qEl);

    // Options container
    var optsContainer = document.createElement('div');
    optsContainer.className = 'uce-quiz-options';
    optsContainer.setAttribute('data-uce-quiz-step', stepId);

    for (var i = 0; i < options.length; i++) {
      var opt = options[i];
      var isCorrect = !!opt.correct;
      var optDiv = document.createElement('div');
      optDiv.className = 'uce-quiz-option';
      optDiv.setAttribute('data-uce-option', String.fromCharCode(65 + i));
      optDiv.setAttribute('data-uce-correct', isCorrect ? 'true' : 'false');
      optDiv.setAttribute('role', 'radio');
      optDiv.setAttribute('aria-checked', 'false');

      optDiv.innerHTML =
        '<span class="uce-quiz-letter">' + letters[i] + '</span>' +
        '<span class="uce-quiz-text">' + escapeHtml(opt.text || '') + '</span>' +
        '<span class="uce-quiz-icon"></span>';

      // Click handler
      (function(optEl, correct, optId) {
        optEl.addEventListener('click', function() {
          if (optEl.classList.contains('uce-quiz-option--answered')) return;

          // Record the selected option letter
          var optionLetter = optEl.getAttribute('data-uce-option');
          UCEState.recordQuizAnswer(useCaseId, stepId, optionLetter);

          // Re-render to unlock the Next button
          renderFlow();

          /* Resolve expressions now that the quiz answer is saved */
          if (window.ExpressionResolver) {
            window.ExpressionResolver.refreshAllSteps();
          }

          // Mark all options as answered
          var allOpts = optsContainer.querySelectorAll('.uce-quiz-option');
          for (var j = 0; j < allOpts.length; j++) {
            allOpts[j].classList.add('uce-quiz-option--answered');
            allOpts[j].setAttribute('aria-checked', 'false');
          }

          if (correct) {
            // Correct answer
            optEl.classList.add('uce-quiz-option--correct');
            optEl.setAttribute('aria-checked', 'true');
            optEl.querySelector('.uce-quiz-icon').innerHTML = '<i class="fa-solid fa-circle-check"></i>';

            // Show feedback
            showFeedback(optsContainer, feedbackCorrect, true);

            emitEvent('quiz-result', { stepId: stepId, correct: true, attempts: 1 });
          } else {
            // Incorrect answer — mark selected as wrong AND highlight correct
            optEl.classList.add('uce-quiz-option--incorrect');
            optEl.querySelector('.uce-quiz-icon').innerHTML = '<i class="fa-solid fa-circle-xmark"></i>';

            // Find and highlight the correct option
            for (var k = 0; k < allOpts.length; k++) {
              if (allOpts[k].getAttribute('data-uce-correct') === 'true') {
                allOpts[k].classList.add('uce-quiz-option--correct');
                allOpts[k].querySelector('.uce-quiz-icon').innerHTML = '<i class="fa-solid fa-circle-check"></i>';
                break;
              }
            }

            // Show feedback
            showFeedback(optsContainer, feedbackIncorrect, false);

            emitEvent('quiz-result', { stepId: stepId, correct: false, attempts: 1 });
          }
        });
      })(optDiv, isCorrect, opt.id);

      optsContainer.appendChild(optDiv);
    }

    // Restore answered state from saved quiz answer
    var savedAnswer = state.quiz_answers[stepId];
    if (savedAnswer) {
      // Find the option that was answered
      var allOpts = optsContainer.querySelectorAll('.uce-quiz-option');
      for (var k = 0; k < allOpts.length; k++) {
        var optLetter = allOpts[k].getAttribute('data-uce-option');
        var optIsCorrect = allOpts[k].getAttribute('data-uce-correct') === 'true';

        // Mark all as answered
        allOpts[k].classList.add('uce-quiz-option--answered');

        if (optLetter === savedAnswer) {
          // The user's selected option
          if (optIsCorrect) {
            allOpts[k].classList.add('uce-quiz-option--correct');
            allOpts[k].setAttribute('aria-checked', 'true');
            allOpts[k].querySelector('.uce-quiz-icon').innerHTML = '<i class="fa-solid fa-circle-check"></i>';
          } else {
            allOpts[k].classList.add('uce-quiz-option--incorrect');
            allOpts[k].querySelector('.uce-quiz-icon').innerHTML = '<i class="fa-solid fa-circle-xmark"></i>';
          }
        }

        // If user was wrong, highlight the correct one
        if (optLetter === savedAnswer && !optIsCorrect && optIsCorrect === false) {
          // Already handled above
        }
        if (optIsCorrect && savedAnswer !== optLetter) {
          allOpts[k].classList.add('uce-quiz-option--correct');
          allOpts[k].querySelector('.uce-quiz-icon').innerHTML = '<i class="fa-solid fa-circle-check"></i>';
        }
      }
    }

    // Feedback container
    var feedbackContainer = document.createElement('div');
    feedbackContainer.className = 'uce-quiz-feedback';
    feedbackContainer.setAttribute('data-uce-feedback', stepId);
    div.appendChild(optsContainer);
    div.appendChild(feedbackContainer);

    // Show feedback after the container exists
    if (savedAnswer) {
      var fbEl = div.querySelector('.uce-quiz-feedback');
      if (fbEl) {
        var userWasCorrect = false;
        for (var m = 0; m < allOpts.length; m++) {
          if (allOpts[m].getAttribute('data-uce-option') === savedAnswer && allOpts[m].getAttribute('data-uce-correct') === 'true') {
            userWasCorrect = true;
            break;
          }
        }
        fbEl.className = 'uce-quiz-feedback uce-quiz-feedback--' + (userWasCorrect ? 'correct' : 'incorrect');
        var icon = userWasCorrect ? '\u2713 ' : '\u2717 ';
        var fbText = icon + (userWasCorrect ? feedbackCorrect : feedbackIncorrect);
        var fbParsed = '';
        try {
          var _marked = window.marked || (typeof globalThis !== 'undefined' && globalThis.marked);
          if (_marked && typeof _marked.parse === 'function') {
            fbParsed = _marked.parse(fbText);
          }
        } catch (e) {
          /* fall through to plain text */
        }
        if (!fbParsed) fbParsed = fbText;
        fbEl.innerHTML = fbParsed;
      }
    }

    return div;
  }

  function showFeedback(container, message, correct) {
    var fbEl = container.nextElementSibling;
    if (!fbEl) return;
    fbEl.className = 'uce-quiz-feedback uce-quiz-feedback--' + (correct ? 'correct' : 'incorrect');
    var icon = correct ? '\u2713 ' : '\u2717 ';
    var parsed = '';
    try {
      var _marked = window.marked || (typeof globalThis !== 'undefined' && globalThis.marked);
      if (_marked && typeof _marked.parse === 'function') {
        parsed = _marked.parse(icon + message);
      }
    } catch (e) {
      /* fall through to plain text */
    }
    if (!parsed) parsed = icon + message;
    fbEl.innerHTML = parsed;
  }

  /* - Code Editor (Monaco) - */

  function renderCodeEditorStep(step) {
    var stepId = step.id || '';
    var hasFiles = step.files && Array.isArray(step.files) && step.files.length > 0;
    var isTabbed = hasFiles && step.files.length > 1;

    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--code-editor';
    div.setAttribute('data-uce-step-id', stepId);

    // Resolve body — supports markdown_body (parsed via marked) or plain body
    var resolvedBody = resolveStepBody(step);
    if (resolvedBody) {
      var bodyP = document.createElement('p');
      bodyP.className = 'uce-editor-body';
      bodyP.innerHTML = resolvedBody;
      div.appendChild(bodyP);
    }

    if (isTabbed) {
      renderCodeEditorTabbed(div, step, step.files);
    } else {
      renderCodeEditorSingle(div, step);
    }

    return div;
  }

  /* ── Code editor helpers ──────────────────────────────────── */

  /** Get a display label for a file (name field). */
  function getFileLabel(fileEntry) {
    if (fileEntry.name) return fileEntry.name;
    return 'file';
  }

  /** Normalize a Monaco language code. */
  function normalizeLang(lang) {
    if (!lang) return 'text';
    if (lang === 'yml') return 'yaml';
    if (lang === 'sh' || lang === 'bash') return 'shell';
    if (lang === 'ts') return 'typescript';
    if (lang === 'py') return 'python';
    return lang;
  }

  /** Render a single-file code editor (existing behavior). */
  function renderCodeEditorSingle(parent, step) {
    var language = normalizeLang(step.language || 'text');
    var initialCode = step.initial_code || '';
    var stepId = step.id || '';
    var fileLabel = '';

    // If step-level properties are empty, try reading from the files array
    if (!initialCode && step.files && step.files.length === 1) {
      var f = step.files[0];
      language = normalizeLang(f.language || language);
      initialCode = f.initial_code || '';
      fileLabel = f.name || '';
    }

    var saveKey = 'uce_code:' + useCaseId + ':' + stepId;

    var savedCode = '';
    try { savedCode = localStorage.getItem(saveKey) || ''; } catch(e) {}
    var codeContent = savedCode || initialCode;

    var editorContainer = document.createElement('div');
    editorContainer.className = 'uce-editor-container';
    editorContainer.setAttribute('data-uce-editor-step', stepId);

    var label = document.createElement('label');
    label.className = 'uce-editor-label';
    label.textContent = fileLabel || language.toUpperCase();

    // Copy button — right-aligned in the label bar
    var copyBtn = document.createElement('button');
    copyBtn.className = 'uce-editor-copy-btn';
    copyBtn.setAttribute('aria-label', 'Copy code');
    copyBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
    copyBtn.addEventListener('click', function () {
      var instId = stepId;
      if (window.__UCE_MONACO_INSTANCES__ && window.__UCE_MONACO_INSTANCES__[instId]) {
        var text = window.__UCE_MONACO_INSTANCES__[instId].getValue();
        if (navigator.clipboard && navigator.clipboard.writeText) {
          navigator.clipboard.writeText(text).then(function() {
            showCopyTooltip(copyBtn);
          }).catch(function() {
            fallbackCopy(text, copyBtn);
          });
        } else {
          fallbackCopy(text, copyBtn);
        }
      }
    });

    // Download button — right-aligned in the label bar
    var dlBtn = document.createElement('button');
    dlBtn.className = 'uce-editor-download-btn';
    dlBtn.setAttribute('aria-label', 'Download file');
    dlBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>';
    dlBtn.addEventListener('click', function () {
      var instId = stepId;
      var fileName = fileLabel || 'file.txt';
      if (window.__UCE_MONACO_INSTANCES__ && window.__UCE_MONACO_INSTANCES__[instId]) {
        var text = window.__UCE_MONACO_INSTANCES__[instId].getValue();
        var blob = new Blob([text], { type: 'text/plain' });
        var url = URL.createObjectURL(blob);
        var a = document.createElement('a');
        a.href = url;
        a.download = fileName;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      }
    });

    var labelWrap = document.createElement('div');
    labelWrap.className = 'uce-editor-label-row';

    var labelGroup = document.createElement('div');
    labelGroup.className = 'uce-editor-label-group';
    labelGroup.appendChild(label);

    var btnGroup = document.createElement('div');
    btnGroup.className = 'uce-editor-btn-group';
    btnGroup.appendChild(copyBtn);
    btnGroup.appendChild(dlBtn);

    labelWrap.appendChild(labelGroup);
    labelWrap.appendChild(btnGroup);
    editorContainer.appendChild(labelWrap);

    var editorDiv = document.createElement('div');
    editorDiv.className = 'uce-monaco-editor';
    editorDiv.setAttribute('id', 'monaco-' + stepId);
    editorDiv.style.height = '300px';
    editorContainer.appendChild(editorDiv);

    var actionsRow = document.createElement('div');
    actionsRow.className = 'uce-editor-actions';

    var loadInd = document.createElement('span');
    loadInd.className = 'uce-editor-load-indicator';
    loadInd.textContent = 'Saving...';
    loadInd.style.display = 'none';
    actionsRow.appendChild(loadInd);
    editorContainer.appendChild(actionsRow);

    parent.appendChild(editorContainer);
    initMonaco(editorDiv, stepId, language, codeContent, saveKey, loadInd);
  }

  /** Render a multi-file tabbed code editor. */
  function renderCodeEditorTabbed(parent, step, files) {
    var activeIndex = 0;

    /* Editor wrapper — mirrors narrative tab structure */
    var editorWrapper = document.createElement('div');
    editorWrapper.className = 'uce-editor-container';
    editorWrapper.setAttribute('data-uce-editor-step', step.id);

    /* Tab bar wrapper */
    var tabBarWrap = document.createElement('div');
    tabBarWrap.className = 'uce-panel-tabs-wrapper';

    /* Tab bar */
    var tabBar = document.createElement('div');
    tabBar.className = 'uce-panel-tabs';

    for (var fi = 0; fi < files.length; fi++) {
      (function (idx) {
        var tabBtn = document.createElement('button');
        tabBtn.className = 'uce-panel-tab' + (idx === 0 ? ' uce-panel-tab--active' : '');
        tabBtn.setAttribute('data-file-index', idx);
        tabBtn.textContent = getFileLabel(files[idx]);
        tabBtn.addEventListener('click', function () {
          setActiveCodeTab(tabBar, files[idx], idx);
        });
        tabBar.appendChild(tabBtn);
      })(fi);
    }

    tabBarWrap.appendChild(tabBar);

    /* Button group — right-aligned in the tab bar */
    var btnGroup = document.createElement('div');
    btnGroup.className = 'uce-editor-btn-group';

    // Copy button
    var copyBtn = document.createElement('button');
    copyBtn.className = 'uce-editor-copy-btn';
    copyBtn.setAttribute('aria-label', 'Copy code');
    copyBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>';
    copyBtn.addEventListener('click', function () {
      var activeIdx = 0;
      var activeTabs = tabBar.querySelectorAll('.uce-panel-tab--active');
      if (activeTabs.length > 0) {
        activeIdx = parseInt(activeTabs[0].getAttribute('data-file-index'), 10);
      }
      var activeFile = files[activeIdx];
      var instId = step.id + '-' + activeFile.name;
      if (window.__UCE_MONACO_INSTANCES__ && window.__UCE_MONACO_INSTANCES__[instId]) {
        var text = window.__UCE_MONACO_INSTANCES__[instId].getValue();
        if (navigator.clipboard && navigator.clipboard.writeText) {
          navigator.clipboard.writeText(text).then(function() {
            showCopyTooltip(copyBtn);
          }).catch(function() {
            fallbackCopy(text, copyBtn);
          });
        } else {
          fallbackCopy(text, copyBtn);
        }
      }
    });
    btnGroup.appendChild(copyBtn);

    // Download button
    var dlBtn = document.createElement('button');
    dlBtn.className = 'uce-editor-download-btn';
    dlBtn.setAttribute('aria-label', 'Download file');
    dlBtn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>';
    dlBtn.addEventListener('click', function () {
      var activeIdx = 0;
      var activeTabs = tabBar.querySelectorAll('.uce-panel-tab--active');
      if (activeTabs.length > 0) {
        activeIdx = parseInt(activeTabs[0].getAttribute('data-file-index'), 10);
      }
      var activeFile = files[activeIdx];
      var instId = step.id + '-' + activeFile.name;
      var fileName = activeFile.name || 'file.txt';
      if (window.__UCE_MONACO_INSTANCES__ && window.__UCE_MONACO_INSTANCES__[instId]) {
        var text = window.__UCE_MONACO_INSTANCES__[instId].getValue();
        var blob = new Blob([text], { type: 'text/plain' });
        var url = URL.createObjectURL(blob);
        var a = document.createElement('a');
        a.href = url;
        a.download = fileName;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      }
    });
    btnGroup.appendChild(dlBtn);

    tabBarWrap.appendChild(btnGroup);

    editorWrapper.appendChild(tabBarWrap);

    /* Active editor container */
    var activeEditorWrap = document.createElement('div');
    activeEditorWrap.className = 'uce-editor-active';
    editorWrapper.appendChild(activeEditorWrap);

    parent.appendChild(editorWrapper);

    /* Render initial file */
    if (files.length > 0) {
      setActiveCodeTab(tabBar, files[activeIndex], activeIndex);
    }
  }

  /** Activate a file tab and render its editor inside the active container. */
  function setActiveCodeTab(tabBar, fileEntry, index) {
    var tabs = tabBar.querySelectorAll('.uce-panel-tab');
    for (var t = 0; t < tabs.length; t++) {
      tabs[t].classList.remove('uce-panel-tab--active');
    }
    tabs[index].classList.add('uce-panel-tab--active');

    var wrap = tabBar.parentElement.parentElement.querySelector('.uce-editor-active');
    if (!wrap) return;
    wrap.innerHTML = '';

    var language = normalizeLang(fileEntry.language || 'text');
    var initialCode = fileEntry.initial_code || '';
    var stepId = tabBar.closest('[data-uce-step-id]').getAttribute('data-uce-step-id') || '';
    var saveKey = 'uce_code:' + useCaseId + ':' + stepId + ':' + fileEntry.name;

    var savedCode = '';
    try { savedCode = localStorage.getItem(saveKey) || ''; } catch(e) {}
    var codeContent = savedCode || initialCode;

    var editorDiv = document.createElement('div');
    editorDiv.className = 'uce-monaco-editor';
    editorDiv.setAttribute('id', 'monaco-' + stepId + '-' + encodeURIComponent(fileEntry.name));
    editorDiv.style.height = '300px';

    var actionsRow = document.createElement('div');
    actionsRow.className = 'uce-editor-actions';
    var loadInd = document.createElement('span');
    loadInd.className = 'uce-editor-load-indicator';
    loadInd.textContent = 'Saving...';
    loadInd.style.display = 'none';
    actionsRow.appendChild(loadInd);

    wrap.appendChild(editorDiv);
    wrap.appendChild(actionsRow);

    initMonaco(editorDiv, stepId + '-' + fileEntry.name, language, codeContent, saveKey, loadInd);
  }

  /** Shared Monaco initialization logic. */
  function initMonaco(editorDiv, editorId, language, codeContent, saveKey, loadInd) {
    if (typeof require !== 'function') {
      loadInd.style.display = '';
      loadInd.textContent = 'Monaco unavailable';
      loadInd.style.color = '#EF4444';
      return;
    }

    require.config({ paths: { vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs' } });

    require(['vs/editor/editor.main'], function() {
      var mon = window.monaco;
      if (!mon) {
        loadInd.style.display = '';
        loadInd.textContent = 'Monaco not found';
        loadInd.style.color = '#EF4444';
        return;
      }

      try {
        var editorInstance = mon.editor.create(editorDiv, {
          value: codeContent,
          language: language,
          theme: 'vs-dark',
          automaticLayout: true,
          minimap: { enabled: false },
          fontSize: 13,
          fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
          lineNumbers: 'on',
          scrollBeyondLastLine: false,
          renderWhitespace: 'selection',
          padding: { top: 12, bottom: 12 },
          bracketPairColorization: { enabled: true }
        });

        var isDirty = false;
        var saveTimer = null;
        editorInstance.onDidChangeModelContent(function() {
          if (!isDirty) {
            isDirty = true;
            loadInd.style.display = '';
            loadInd.textContent = 'Modified';
            loadInd.style.color = '#F59E0B';
          }
          clearTimeout(saveTimer);
          saveTimer = setTimeout(function() {
            try { localStorage.setItem(saveKey, editorInstance.getValue()); } catch(e) {}
            isDirty = false;
            loadInd.textContent = 'Saved';
            loadInd.style.color = '';
            setTimeout(function() { loadInd.style.display = 'none'; }, 1500);
          }, 500);
        });

        window.__UCE_MONACO_INSTANCES__ = window.__UCE_MONACO_INSTANCES__ || {};
        window.__UCE_MONACO_INSTANCES__[editorId] = {
          getValue: function() { return editorInstance.getValue(); },
          setValue: function(v) { editorInstance.setValue(v); },
          getEditor: function() { return editorInstance; }
        };

        var ro = new ResizeObserver(function() {
          editorInstance.layout();
        });
        ro.observe(editorDiv);
      } catch(e) {
        console.error('[UCE] Monaco create failed:', e);
        loadInd.style.display = '';
        loadInd.textContent = 'Editor error';
        loadInd.style.color = '#EF4444';
      }
    }, function(err) {
      console.error('[UCE] Monaco require failed:', err);
      loadInd.style.display = '';
      loadInd.textContent = 'Failed to load';
      loadInd.style.color = '#EF4444';
    });
  }

  /* ── Navigation ─────────────────────────────────────────────────── */

  function bindNavigation() {
    if (!dom.container) return;

    // Helper: collapse all expanded steps, then show only the target index
    function collapseAllExcept(targetIndex) {
      var stepEls = dom.container.querySelectorAll('.uce-flow-step');
      for (var i = 0; i < stepEls.length; i++) {
        var el = stepEls[i];
        var idx = parseInt(el.getAttribute('data-uce-step-index'), 10);
        // Remove expanded state from all steps except the target
        if (idx !== targetIndex) {
          el.classList.remove('uce-flow-step--expanded');
          var body = el.querySelector('.uce-step-body');
          if (body) body.style.display = 'none';
          var chevron = el.querySelector('.uce-step-chevron');
          if (chevron) chevron.style.transform = '';
        }
      }
    }

    // Event delegation on container for step nav buttons and jump-to-current
    dom.container.addEventListener('click', function (e) {
      var btn = e.target.closest('[data-uce-action]');

      // Handle nav button actions
      if (btn) {
        e.preventDefault();
        var action = btn.getAttribute('data-uce-action');

        if (action === 'prev' && currentIndex > 0) {
          var oldIndex = currentIndex;
          // Move active_step backward (just navigation)
          currentIndex--;
          // Collapse the old step, show only the new one
          collapseAllExcept(currentIndex);
          updateActiveStep();
          updateStepStates();
          updateProgress();
          scrollToStep(currentIndex);
          emitEvent('step-change', { index: currentIndex, stepId: visibleSteps[currentIndex].id });
        } else if (action === 'next') {
          // Reject if Next button is locked (interactive step not completed)
          if (btn.getAttribute('data-uce-locked') === 'true') {
            e.preventDefault();
            return;
          }
          if (currentIndex < visibleSteps.length - 1) {
            // Check if we're on the current_step (need to finish)
            var savedState = UCEState.getState(useCaseId);
            var currentStepId = savedState.current_step;

            if (visibleSteps[currentIndex].id === currentStepId) {
              // We're on the step we need to finish — complete it and advance
              markCurrentComplete();
              currentIndex++;
              // Collapse the completed step, show only the new current
              collapseAllExcept(currentIndex);
            } else {
              // Just navigation — move active_step forward
              currentIndex++;
              collapseAllExcept(currentIndex);
            }

            updateActiveStep();
            updateStepStates();
            updateProgress();
            scrollToStep(currentIndex);
            emitEvent('step-change', { index: currentIndex, stepId: visibleSteps[currentIndex].id });
          } else {
            completeUseCaseAndReturn();
          }
        }
        return; // Don't fall through to jump logic
      }

      // Jump-to-current: clicking a completed step that is NOT the active step
      // but IS the current_step (needs finishing) — jumps directly to it
      var stepEl = e.target.closest('.uce-flow-step');
      if (stepEl) {
        var stepIdx = parseInt(stepEl.getAttribute('data-uce-step-index'), 10);
        var savedState = UCEState.getState(useCaseId);
        var currentStepId = savedState.current_step;
        console.log('[UCE] Jump check:', { stepIdx, currentStepId, currentIndex, matched: currentStepId && visibleSteps[stepIdx] && visibleSteps[stepIdx].id === currentStepId, notCurrent: stepIdx !== currentIndex });
        if (currentStepId && visibleSteps[stepIdx] && visibleSteps[stepIdx].id === currentStepId && stepIdx !== currentIndex) {
          console.log('[UCE] Jumping to step:', stepIdx);
          currentIndex = stepIdx;
          collapseAllExcept(currentIndex);
          updateActiveStep();
          updateStepStates();
          updateProgress();
          scrollToStep(currentIndex);
          emitEvent('step-change', { index: currentIndex, stepId: visibleSteps[currentIndex].id });
        }
      }
    });
  }

  /** Update active_step to match current navigation position */
  function updateActiveStep() {
    var state = UCEState.getState(useCaseId);
    state.active_step = visibleSteps[currentIndex].id;
    state.updated_at = new Date().toISOString();
    UCEState.save();
  }

  /** Mark the current step as completed and advance current_step */
  function markCurrentComplete() {
    if (currentIndex >= 0 && currentIndex < visibleSteps.length) {
      var stepId = visibleSteps[currentIndex].id;
      if (stepId) {
        UCEState.completeStep(useCaseId, stepId);
      }
      // Update current_step to the NEXT step (the one we're moving to)
      if (currentIndex + 1 < visibleSteps.length) {
        var nextState = UCEState.getState(useCaseId);
        nextState.current_step = visibleSteps[currentIndex + 1].id;
        nextState.updated_at = new Date().toISOString();
        UCEState.save();
      }
    }
  }

  /**
   * Re-sync DOM classes for all steps using three-state logic:
   *   current_step → CURRENT badge (where user needs to finish)
   *   active_step  → expanded/visible body (where user is looking)
   *   completed_steps → checkmark + chevron (what's done)
   */
  function updateStepStates() {
    var stepEls = dom.container.querySelectorAll('.uce-flow-step');
    
    // Get the three source-of-truth values from localStorage
    var savedState = UCEState.getState(useCaseId);
    var currentStepId = savedState.current_step;  // Where CURRENT badge goes
    var activeStepId = savedState.active_step;     // What's expanded/viewed
    var currentStepIndex = -1;
    var activeStepIndex = -1;
    
    // Find indices for current_step and active_step
    if (currentStepId) {
      for (var si = 0; si < visibleSteps.length; si++) {
        if (visibleSteps[si].id === currentStepId) {
          currentStepIndex = si;
          break;
        }
      }
    }
    if (activeStepId) {
      for (var si = 0; si < visibleSteps.length; si++) {
        if (visibleSteps[si].id === activeStepId) {
          activeStepIndex = si;
          break;
        }
      }
    }
    
    // If use case is fully complete, no step is current
    var isComplete = (currentStepId === '__complete__');

    // Fall back to currentIndex if not found (but only when NOT complete)
    if (!isComplete) {
      if (currentStepIndex === -1) currentStepIndex = currentIndex;
      if (activeStepIndex === -1) activeStepIndex = currentIndex;
    }

    // First pass: remove all dynamic elements and classes
    for (var i = 0; i < stepEls.length; i++) {
      var el = stepEls[i];
      var existingToggle = el.querySelector('.uce-step-toggle');
      if (existingToggle) existingToggle.remove();
      // Clear current badge placeholder
      var placeholder = el.querySelector('.uce-current-placeholder');
      if (placeholder) {
        placeholder.innerHTML = '';
        placeholder.style.display = 'none';
      }
      el.classList.remove('uce-flow-step--current', 'uce-flow-step--completed', 'uce-flow-step--expanded');
    }

    // Second pass: apply correct classes and elements
    for (var i = 0; i < stepEls.length; i++) {
      var el = stepEls[i];
      var stepIndex = parseInt(el.getAttribute('data-uce-step-index'), 10);
      var stepId = el.getAttribute('data-uce-step');

      // Active step (being viewed) — gets --current for CSS body/nav visibility
      if (stepIndex === activeStepIndex) {
        el.classList.add('uce-flow-step--current');
        el.classList.add('uce-flow-step--expanded');
        // CRITICAL: Clear inline style.display so CSS takes over
        var bodyEl = el.querySelector('.uce-step-body');
        if (bodyEl) bodyEl.style.display = '';
        // Update number to icon if step has one
        var numEl = el.querySelector('.uce-step-number');
        if (numEl) {
          var stepDef = steps.find(function(s) { return s.id === stepId; });
          if (stepDef && stepDef.icon) {
            numEl.className = 'uce-step-number uce-step-number--icon';
            numEl.innerHTML = '<i class="fa-solid ' + stepDef.icon + '"></i>';
          } else {
            numEl.className = 'uce-step-number';
            numEl.textContent = String(stepIndex + 1);
          }
        }
      }
      // Current step (needs to be finished) — gets CURRENT badge only
      else if (stepIndex === currentStepIndex) {
        // Mark as jumpable for click-to-navigate
        el.classList.add('uce-flow-step--jumpable');
        // Show CURRENT badge only when active_step differs from current_step
        // (i.e., we're on a step the user must finish, not just browsing back to)
        if (activeStepIndex !== currentStepIndex) {
          var placeholder = el.querySelector('.uce-current-placeholder');
          if (placeholder) {
            placeholder.innerHTML = '<span class="uce-badge uce-badge--current">CURRENT</span>';
            placeholder.style.display = '';
          }
        }
        // Update number to icon if step has one
        var numEl = el.querySelector('.uce-step-number');
        if (numEl) {
          var stepDef = steps.find(function(s) { return s.id === stepId; });
          if (stepDef && stepDef.icon) {
            numEl.className = 'uce-step-number uce-step-number--icon';
            numEl.innerHTML = '<i class="fa-solid ' + stepDef.icon + '"></i>';
          } else {
            numEl.className = 'uce-step-number';
            numEl.textContent = String(stepIndex + 1);
          }
        }
      }
      // When use case is complete — treat everything as completed
      else if (isComplete) {
        el.classList.add('uce-flow-step--completed');
        // Add chevron to completed step
        var header = el.querySelector('.uce-step-header');
        if (header) {
          var chevronBtn = document.createElement('button');
          chevronBtn.className = 'uce-step-toggle';
          chevronBtn.setAttribute('aria-label', 'Toggle step content');
          chevronBtn.innerHTML = '<svg class="uce-step-chevron" viewBox="0 0 20 20" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 7l5 5 5-5"/></svg>';
          header.appendChild(chevronBtn);
        }
        // Update number to checkmark
        var numEl = el.querySelector('.uce-step-number');
        if (numEl) numEl.innerHTML = '<i class="fa-solid fa-check"></i>';
      }
      // Step is completed
      else if (UCEState.isStepCompleted(useCaseId, stepId)) {
        el.classList.add('uce-flow-step--completed');
        // Add chevron to completed step
        var header = el.querySelector('.uce-step-header');
        if (header) {
          var chevronBtn = document.createElement('button');
          chevronBtn.className = 'uce-step-toggle';
          chevronBtn.setAttribute('aria-label', 'Toggle step content');
          chevronBtn.innerHTML = '<svg class="uce-step-chevron" viewBox="0 0 20 20" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 7l5 5 5-5"/></svg>';
          header.appendChild(chevronBtn);
        }
        // Update number to checkmark
        var numEl = el.querySelector('.uce-step-number');
        if (numEl) numEl.innerHTML = '<i class="fa-solid fa-check"></i>';
      }
      // Future step
      else {
        // Reset number for future steps
        var numEl = el.querySelector('.uce-step-number');
        if (numEl) numEl.textContent = String(stepIndex + 1);
      }
    }
  }

  /**
   * Complete the use case and optionally return to origin (side quest flow).
   * Called when user clicks "Complete" on the last step.
   */
  function completeUseCaseAndReturn() {
    // Mark all remaining visible steps as completed
    for (var i = 0; i < visibleSteps.length; i++) {
      var sid = visibleSteps[i].id;
      if (sid && !UCEState.isStepCompleted(useCaseId, sid)) {
        UCEState.completeStep(useCaseId, sid);
      }
    }
    // Set current_step to 'complete' sentinel so no step is treated as "current"
    var state = UCEState.getState(useCaseId);
    state.current_step = '__complete__';
    state.active_step = null;
    state.updated_at = new Date().toISOString();
    UCEState.save();

    // Check for return-to-origin
    try {
      var ret = sessionStorage.getItem('uce_return_to');
      if (ret) {
        var info = JSON.parse(ret);

        // Mark side quest as completed on the ORIGIN use case's state
        if (info.sq_id && info.from_data_key) {
          UCEState.completeSideQuest(info.from_data_key, info.sq_id);
          UCEState.save();
          // Force-sync persistence before redirect (debounce would lose it)
          UCEState._forceSync();
        }

        // Keep return info for the origin page's checkReturnToOrigin
        sessionStorage.setItem('uce_return_to', JSON.stringify(info));

        // Redirect back to origin
        var target = (window.__UCE_BASEURL__ || '') + '/use-cases/' + info.from;
        window.location.href = target;
        return;
      }
    } catch (e) { /* noop */ }

    // No return — re-render to hide Finish button, update breadcrumb
    renderFlow();
    updateStepStates();
    updateProgress();

    // Render whats_next completion section
    if (whatsNext) {
      renderWhatsNextSection(whatsNext);
    }

    emitEvent('use-case-complete', { useCaseId: useCaseId });
  }

  function scrollToStep(index) {
    var stepEl = dom.container.querySelector('[data-uce-step-index="' + String(index) + '"]');
    if (stepEl) {
      stepEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }

  /* ── Progress Bar (Phase 2-02) ──────────────────────────────────── */

  function updateProgress() {
    var total = visibleSteps.length;
    var completed = UCEState.getCompletedCount(useCaseId);
    var pct = total > 0 ? Math.round((completed / total) * 100) : 0;

    // Add .complete class when all steps are done
    if (dom.progressFill) {
      dom.progressFill.className = (completed >= total) ? 'uce-progress-fill complete' : 'uce-progress-fill';
    }

    if (dom.progressFill) {
      dom.progressFill.style.width = pct + '%';
    }
    if (dom.topbarCount) {
      dom.topbarCount.textContent = completed + ' of ' + total;
    }

  }



  /* ── Utilities ──────────────────────────────────────────────────── */

  function escapeHtml(str) {
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function emitEvent(name, detail) {
    var event = new CustomEvent('uce:' + name, {
      bubbles: true,
      detail: detail || {}
    });
    document.dispatchEvent(event);
  }

  /* ── Boot ───────────────────────────────────────────────────────── */

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  /* ── Expression Resolver (Phase 1-01) ───────────────────────────── */

  /**
   * ExpressionResolver — Core resolver engine for [[stepId.field]] / ${{stepId.field}} expressions.
   *
   * Appended to the main UCE IIFE so it can close over escapeHtml and emitEvent.
   * Attaches to window.ExpressionResolver for external consumption.
   *
   * Public API:
   *   resolve(text, useCaseId)          — Resolve all [[expr]] in plain text
   *   resolveNode(element)              — Recursively resolve text inside DOM element
   *   refreshAllSteps()                 — Re-scan all rendered step elements
   *
   * Internal helpers (not on public API):
   *   _extractExpressions(text)         — Regex helper to find [[...]] or ${{...}} patterns
   *   _resolveField(stepId, fieldName, steps, state) — Resolve one field reference
   *   _formatPlaceholder(fieldPath)     — Format broken-ref visible text
   *   _buildStepMap(steps)             — Create { stepId: step } lookup
   */
  var ExpressionResolver = (function () {
    'use strict';

    /* ── Regex: match [[...]], ${{...}}, and {{...}} ────────────────── */
    /* Supports hyphens in step IDs and field names (e.g., choose-cloud.label) */
    var EXPR_REGEX = /\[\[\s*([.\w-][.\w-]*)\s*\]\]|\$\{\{\s*([.\w-][.\w-]*)\s*\}\}|\{\{\s*([.\w-][.\w-]*)\s*\}\}/g;

    /* ── Internal: build stepId -> step map ───────────────────────── */
    function _buildStepMap(steps) {
      var map = {};
      for (var i = 0; i < steps.length; i++) {
        if (steps[i] && steps[i].id) {
          map[steps[i].id] = steps[i];
        }
      }
      return map;
    }

    /* ── Internal: extract all [[...]], ${{...}}, and {{...}} matches from text ─ */
    function _extractExpressions(text) {
      var matches = [];
      var m = null;
      var re = new RegExp(EXPR_REGEX.source, 'g');
      while ((m = re.exec(text)) !== null) {
        /* Group 1 = [[...]], Group 2 = ${{...}}, Group 3 = {{...}} */
        var expr = m[1] || m[2] || m[3];
        if (!expr) {
          console.warn('[UCE] ExpressionResolver: no expression captured from match:', m);
          continue;
        }
        matches.push({
          full: m[0],
          expression: expr,
          index: m.index,
          syntax: m[1] ? 'double-bracket' : m[2] ? 'jinja' : 'liquid'
        });
      }
      return matches;
    }

    /* Sentinel: field exists but value is empty/unresolved (not a broken ref) */
    var EMPTY_VALUE = '__UCE_EMPTY__';

    /* ── Internal: resolve one field reference ────────────────────── */
    function _resolveField(stepId, fieldName, stepMap, state) {
      /* System variables — not backed by a step */
      if (stepId === 'platform') {
        var os = 'unknown';
        var ua = navigator.userAgent;
        if (/Mac|Macintosh|iPhone|iPad|iPod/.test(ua)) {
          os = 'darwin';
        } else if (/Linux/.test(ua)) {
          os = 'linux';
        } else if (/Win/.test(ua)) {
          os = 'windows';
        }
        return fieldName === 'os' ? os : null;
      }

      var step = stepMap[stepId];

      if (!step) {
        console.warn('[UCE] ExpressionResolver: unknown step id "' + stepId + '"');
        return null;
      }

      var value = null;

      /* Choice step: look up branches_chosen, then fetch field from chosen branch */
      if (step.kind === 'choice' && step.branches) {
        var chosenKey = state.branches_chosen[stepId];
        if (chosenKey && step.branches[chosenKey]) {
          value = step.branches[chosenKey][fieldName];
        }
      }
      /* Quiz step: look up quiz_answers, then fetch field from options */
      else if (step.kind === 'quiz' && step.options) {
        var optionLetter = state.quiz_answers[stepId];
        if (optionLetter) {
          var opts = step.options;
          /* Support both array and object formats */
          if (Array.isArray(opts)) {
            /* Find option by letter property */
            for (var o = 0; o < opts.length; o++) {
              if (opts[o].letter === optionLetter) {
                value = opts[o][fieldName];
                break;
              }
            }
          } else {
            /* Object keyed by letter */
            value = opts[optionLetter] ? opts[optionLetter][fieldName] : null;
          }
        }
      }
      /* Side quest step: check completion status */
      else if (step.kind === 'side_quest') {
        /* For side quest steps, check if referenced step is completed */
        value = state.side_quests_completed.indexOf(stepId) !== -1 ? 'completed' : 'pending';
      }
      /* Narrative / section_header / checkpoint / code_editor: read field directly from step */
      else {
        value = step[fieldName];
      }

      /* Field exists but is empty/unresolved — not a broken ref */
      if (value === undefined || value === null || value === '') {
        return EMPTY_VALUE;
      }

      return String(value);
    }

    /* ── Internal: format broken-ref placeholder ──────────────────── */
    function _formatPlaceholder(fieldPath) {
      return '[broken ref: ' + fieldPath + ']';
    }

    /* ── Internal: clean up whitespace after expression removal ───── */
    function _cleanupWhitespace(text) {
      /* Collapse multiple spaces into one */
      text = text.replace(/  +/g, ' ');
      /* Trim leading/trailing spaces */
      text = text.trim();
      return text;
    }

    /* ── Public: resolve all [[expr]] / ${{expr}} patterns in plain text ─ */
    function resolve(text, useCaseId) {
      if (!text) {
        return '';
      }

      var state = UCEState.getState(useCaseId);
      var stepMap = _buildStepMap(steps);
      var matches = _extractExpressions(text);

      if (matches.length === 0) {
        return text;
      }

      var result = text;

      /* Process matches in reverse order to preserve indices */
      for (var i = matches.length - 1; i >= 0; i--) {
        var match = matches[i];
        var parts = match.expression.split('.');
        var refStepId = parts[0];
        var fieldName = parts.length > 1 ? parts[1] : null;

        /* Handle leading dot: .stepId.field -> stepId.field */
        if (refStepId === '' && parts.length >= 2) {
          refStepId = parts[1];
          fieldName = parts.length > 2 ? parts[2] : null;
        }

        /* No field specified — skip (no default support yet, Phase 2) */
        if (!fieldName) {
          continue;
        }

        var resolved = _resolveField(refStepId, fieldName, stepMap, state);

        if (resolved === null) {
          /* Broken ref: step ID or field name not found */
          result = result.replace(match.full, _formatPlaceholder(match.expression));
        } else if (resolved === EMPTY_VALUE) {
          /* Valid expression but value not yet set — remove silently */
          result = result.replace(match.full, '');
        } else {
          /* Replace with resolved value */
          result = result.replace(match.full, resolved);
        }
      }

      return _cleanupWhitespace(result);
    }

    /* ── Internal: resolve text inside a single text node ─────────── */
    function _resolveTextNode(textNode) {
      var originalText = textNode.nodeValue;
      if (!originalText) {
        return;
      }

      var state = UCEState.getState(useCaseId);
      var stepMap = _buildStepMap(steps);
      var matches = _extractExpressions(originalText);

      if (matches.length === 0) {
        return;
      }

      var resolvedText = originalText;

      /* Process matches in reverse order to preserve indices */
      for (var i = matches.length - 1; i >= 0; i--) {
        var match = matches[i];
        var parts = match.expression.split('.');
        var refStepId = parts[0];
        var fieldName = parts.length > 1 ? parts[1] : null;

        /* Handle leading dot: .stepId.field -> stepId.field */
        if (refStepId === '' && parts.length >= 2) {
          refStepId = parts[1];
          fieldName = parts.length > 2 ? parts[2] : null;
        }

        if (!fieldName) {
          continue;
        }

        var resolved = _resolveField(refStepId, fieldName, stepMap, state);

        if (resolved === null) {
          /* Broken ref: step ID or field name not found */
          resolvedText = resolvedText.replace(match.full, _formatPlaceholder(match.expression));
        } else if (resolved === EMPTY_VALUE) {
          /* Valid expression but value not yet set — remove silently */
          resolvedText = resolvedText.replace(match.full, '');
        } else {
          resolvedText = resolvedText.replace(match.full, resolved);
        }
      }

      resolvedText = _cleanupWhitespace(resolvedText);

      /* Only update if something changed */
      if (resolvedText !== originalText) {
        textNode.nodeValue = resolvedText;
      }
    }

    /* ── Public: recursively resolve text inside any DOM element ──── */
    function resolveNode(element) {
      if (!element) {
        return;
      }

      /* Skip script and style elements — never touch their content */
      if (element.tagName === 'SCRIPT' || element.tagName === 'STYLE') {
        return;
      }

      var children = element.childNodes;
      for (var i = 0; i < children.length; i++) {
        var child = children[i];
        if (child.nodeType === Node.TEXT_NODE) {
          _resolveTextNode(child);
        } else if (child.nodeType === Node.ELEMENT_NODE) {
          resolveNode(child);
        }
      }
    }

    /* ── Public: re-scan all rendered step elements and re-resolve ── */
    function refreshAllSteps() {
      var container = document.getElementById(STEPS_CONTAINER_ID);
      if (!container) {
        return;
      }

      var stepElements = container.querySelectorAll('.uce-flow-step');
      for (var i = 0; i < stepElements.length; i++) {
        resolveNode(stepElements[i]);
      }

      /* Emit event for other modules to react */
      emitEvent('expressions-refreshed', { count: stepElements.length });
    }

    /* ── Export public API ────────────────────────────────────────── */
    return {
      resolve: resolve,
      resolveNode: resolveNode,
      refreshAllSteps: refreshAllSteps
    };
  })();

  /* Expose on window for external consumption */
  window.ExpressionResolver = ExpressionResolver;

})();

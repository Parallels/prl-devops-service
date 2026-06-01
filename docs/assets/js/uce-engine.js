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
  var visibleSteps = [];
  var state = null;
  var currentIndex = 0;

  /* ── DOM refs (lazily cached) ──────────────────────────────────── */
  var dom = {};

  function cacheDom() {
    dom.container = document.getElementById(STEPS_CONTAINER_ID);
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
      initState.active_step = visibleSteps[currentIndex].id;
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


    // Bind navigation events
    bindNavigation();
    bindPanelEvents();

    // Check for return-to-origin (side quest completion redirect)
    checkReturnToOrigin();

    // Emit ready event
    emitEvent('engine-ready', { useCaseId: useCaseId, totalSteps: steps.length });
  }

  /**
   * Find the index of the current step.
   * Priority: last completed > first step.
   * @param {Array} stepsArr — optional step array (defaults to visibleSteps)
   */
  function findCurrentIndex(stepsArr) {
    var arr = stepsArr || visibleSteps;
    // Check if any visible step is completed
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
      default:
        bodyEl = renderNarrative(step);
    }

    if (bodyEl) {
      el.appendChild(bodyEl);
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

    // Determine if we have a panel
    var panelEl = step.panel ? renderPanel(step.panel) : null;
    var hasPanel = !!panelEl;
    var hasBody = !!step.body;

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

    // Left/top column — narrative body + side quests
    if (hasBody || (step.side_quests && step.side_quests.length > 0)) {
      var leftCol = document.createElement('div');
      leftCol.className = 'uce-narrative-left';

      // Body content
      if (hasBody) {
        var bodyEl = document.createElement('div');
        bodyEl.className = 'uce-narrative-body';
        // For api_call panels, enhance the body with step tag and title
        if (step.panel && step.panel.kind === 'api_call') {
          var panelCfg = step.panel;
          var panelTitle = panelCfg.title || step.title || '';
          var stepTag = step.tag || '';
          
          var tagHtml = stepTag
            ? '<span class="uce-step-tag">' + escapeHtml(stepTag) + '</span>'
            : '';
          
          bodyEl.innerHTML =
            '<div class="uce-api-call-header">' +
            tagHtml +
            '<h3 class="uce-api-call-title">' + escapeHtml(panelTitle) + '</h3>' +
            '</div>' +
            step.body;
        } else {
          bodyEl.innerHTML = step.body;
        }
        leftCol.appendChild(bodyEl);
      }

      // Side quest triggers (rendered inside left column, below body)
      if (step.side_quests && step.side_quests.length > 0) {
        var sqTriggers = renderSideQuestTriggers(step.side_quests, step.id);
        for (var k = 0; k < sqTriggers.length; k++) {
          leftCol.appendChild(sqTriggers[k]);
        }
      }

      // Step-level checklist — renders after body and side quests
      if (step.checklist && step.checklist.length > 0) {
        var checklistUl = document.createElement('ul');
        checklistUl.className = 'uce-checklist';
        for (var i = 0; i < step.checklist.length; i++) {
          var li = document.createElement('li');
          li.className = 'uce-checklist-item';
          // Render as markdown (supports bold, code, links, etc.)
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
        leftCol.appendChild(checklistUl);
      }

      grid.appendChild(leftCol);
    }

    // Right/bottom column — panel (if present)
    if (hasPanel) {
      var rightCol = document.createElement('div');
      rightCol.className = 'uce-narrative-right';

      // Expand toggle button (only in columns mode with panel)
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

      var panel = panelEl;
      if (panel) {
        rightCol.appendChild(panel);
      }
      grid.appendChild(rightCol);
    }

    div.appendChild(grid);

    return div;
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
        var termCmd = escapeHtml(panel.cmd || '');
        var showCopy = panel.copy_button !== false;
        var termOutputHtml = '';
        if (panel.output) {
          termOutputHtml = escapeHtml(panel.output);
        }
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
            '<pre class="mac-window-body"><code>' + termCmd + '</code></pre>' +
          '</div>';
        var outputWindow = termOutputHtml
          ? '<div class="mac-window">' +
              '<div class="mac-window-header">' +
                '<span class="mac-dot mac-dot--red"></span>' +
                '<span class="mac-dot mac-dot--yellow"></span>' +
                '<span class="mac-dot mac-dot--green"></span>' +
                '<span class="mac-window-title">Output</span>' +
              '</div>' +
              '<pre class="mac-window-body"><code>' + termOutputHtml + '</code></pre>' +
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
    var html = '<div class="tw-container" id="tw-container">' +
      '<div class="tw-stage">' +
        '<div class="tw-images">' +
          '<img class="tw-img-main" src="' + window.__UCE_BASEURL__ + '/assets/img/ui/' + escapeHtml(mainImage) + '" alt="' + escapeHtml(tw.title) + ' \u2014 default view" />' +
        '</div>' +
        '<div class="tw-hotspot-layer"></div>' +
        '<div class="tw-scene-badge" id="tw-scene-badge">' +
          '<span class="tw-scene-badge-label"></span>' +
          '<button class="tw-scene-badge-close" id="tw-scene-badge-close" aria-label="Exit scene">\u00d7</button>' +
        '</div>' +
      '</div>' +
      '<div class="tw-info-panel" id="tw-info-panel">' +
        '<button class="tw-info-close" id="tw-info-close" aria-label="Close">\u2715</button>' +
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

      var container = panelEl.querySelector('#tw-container');
      if (!container) return;

      var stage = container.querySelector('.tw-stage');
      var imgMain = container.querySelector('.tw-img-main');
      var hotspotLayer = container.querySelector('.tw-hotspot-layer');
      var infoPanel = container.querySelector('.tw-info-panel');
      var infoStepLabel = container.querySelector('.tw-info-step-label');
      var infoTitle = container.querySelector('.tw-info-title');
      var infoDesc = container.querySelector('.tw-info-desc');
      var dotsContainer = container.querySelector('.tw-dots');
      var btnPrev = container.querySelector('.tw-nav-prev');
      var btnNext = container.querySelector('.tw-nav-next');
      var sceneBadge = container.querySelector('.tw-scene-badge');
      var sceneBadgeLabel = container.querySelector('.tw-scene-badge-label');
      var sceneBadgeClose = container.querySelector('#tw-scene-badge-close');
      var infoClose = container.querySelector('#tw-info-close');

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
        var newSrc = window.__UCE_BASEURL__ + '/assets/img/ui/' + sceneImage;
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
        var parentSrc = window.__UCE_BASEURL__ + '/assets/img/ui/' + tw.main_image;
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

      if (isSelected) {
        // Selected: wrap badge + reset in a flex row
        html += '<div class="uce-branch-actions">' +
          '<span class="uce-branch-badge"><span class="uce-badge uce-badge--selected">SELECTED</span></span>' +
          '<button class="uce-branch-reset" data-uce-step="' + escapeHtml(stepId) + '" aria-label="Reset choice">&times; Reset</button>' +
          '</div>';
      } else if (!isLocked) {
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
        from: useCaseId,
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
      if (info.from === useCaseId) {
        // Remove the session storage
        sessionStorage.removeItem('uce_return_to');
        showReturnBanner(info);
      }
    } catch (e) {
      try { sessionStorage.removeItem('uce_return_to'); } catch(ex) { /* noop */ }
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
    qEl.textContent = question;
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
        fbEl.textContent = (userWasCorrect ? '\u2713 ' : '\u2717 ') + (userWasCorrect ? feedbackCorrect : feedbackIncorrect);
      }
    }

    return div;
  }

  function showFeedback(container, message, correct) {
    var fbEl = container.nextElementSibling;
    if (!fbEl) return;
    fbEl.className = 'uce-quiz-feedback uce-quiz-feedback--' + (correct ? 'correct' : 'incorrect');
    fbEl.textContent = (correct ? '\u2713 ' : '\u2717 ') + message;
  }

  /* - Code Editor (Monaco) - */

  function renderCodeEditorStep(step) {
    console.log('[UCE] renderCodeEditorStep called:', step.id, 'lang:', step.language);
    var stepId = step.id || '';
    var language = step.language || 'text';
    var initialCode = step.initial_code || '';
    var validationCmd = step.validation_cmd || '';
    var validationFn = step.validation_js || '';
    var saveKey = 'uce_code:' + useCaseId + ':' + stepId;

    var savedCode = '';
    try { savedCode = localStorage.getItem(saveKey) || ''; } catch(e) {}
    var codeContent = savedCode || initialCode;

    var div = document.createElement('div');
    div.className = 'uce-step-body uce-step-body--code-editor';
    div.setAttribute('data-uce-step-id', stepId);

    if (step.body) {
      var bodyP = document.createElement('p');
      bodyP.className = 'uce-editor-body';
      bodyP.innerHTML = step.body;
      div.appendChild(bodyP);
    }

    var editorContainer = document.createElement('div');
    editorContainer.className = 'uce-editor-container';
    editorContainer.setAttribute('data-uce-editor-step', stepId);

    var label = document.createElement('label');
    label.className = 'uce-editor-label';
    label.textContent = language.toUpperCase();
    editorContainer.appendChild(label);

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
    div.appendChild(editorContainer);

    // Initialize Monaco
    (function() {
      var lang = language;
      if (lang === 'yml') lang = 'yaml';
      if (lang === 'sh' || lang === 'bash') lang = 'shell';
      if (lang === 'ts') lang = 'typescript';
      if (lang === 'py') lang = 'python';

      console.log('[UCE] Monaco init, stepId:', stepId, 'lang:', lang, 'require:', typeof require);

      if (typeof require !== 'function') {
      console.error('[UCE] require is not available');
      loadInd.style.display = '';
      loadInd.style.display = '';
      loadInd.textContent = 'Monaco unavailable';
      loadInd.style.color = '#EF4444';
        return;
        }

      require.config({ paths: { vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs' } });

      require(['vs/editor/editor.main'], function() {
        var mon = window.monaco;
        if (!mon) { console.error('[UCE] monaco not found'); loadInd.style.display = ''; loadInd.textContent = 'Monaco not found'; loadInd.style.color = '#EF4444'; return; }

        try {
          var editorInstance = mon.editor.create(editorDiv, {
            value: codeContent,
            language: lang,
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
          window.__UCE_MONACO_INSTANCES__[stepId] = {
            getValue: function() { return editorInstance.getValue(); },
            setValue: function(v) { editorInstance.setValue(v); },
            getEditor: function() { return editorInstance; }
          };

          var ro = new ResizeObserver(function() {
            editorInstance.layout();
          });
          ro.observe(editorDiv);
          console.log('[UCE] Monaco editor created for:', stepId);
        } catch(e) {
          console.error('[UCE] Monaco create failed:', e);
          loadInd.style.display = '';
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
    })();

    return div;
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
        sessionStorage.removeItem('uce_return_to');

        // Mark side quest as completed
        if (info.sq_id) {
          UCEState.completeSideQuest(useCaseId, info.sq_id);
          UCEState.save();
        }

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
    var total = steps.length;
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
})();

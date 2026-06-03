/**
 * uce-state.js — State Manager for the Use Case Engine
 *
 * Manages user progress in localStorage with a clean API surface.
 * Key format: uce_progress:<use_case_id>
 *
 * Usage:
 *   var S = UCEState.getState('my-use-case');
 *   S.completeStep('step-id');
 *   UCEState.save();
 */
var UCEState = (function () {
  'use strict';

  var STORAGE_PREFIX = 'uce_progress:';
  var SAVE_DEBOUNCE_MS = 100;

  /* ── Internal State ─────────────────────────────────────────────── */
  var _currentState = {};  // { useCaseId: stateObject }
  var _dirty = false;
  var _saveTimer = null;

  /* ── Helpers ────────────────────────────────────────────────────── */

  /** Generate the localStorage key for a use case. */
  function _key(id) {
    return STORAGE_PREFIX + id;
  }

  /* ── Condition Evaluator ────────────────────────────────────────── */

  /**
   * Evaluate a single condition like "choose-cloud.label == 'aws'".
   * Supported operators: ==, !=, =~, !~
   * Returns boolean.
   */
  function _evalCondition(condition, stepMap, state) {
    if (!condition) return true;

    /* Parse: stepId.field operator 'value' */
    var m = String(condition).match(/^([^.]+)\.([^=]+?)\s*(==|!=|=~|!~)\s*['"](.*)['"]\s*$/);
    if (!m) {
      /* Try without quotes: stepId.field == value */
      m = String(condition).match(/^([^.]+)\.([^=]+?)\s*(==|!=|=~|!~)\s*(.+)$/);
      if (!m) return true; /* Unparseable — assume true */
    }

    var refStepId = m[1];
    var refField = m[2].trim();
    var op = m[3];
    var target = m[4].trim();

    var step = stepMap[refStepId];
    if (!step) return false;

    var actual = _resolveRef(refStepId, refField, step, state);
    if (actual === null) return false;

    actual = String(actual);
    target = String(target);

    switch (op) {
      case '==': return actual === target;
      case '!=': return actual !== target;
      case '=~': return new RegExp(target).test(actual);
      case '!~': return !new RegExp(target).test(actual);
      default: return false;
    }
  }

  /**
   * Resolve a reference like "choose-cloud.label" against state.
   */
  function _resolveRef(stepId, field, step, state) {
    if (step.kind === 'choice' && step.branches) {
      var chosen = state.branches_chosen[stepId];
      if (chosen && step.branches[chosen]) {
        return step.branches[chosen][field];
      }
    } else if (step.kind === 'quiz' && step.options) {
      var letter = state.quiz_answers[stepId];
      if (letter) {
        var opts = step.options;
        if (Array.isArray(opts)) {
          for (var o = 0; o < opts.length; o++) {
            if (opts[o].letter === letter) return opts[o][field];
          }
        } else {
          return opts[letter] ? opts[letter][field] : null;
        }
      }
    } else if (step.kind === 'side_quest') {
      return state.side_quests_completed.indexOf(stepId) !== -1 ? 'completed' : 'pending';
    } else {
      return step[field] !== undefined ? step[field] : null;
    }
    return null;
  }

  /**
   * Evaluate an if-condition string (supports && and ||).
   */
  function _evaluateCondition(condition, stepMap, state) {
    if (!condition) return true;
    /* Split on || first (lowest precedence) */
    var orParts = String(condition).split(/\s*\|\|\s*/);
    for (var oi = 0; oi < orParts.length; oi++) {
      var orGroup = orParts[oi];
      /* Split on && */
      var andParts = orGroup.split(/\s*&\&\s*/);
      var allTrue = true;
      for (var ai = 0; ai < andParts.length; ai++) {
        var part = andParts[ai].trim();
        /* Strip negation */
        var negated = false;
        if (part.indexOf('!') === 0) {
          negated = true;
          part = part.substring(1).trim();
        }
        var result = _evalCondition(part, stepMap, state);
        if (negated) result = !result;
        if (!result) { allTrue = false; break; }
      }
      if (allTrue) return true;
    }
    return false;
  }

  /** Load a single use case's state from localStorage. */
  function _loadFromStorage(id) {
    try {
      var raw = localStorage.getItem(_key(id));
      if (raw) {
        return JSON.parse(raw);
      }
    } catch (e) {
      // Corrupted data — start fresh
      console.warn('[UCEState] Failed to parse state for', id, e);
    }
    return null;
  }

  /** Write the current state to localStorage (debounced). */
  function _persist() {
    if (!_dirty) return;
    clearTimeout(_saveTimer);
    _saveTimer = setTimeout(function () {
      for (var id in _currentState) {
        if (_currentState.hasOwnProperty(id)) {
          try {
            localStorage.setItem(_key(id), JSON.stringify(_currentState[id]));
          } catch (e) {
            console.error('[UCEState] Failed to save state for', id, e);
          }
        }
      }
      _dirty = false;
    }, SAVE_DEBOUNCE_MS);
  }

  /**
   * Force-synchronously persist all dirty states to localStorage.
   * Used before navigation to avoid losing debounced writes.
   */
  function _forceSync() {
    for (var id in _currentState) {
      if (_currentState.hasOwnProperty(id) && _currentState[id]) {
        try {
          localStorage.setItem(_key(id), JSON.stringify(_currentState[id]));
        } catch (e) {
          console.error('[UCEState] Failed to force-sync state for', id, e);
        }
      }
    }
    _dirty = false;
  }

  /** Create a fresh default state object. */
  function _freshState() {
    return {
      current_step: null,
      active_step: null,
      completed_steps: [],
      branches_chosen: {},
      quiz_answers: {},
      side_quests_completed: [],
      side_quest_tags: {},
      updated_at: null
    };
  }

  /* ── Public API ─────────────────────────────────────────────────── */

  /**
   * Force-synchronously persist all dirty state to localStorage.
   * Use this before page navigation to avoid losing unsaved state.
   */
  function _forceSync() {
    for (var id in _currentState) {
      if (_currentState.hasOwnProperty(id)) {
        try {
          localStorage.setItem(_key(id), JSON.stringify(_currentState[id]));
        } catch (e) {
          console.error('[UCEState] Failed to force-sync state for', id, e);
        }
      }
    }
    _dirty = false;
  }

  /**
   * Get (and lazily-load) state for a use case.
   * @param {string} useCaseId
   * @returns {object} State object (always non-null).
   */
  function getState(useCaseId) {
    if (!_currentState[useCaseId]) {
      var stored = _loadFromStorage(useCaseId);
      _currentState[useCaseId] = stored || _freshState();
    }
    return _currentState[useCaseId];
  }

  /**
   * Force-save all dirty state to localStorage.
   */
  function save() {
    _dirty = true;
    _persist();
  }

  /**
   * Mark a step as completed.
   * @param {string} useCaseId
   * @param {string} stepId
   */
  function completeStep(useCaseId, stepId) {
    var state = getState(useCaseId);
    if (state.completed_steps.indexOf(stepId) === -1) {
      state.completed_steps.push(stepId);
    }
    state.current_step = stepId;
    state.updated_at = new Date().toISOString();
    _dirty = true;
    _persist();
  }

  /**
   * Record a branch choice.
   * @param {string} useCaseId
   * @param {string} choiceStepId
   * @param {string} branchKey — the key of the chosen branch (e.g., 'aws')
   */
  function chooseBranch(useCaseId, choiceStepId, branchKey) {
    var state = getState(useCaseId);
    state.branches_chosen[choiceStepId] = branchKey;
    state.updated_at = new Date().toISOString();
    _dirty = true;
    _persist();
  }

  /**
   * Record a quiz answer (option letter).
   * @param {string} useCaseId
   * @param {string} quizStepId
   * @param {string} optionLetter — e.g. 'A', 'B'
   */
  function recordQuizAnswer(useCaseId, quizStepId, optionLetter) {
    var state = getState(useCaseId);
    state.quiz_answers[quizStepId] = optionLetter;
    state.updated_at = new Date().toISOString();
    _dirty = true;
    _persist();
  }

  /**
   * Mark a side quest as completed.
   * @param {string} useCaseId
   * @param {string} sqId
   */
  function completeSideQuest(useCaseId, sqId) {
    var state = getState(useCaseId);
    if (state.side_quests_completed.indexOf(sqId) === -1) {
      state.side_quests_completed.push(sqId);
    }
    state.updated_at = new Date().toISOString();
    _dirty = true;
    _persist();
  }

  /**
   * Tag a side quest (e.g., 'taken', 'skipped').
   * @param {string} useCaseId
   * @param {string} sqId
   * @param {string} status — 'taken' | 'skipped'
   */
  function tagSideQuest(useCaseId, sqId, status) {
    var state = getState(useCaseId);
    if (!state.side_quest_tags[sqId]) {
      state.side_quest_tags[sqId] = { status: status, view_count: 0 };
    } else {
      state.side_quest_tags[sqId].status = status;
      state.side_quest_tags[sqId].view_count++;
    }
    state.updated_at = new Date().toISOString();
    _dirty = true;
    _persist();
  }

  /**
   * Reset all state for a use case.
   * @param {string} useCaseId
   */
  function resetState(useCaseId) {
    _currentState[useCaseId] = _freshState();
    try {
      localStorage.removeItem(_key(useCaseId));
    } catch (e) { /* noop */ }
  }

  /**
   * Check if a step is completed.
   * @param {string} useCaseId
   * @param {string} stepId
   * @returns {boolean}
   */
  function isStepCompleted(useCaseId, stepId) {
    var state = getState(useCaseId);
    return state.completed_steps.indexOf(stepId) !== -1;
  }

  /**
   * Check if a side quest has been taken.
   * @param {string} useCaseId
   * @param {string} sqId
   * @returns {boolean}
   */
  function isSideQuestCompleted(useCaseId, sqId) {
    var state = getState(useCaseId);
    return state.side_quests_completed.indexOf(sqId) !== -1;
  }

  /**
   * Get the tag for a side quest.
   * @param {string} useCaseId
   * @param {string} sqId
   * @returns {object|null} { status, view_count } or null
   */
  function getSideQuestTag(useCaseId, sqId) {
    var state = getState(useCaseId);
    return state.side_quest_tags[sqId] || null;
  }

  /**
   * Filter step list to only show reachable steps based on branch choices.
   * Steps behind unchosen branches are hidden.
   *
   * Algorithm:
   *   1. Collect all chosen branch mappings: { choiceStepId: branchKey }
   *   2. Walk the steps in order.
   *   3. When we hit a choice step, determine which branch was chosen.
   *   4. Include all steps up to and including the choice.
   *   5. After the choice, only include steps that are reachable via the chosen branch's `next`.
   *      Steps that are only reachable via unchosen branches are skipped.
   *   6. After all choices are resolved, include all remaining steps.
   *
   * @param {string} useCaseId
   * @param {Array<object>} steps — full step list from YAML
   * @returns {Array<object>} filtered reachable steps
   */
  function getVisibleSteps(useCaseId, steps) {
    var state = getState(useCaseId);
    var branchesChosen = state.branches_chosen;
    var visible = [];
    var pastAllChoices = false;

    // Build a lookup: stepId -> step
    var stepMap = {};
    for (var i = 0; i < steps.length; i++) {
      stepMap[steps[i].id] = steps[i];
    }

    // Collect choice step IDs that have been answered
    var choiceStepIds = Object.keys(branchesChosen);

    for (var i = 0; i < steps.length; i++) {
      var step = steps[i];

      /* ── if-condition gate ─────────────────────────────────── */
      if (step.if && !_evaluateCondition(step.if, stepMap, state)) {
        continue;
      }

      // Section headers and narrative steps are always visible
      if (step.kind === 'section_header') {
        visible.push(step);
        continue;
      }

      // Quiz steps are always visible (they don't branch)
      if (step.kind === 'quiz') {
        visible.push(step);
        continue;
      }

      // Code editor steps are always visible
      if (step.kind === 'code_editor') {
        visible.push(step);
        continue;
      }

      // Choice step
      if (step.kind === 'choice') {
        visible.push(step);
        // Once we see a choice, mark that we're past the branching point
        pastAllChoices = true;
        continue;
      }

      // Checkpoint and narrative steps: check reachability
      if (pastAllChoices && choiceStepIds.length > 0) {
        // A step is reachable if:
        //   a) Its `next` field leads to it from a chosen branch, OR
        //   b) It comes after the last choice in the original step order, OR
        //   c) It's referenced as `returns_to` from a completed side quest
        var isReachable = true;

        // Check if this step is the `next` of a chosen branch
        var isDirectReachable = false;
        for (var ci = 0; ci < choiceStepIds.length; ci++) {
          var csid = choiceStepIds[ci];
          var branchKey = branchesChosen[csid];
          var choiceStep = stepMap[csid];
          if (choiceStep && choiceStep.branches && choiceStep.branches[branchKey]) {
            var nextStepId = choiceStep.branches[branchKey].next;
            // Walk forward from the chosen branch's next to see if this step is encountered
            if (isStepAfterId(steps, nextStepId, step.id)) {
              isDirectReachable = true;
              break;
            }
          }
        }

        // Also check if this step is after all choices in original order
        var lastChoiceIdx = -1;
        for (var si = 0; si < steps.length; si++) {
          if (steps[si].kind === 'choice' && branchesChosen[steps[si].id]) {
            lastChoiceIdx = si;
          }
        }
        var afterLastChoice = i > lastChoiceIdx;

        // Or it's the returns_to of a completed side quest
        var isReturnsTo = false;
        for (var qi = 0; qi < state.side_quests_completed.length; qi++) {
          var sqId = state.side_quests_completed[qi];
          // We can't resolve this here without side_quests data, so skip
        }

        if (!isDirectReachable && !afterLastChoice) {
          isReachable = false;
        }

        if (!isReachable) continue;
      }

      visible.push(step);
    }

    return visible;
  }

  /**
   * Check if stepWithId appears after stepAfterId in the steps array.
   */
  function isStepAfterId(steps, stepAfterId, stepWithId) {
    var afterIdx = -1;
    var withIdx = -1;
    for (var i = 0; i < steps.length; i++) {
      if (steps[i].id === stepAfterId) afterIdx = i;
      if (steps[i].id === stepWithId) withIdx = i;
    }
    return withIdx > afterIdx;
  }

  /**
   * Get the total number of steps (including hidden ones).
   * @param {string} useCaseId
   * @param {Array<object>} steps
   * @returns {number}
   */
  function getTotalSteps(useCaseId, steps) {
    return steps.length;
  }

  /**
   * Get the number of completed steps.
   * @param {string} useCaseId
   * @returns {number}
   */
  function getCompletedCount(useCaseId) {
    var state = getState(useCaseId);
    return state.completed_steps.length;
  }

  /**
   * Check if a side quest's linked use case is fully completed.
   * Side quests point to other use cases via `links_to`. We check
   * localStorage for that use case's progress and verify current_step
   * is '__complete__'.
   * @param {string} sqLinksTo — the use-case ID that the side quest points to
   * @returns {boolean}
   */
  function isSideQuestUseCaseCompleted(sqLinksTo) {
    if (!sqLinksTo) return false;
    try {
      var raw = localStorage.getItem('uce_progress:' + sqLinksTo);
      if (!raw) return false;
      var state = JSON.parse(raw);
      return state && state.current_step === '__complete__';
    } catch (e) {
      return false;
    }
  }

  /* ── Export ─────────────────────────────────────────────────────── */
  return {
    getState: getState,
    save: save,
    completeStep: completeStep,
    chooseBranch: chooseBranch,
    completeSideQuest: completeSideQuest,
    tagSideQuest: tagSideQuest,
    resetState: resetState,
    isStepCompleted: isStepCompleted,
    isSideQuestCompleted: isSideQuestCompleted,
    isSideQuestUseCaseCompleted: isSideQuestUseCaseCompleted,
    getSideQuestTag: getSideQuestTag,
    getVisibleSteps: getVisibleSteps,
    getTotalSteps: getTotalSteps,
    getCompletedCount: getCompletedCount,
    recordQuizAnswer: recordQuizAnswer,
    _evaluateCondition: _evaluateCondition,
    _forceSync: _forceSync
  };
})();

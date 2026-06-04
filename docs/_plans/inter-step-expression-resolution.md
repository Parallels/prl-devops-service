# Plan: Inter-Step Expression Resolution


## Overview

Enable steps to reference data from other steps using an expression syntax using mustache/double-brace style (two opening/closing braces surrounding dot-notation paths). When an earlier interactive step (choice, quiz, side quest) is completed, subsequent steps automatically re-render with resolved values substituted into their text fields — titles, descriptions, body text, choice labels, quiz options, API panel titles, side quest names, and every text node in the rendered DOM.

**User Story:** As a use case author, I want to write expressions anywhere in step text so that once the user picks a branch, all downstream steps dynamically display the correct provider name (e.g., "AWS", "google-cli") without hardcoding multiple versions of each step for each branch.


---


## Decisions Made (Already Answered by User)


### D1: Expression Syntax -- Double Brace Style

- **Mustache-style double-curly-brackets.** Flat key paths only (no nested variable refs). Dot notation for deep access.
- Leading dot (optional): `.stepId.field` or `stepId.field`.
- Example evaluates to `"AWS"` when AWS branch selected, empty if unresolved.
- No support for nested variable references (e.g. no `$dynamicKey` inside another expr).


### D2: Branch Auto-Resolution

When user picks a branch on a **choice** step, the parent step ID resolves to the FULL branch option object:

```yaml
- id: choose-cloud
  kind: choice
  branches:
    aws:
      label: AWS
      value: aws-cli
      description: EC2 + CodePipeline path
      icon: fa-aws
```

After selecting AWS:
- Step field `label` → `"AWS"`
- Step field `value` → `"aws-cli"`
- Step field `description` → `"EC2 + CodePipeline path"`

Same behavior for quiz steps: after answering, full chosen option object resolves under the step ID. Side quests resolve their metadata similarly.


### D3: Hide Unresolved Expressions

If an expression references data not yet completed (user not picked a branch), the expression is silently removed from output. Surrounding static text preserved. Clean up leading/trailing whitespace.

Example title:
```
Original title: [expr placeholder] Infrastructure Setup
Before selection → "Infrastructure Setup"
After selection (AWS) → "AWS Infrastructure Setup"
```


### D4: Error Handling -- Visible Placeholder

Typo like `placeholder` shown visibly in rendered HTML so developers catch typos immediately. Format: `[broken ref: original_expr_content]` styled red/orange warning styling. Console.warn logs detailed error for debugging.


### D5: Apply EVERYWHERE

Expression resolution applies to ALL text-containing fields and elements:

| Scope | Fields / Elements Resolved |
|-------|--------------------------|
| Step header | Step title (always) |
| Step body | All HTML/text inside .uce-panel-body |
| Choice cards | Each card's label and description field |
| Quiz options | Each option's text content within choice cards |
| Section headers | Title text of section-header type steps |
| API call panels | Window title text, terminal command text |
| Side quests | Question/title text in side quest cards |
| Any text node | Fallback scanner on leaf text nodes (catches edge cases) |


---


## Architecture


### A. New Module -- expressionResolver

Add new module inside uce-engine.js with these responsibilities:

Core public API:
- `resolve(text, useCaseId)` -- Resolve all expressions in plain text, return resolved string.
- `resolveNode(element, useCaseId)` -- Recursively resolve text inside any DOM element.
- `refreshAllSteps()` -- Re-scan all currently rendered step elements and re-resolve them. Primary hook for reactive updates.
- `refreshSingleStep(stepIndex)` -- Targeted refresh for one specific step by index (optional optimization later).

Internal helpers:
- `_extractExpressions(text)` -- Find all expressions using regex. Pattern matches two-bracket-delimited content (non-greedy).
- `_resolveField(stepId, fieldName, scope)` -- Traverse state/data to get leaf value.
- `_formatPlaceholder(fieldPath, errorMsg)` -- Format broken-ref visible text.


#### Method Specs


##### resolve(text, useCaseId) -> String

Input: raw string like `Install [[providerValue]] cli on Ubuntu`.  
Output: resolved string like `Install aws-cli on Ubuntu` or `install on ubuntu` (if unresolved).

Implementation:
1. Regex find all pattern matches (non-greedy).
2. For each match, strip leading dots/whitespace. Split by dot into segments. First segment is stepId, rest is field path.
3. Fetch referenced data from UCEState state via step.id lookup in the steps array.
4. Traverse returned object with field path to get leaf value.
5. If missing -> return `[broken ref: original_expr_content]` and log console.warn.
6. Replace expression in original text with resolved value. Return fully substituted string.


##### resolveNode(element, useCaseId) -> void

Walks DOM subtree from element, finding all text nodes, resolving them. Skips script/style/embedded tags. Only processes visible leaf text nodes.

Important constraints: Must handle HTML `<strong>`, `<code>`, `<em>` correctly -- resolve around tag structure, not disrupt child hierarchy. E.g., `Install [[cliValue]] cli` inside `<strong>` block should work without breaking HTML wrapper.


##### refreshAllSteps() -> void

Called after any answer/change event. Re-scans all currently rendered step elements and re-resolves all expressions. Main trigger for reactive rendering flow.


##### refreshSingleStep(stepIndex) -> void

Selective refresh for one specific step by index. Used sparingly for performance gains when only isolated sections change.


### B. Field Resolution Logic

Given expression referencing `choose-cloud.label`:
1. Parse stepId = `"choose-cloud"`, path = `["label"]`.
2. Look up `state.branches_chosen.choose-cloud` -> gets branch key (e.g., `"aws"`).
3. From the original YAML data, fetch `steps.find(s => s.id === "choose-cloud").branches.aws`.
4. Traverse returned object with path -> returns `"label"` field = `"AWS"`.
5. Convert result to string.

For a quiz example `quiz-pipeline.selected`:
1. Parse stepId = `"quiz-pipeline"`, path = `["selected"]`.
2. Look up `state.quiz_answers.quiz-pipeline` -> e.g., `"A"` (letter index).
3. Map letter back to actual option object from YAML config.
4. Return requested field value.


### C. State Tracking

New state fields needed in uce-state.js structure:

```javascript
{
  expression_cache: {
    [useCaseId]: {
      [stepRef]: resolvedValue
    }
  },
  resolved_steps: {
    [useCaseId]: map<stepId,string,renderedDOMString>
  }
}
```

Caching prevents redundant parsing across re-renders. Cache invalidation depends on whether source step changed (track dependency mapping during init).


### D. Integration Points


#### init phase (uce-engine.js, renderFlow before bindNavEvents):
- After YAML loaded, inject resolved values into each step's title/description/body.
- Single pass to populate initial expressionCache.
- Ensures mid-flow reload (with prior state restored) shows properly resolved values.


#### Choice Selection Path (handleChoiceSelection / completeStep functions):
In existing _handleChoiceSelection() method where completeStep is called:
1. Update UCEState.branches_chosen with new value.
2. Call expressionCache.invalidateRefs(changedStepIds).
3. Call expressionResolver.refreshAllSteps().

Optimization: scan which downstream steps actually contain expressions referencing the modified step ID, then only refresh those + nav bar + breadcrumb elements.


#### Quiz Answer Recording:
At recordQuizAnswer() in uce-state.js, after saving quizAnswers[answerValue], also update cache keys for that choice step. Trigger refreshAllSteps() only after user clicks Next (when answer is confirmed/finalized, matching current UX model).

Performance note: If refreshing every click causes issues, switch to "preview on click, apply on Next confirmation" model instead.


#### Side Quest Completion:
Same flow as choices -- trigger refresh when side quest completes.


#### Nav Bar & Navigation Buttons:
Include navigation list items in refreshAllSteps(). Nav step numbers, CURRENT badges, duration badges all reflect their own step IDs.


### E. Performance Strategy

1. **Cache invalidation via dependency map.** Track which step IDs appear in which expression contexts. Build mapping during init: dependentMap[sourceStepId] = setOfTargetStepIdsThatReferenceIt. On change event, only re-parse steps that reference the changed step. Avoids O(n) scans across all steps.

2. **Debounce rapid refreshes.** During rapid changes, debounce refreshAllSteps() by ~200ms max. Prevents layout thrashing during chained events.

3. **Diff-based targeted re-rendering.** Instead of destroying/rebuilding entire DOM, selectively update ONLY text nodes within existing elements. Replace only specific leaf texts containing expressions rather than recreating full step containers.

4. **Batch DOM mutations.** Collect all DOM updates, apply together inside requestAnimationFrame() batch for minimal reflows.


### F. Edge Cases Table

| Case | Expected Behavior |
|------|------------------|
| Broken step ID in expression | Visible placeholder + red warning styling on affected area. Console.warn logged. |
| Missing field on valid step | Visible placeholder + red warning text. Console.warn logged. |
| Circular reference detection | Detect circular dep; fail gracefully, skip from refresh chain. |
| Nested expressions | No nesting supported per spec (flat only); treat innermost as plaintext. |
| Expression inside HTML attribute | Skip attributes entirely; only process innerText/textNode content. |
| Empty expression pattern | Treated as error/broken ref indicator. |
| Unicode in resolved values | Supported natively via JS strings. Full unicode pass-through. |
| Very deeply nested dotted path | Works normally; just more traversal depth involved in processing. |
| User browses backward/forward | Refresh triggers again; same logic applies uniformly. |
| Rapid successive answers | Debounced refresh avoids jank/stutter effects. |
| Page reload mid-resolution | Cached expressions rebuild automatically during init parse. |


---


## Implementation Order (Execute Sequentially)


### Phase 1: Core Resolver Engine
Implement expressionResolver.resolve() function with regex extraction + basic field resolution logic. Unit-testable in isolation from rest of system.
Files written: expressionResolver sub-module added to end of uce-engine.js or separate include file.


### Phase 2: State Integration
Wire expressionResolver into uce-state.js initialization and lifecycle methods. Add cache tracking fields. Hook into recordQuizAnswer(), handleChoiceSelection(), etc to trigger refresh events at appropriate times.
Files modified: uce-state.js state initialization/update, uce-engine.js event handlers.


### Phase 3: DOM Refresh Functions
Implement refreshAllSteps()/refreshSingleStep() methods with proper integration into init phase (load resolved templates initially on page load). Ensure nav row/breadcrumb updates propagate correctly.
Files modified: uce-engine.js lifecycle hooks and render pipelines.


### Phase 4: Error Handling & Polish
Add visible placeholder styling for bad/wrong-typed refs. Whitespace cleanup after removal. Debouncing logic for fast-refresh scenarios. Performance optimizations including dependency map building.
Files modified: uce-engine.js helper functions, _sass/_uce.scss for error visual styling.


### Phase 5: Final Testing
End-to-end verification including reload flows with cached state. Test all expression scopes in the table above. Verify <200 ms refresh latency for typical documents (~30 steps, reasonable size).


---


## Files to Modify

File | Purpose | Risk Level | Notes
--- | --- | --- | ---
docs/assets/js/uce-engine.js | Import expressionResolver, integrate into init/render/choice-handling paths | Major | Largest impact, most complex integration. Requires careful review of all render/update pipeline hooks to ensure no conflicts.
docs/assets/js/uce-state.js | Add expression_cache and resolved_steps fields to state object | Minor | Simple struct addition to existing state manager. Low risk.
docs/_sass/_uce.scss | Style for broken-ref warning indicator elements | None | Small CSS-only change for red/orange styling of warning text. Safe change.
docs/_layouts/use_case.html | Optional include update if resolver moved to separate bundled file | Minimal | Already loads engine scripts sequentially. May need CDN/include update only if expressionResolver separated out.


---


## Acceptance Criteria (QA Checks)

1. Author writes ONE expression instance once in template. Downstream sections auto-resolve dynamically to the selected cloud/provider name based on user selections. NO manual branching or duplication needed (single copy of each step works for all providers).

2. Invalid/wrong expressions are visibly flagged (not silently dropped) so authors can catch typos immediately during development. Shown as `[broken ref: ...]` styled in red/orange warning color.

3. Reloaded pages show resolved values correctly based on saved localStorage state. If user chose AWS yesterday and returns today, they see previously resolved values matching their prior selections. Nav breadcrumbs reflect current position accurately.

4. Refresh latency after clicking "Next" on an interactive step must be <100-200 ms for normal-sized documents (under ~30 steps with reasonable text size).

5. Every text-containing element described in the "Apply Everywhere" table is covered and tested independently.

6. Whitespace handling: removing/resolving expressions leaves clean prose text (no double spaces, awkward gaps between words). Collapsed to standard spacing rules consistently.


---


## Open Questions for Implementer to Decide Later

Question 1: Should we support defaults inside expressions so authors can provide fallbacks when target data isn't ready yet? Format would be `field|DefaultValue` producing "Default Name" instead of nothing while awaiting resolution. Recommendation: Phase 2 feature (nice to have but not blocking Phase 1 delivery.)

Question 2: Support async evaluation of expressions? Currently assumed synchronous (render HTML first, then resolve immediately). Probably not worth adding async complexity for Phase 1 unless testing reveals immediate need.

Question 3: Do we need inline code formatting inside body text alongside expressions? YES, assumed already working correctly (inline <code> wrapping expressions still resolves properly). But verify during Phase 5 testing that wrapped expressions don't break formatting.

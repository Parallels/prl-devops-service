window.navTelemetry = (function() {
  var _initialized = false;

  function init() {
    if (_initialized) return;
    _initialized = true;

    // Delegate clicks on elements with data-clarity-event
    document.addEventListener('click', function(e) {
      var target = e.target.closest('[data-clarity-event]');
      if (!target) return;

      var eventName = target.getAttribute('data-clarity-event');
      var href = target.getAttribute('href');
      var newTab = target.hasAttribute('data-clarity-new-tab');
      var delay = parseInt(target.getAttribute('data-clarity-delay'), 10) || 200;

      if (typeof window.clarity === 'function') {
        window.clarity('event', eventName);
      }

      if (href) {
        e.preventDefault();
        setTimeout(function() {
          if (newTab) {
            window.open(href, '_blank', 'noopener');
          } else {
            window.location.href = href;
          }
        }, delay);
      }
    });
  }

  function navigate(eventName, url, options) {
    options = options || {};
    var delay = options.delay || 200;
    var newTab = !!options.newTab;

    if (typeof window.clarity === 'function') {
      window.clarity('event', eventName);
    }

    setTimeout(function() {
      if (newTab) {
        window.open(url, '_blank', 'noopener');
      } else {
        window.location.href = url;
      }
    }, delay);
  }

  return { init: init, navigate: navigate };
})();

document.addEventListener('DOMContentLoaded', function() {
  window.navTelemetry.init();
});

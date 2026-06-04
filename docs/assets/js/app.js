function openTab(evt, tabName) {
  var i, x, tablinks;
  x = document.getElementsByClassName("content-tab");
  for (i = 0; i < x.length; i++) {
    x[i].style.display = "none";
  }
  tablinks = document.getElementsByClassName("tab");
  for (i = 0; i < x.length; i++) {
    tablinks[i].className = tablinks[i].className.replace(" is-active", "");
  }
  document.getElementById(tabName).style.display = "block";
  evt.currentTarget.className += " is-active";
}

var getJSON = function (url, callback) {
  var xhr = new XMLHttpRequest();
  xhr.open("GET", url, true);
  xhr.responseType = "json";
  xhr.onload = function () {
    var status = xhr.status;
    if (status === 200) {
      callback(null, xhr.response);
    } else {
      callback(status, xhr.response);
    }
  };
  xhr.send();
};

function getOS() {
  const userAgent = window.navigator.userAgent,
    platform =
      window.navigator?.userAgentData?.platform || window.navigator.platform,
    macosPlatforms = ["macOS", "Macintosh", "MacIntel", "MacPPC", "Mac68K"],
    windowsPlatforms = ["Win32", "Win64", "Windows", "WinCE"],
    iosPlatforms = ["iPhone", "iPad", "iPod"];
  let os = null;

  if (macosPlatforms.indexOf(platform) !== -1) {
    os = "darwin";
  } else if (iosPlatforms.indexOf(platform) !== -1) {
    os = "ios";
  } else if (windowsPlatforms.indexOf(platform) !== -1) {
    os = "windows";
  } else if (/Android/.test(userAgent)) {
    os = "android";
  } else if (/Linux/.test(platform)) {
    os = "linux";
  }

  return os;
}

window.onload = function () {
  let os = getOS();
  if (os == "darwin") {
    let el = document.getElementById("download_tab_mac");
    if (el) {
        el.classList.add("is-active");
        document.getElementById("download_tab_mac_content").style.display = "block";
    }
  } else if (os == "windows") {
    let el = document.getElementById("download_tab_windows");
    if (el) {
      el.classList.add("is-active");
      document.getElementById("download_tab_windows_content").style.display = "block";
    }
  } else if (os == "linux") {
    let el = document.getElementById("download_tab_linux");
    if (el) {
      el.classList.add("is-active");
      document.getElementById("download_tab_linux_content").style.display = "block";
    }
  }

  document.querySelectorAll('aside[data-target="page-menu"]').forEach(function (element) {
    element.addEventListener('click', function () {
      target_menu = document.querySelectorAll('aside[data-id="page-menu"]');
      target_menu.forEach(function (menu) {
        menu.classList.toggle('expanded');
      });
      caret = element.querySelectorAll('i[data-target="page-menu"]');
      caret.forEach(function (c) {
        c.classList.toggle('expanded');
      });
      element.classList.toggle('expanded');
    });
  });

  // Mobile doc nav toggle — slide-in drawer
  const navToggle = document.querySelector('.doc-nav-toggle');
  const menuBar = document.querySelector('.menu-bar');

  if (navToggle && menuBar) {
    // Create overlay element
    let overlay = document.querySelector('.menu-bar-overlay');
    if (!overlay) {
      overlay = document.createElement('div');
      overlay.className = 'menu-bar-overlay';
      document.body.appendChild(overlay);
    }

    function openDrawer() {
      menuBar.classList.add('is-open');
      overlay.classList.add('is-visible');
      navToggle.classList.add('is-open');
      navToggle.setAttribute('aria-expanded', 'true');
    }

    function closeDrawer() {
      menuBar.classList.remove('is-open');
      overlay.classList.remove('is-visible');
      navToggle.classList.remove('is-open');
      navToggle.setAttribute('aria-expanded', 'false');
    }

    navToggle.addEventListener('click', function () {
      const isOpen = menuBar.classList.contains('is-open');
      if (isOpen) {
        closeDrawer();
      } else {
        openDrawer();
      }
    });

    overlay.addEventListener('click', closeDrawer);

    // Close on escape key
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && menuBar.classList.contains('is-open')) {
        closeDrawer();
      }
    });
  }

  // Mobile TOC toggle — show doc-toc-right as floating panel
  const tocToggle = document.querySelector('.doc-toc-toggle');
  const docTocRight = document.querySelector('.doc-toc-right');

  if (tocToggle && docTocRight) {
    // Create overlay element
    let tocOverlay = document.querySelector('.doc-toc-overlay');
    if (!tocOverlay) {
      tocOverlay = document.createElement('div');
      tocOverlay.className = 'doc-toc-overlay';
      document.body.appendChild(tocOverlay);
    }

    function openToc() {
      docTocRight.classList.add('is-mobile-open');
      tocOverlay.classList.add('is-visible');
      tocToggle.classList.add('is-open');
      tocToggle.setAttribute('aria-expanded', 'true');
    }

    function closeToc() {
      docTocRight.classList.remove('is-mobile-open');
      tocOverlay.classList.remove('is-visible');
      tocToggle.classList.remove('is-open');
      tocToggle.setAttribute('aria-expanded', 'false');
    }

    tocToggle.addEventListener('click', function () {
      const isOpen = docTocRight.classList.contains('is-mobile-open');
      if (isOpen) {
        closeToc();
      } else {
        openToc();
      }
    });

    tocOverlay.addEventListener('click', closeToc);

    // Close on escape key
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && docTocRight.classList.contains('is-mobile-open')) {
        closeToc();
      }
    });
  }
};

document.addEventListener("DOMContentLoaded", function () {
  const tabContainers = document.querySelectorAll('[data-type="tabs"]');
  for (let i = 0; i < tabContainers.length; i++) {
    
    const tabContainer = tabContainers[i];
    const defaultTab = tabContainer.getAttribute('data-default');
    setupTabs(tabContainer, defaultTab);
  }
})

document.addEventListener("DOMContentLoaded", function () {
  const tocLinks = Array.prototype.slice.call(document.querySelectorAll('.doc-toc-right a[href^="#"]'), 0);
  if (tocLinks.length === 0) {
    return;
  }

  const headings = tocLinks
    .map(function (link) {
      const id = decodeURIComponent(link.getAttribute('href').slice(1));
      return document.getElementById(id);
    })
    .filter(Boolean);

  if (headings.length === 0) {
    return;
  }

  function setActiveLink(id) {
    tocLinks.forEach(function (link) {
      const isActive = link.getAttribute('href') === '#' + id;
      link.classList.toggle('is-active', isActive);
      link.classList.toggle('active', isActive);
    });
  }

  function updateActiveTocLink() {
    const stickyOffset = 188;
    let current = headings[0];

    headings.forEach(function (heading) {
      if (heading.getBoundingClientRect().top <= stickyOffset) {
        current = heading;
      }
    });

    if (current && current.id) {
      setActiveLink(current.id);
    }
  }

  tocLinks.forEach(function (link) {
    link.addEventListener('click', function () {
      const id = decodeURIComponent(link.getAttribute('href').slice(1));
      setActiveLink(id);
    });
  });

  updateActiveTocLink();
  window.addEventListener('scroll', updateActiveTocLink, { passive: true });
});

document.addEventListener("DOMContentLoaded", function () {
  // Find the visible copy menu (mobile nav version or desktop toc-right version)
  const allMenus = document.querySelectorAll('[data-doc-copy-menu]');
  let copyMenu = null;
  for (let i = 0; i < allMenus.length; i++) {
    const style = window.getComputedStyle(allMenus[i]);
    if (style.display !== 'none') {
      copyMenu = allMenus[i];
      break;
    }
  }
  // Fallback: use first menu if none appear visible (e.g., during SSR)
  if (!copyMenu && allMenus.length > 0) {
    copyMenu = allMenus[0];
  }
  const pageContent = document.querySelector('.page-body .content');
  if (!copyMenu || !pageContent) {
    return;
  }

  const toggle = copyMenu.querySelector('[data-doc-copy-toggle]');
  const label = copyMenu.querySelector('[data-doc-copy-label]');
  const dropdown = copyMenu.querySelector('.doc-copy-dropdown');
  const dropdownPlaceholder = dropdown ? document.createComment('doc-copy-dropdown') : null;
  const actions = Array.prototype.slice.call(copyMenu.querySelectorAll('[data-doc-copy-action]'), 0);
  let dropdownPortaled = false;
  const llmUrls = {
    chatgpt: 'https://chatgpt.com/',
    claude: 'https://claude.ai/new',
    copilot: 'https://github.com/copilot'
  };

  function normalizeText(text) {
    return text.replace(/\s+/g, ' ');
  }

  function absoluteUrl(url) {
    try {
      return new URL(url, window.location.href).href;
    } catch (e) {
      return url;
    }
  }

  function inlineMarkdown(node) {
    if (node.nodeType === Node.TEXT_NODE) {
      return normalizeText(node.textContent);
    }

    if (node.nodeType !== Node.ELEMENT_NODE) {
      return '';
    }

    const tag = node.tagName.toLowerCase();
    const content = Array.prototype.map.call(node.childNodes, inlineMarkdown).join('');

    if (tag === 'br') {
      return '  \n';
    }

    if (tag === 'strong' || tag === 'b') {
      return `**${content.trim()}**`;
    }

    if (tag === 'em' || tag === 'i') {
      return `*${content.trim()}*`;
    }

    if (tag === 'code' && (!node.parentElement || node.parentElement.tagName.toLowerCase() !== 'pre')) {
      return '`' + node.textContent.replace(/`/g, '\\`') + '`';
    }

    if (tag === 'a') {
      const href = node.getAttribute('href');
      const text = content.trim() || href || '';
      if (!href) {
        return text;
      }
      return `[${text}](${absoluteUrl(href)})`;
    }

    if (tag === 'img') {
      const src = node.getAttribute('src');
      const alt = node.getAttribute('alt') || '';
      return src ? `![${alt}](${absoluteUrl(src)})` : '';
    }

    return content;
  }

  function tableToMarkdown(table) {
    const rows = Array.prototype.slice.call(table.querySelectorAll('tr'));
    if (rows.length === 0) {
      return '';
    }

    const markdownRows = rows.map(function (row) {
      const cells = Array.prototype.slice.call(row.children);
      return '| ' + cells.map(function (cell) {
        return inlineMarkdown(cell).trim().replace(/\|/g, '\\|');
      }).join(' | ') + ' |';
    });

    const columnCount = rows[0].children.length;
    const separator = '| ' + Array(columnCount).fill('---').join(' | ') + ' |';
    markdownRows.splice(1, 0, separator);
    return markdownRows.join('\n');
  }

  function listToMarkdown(list, depth) {
    const ordered = list.tagName.toLowerCase() === 'ol';
    const items = Array.prototype.slice.call(list.children).filter(function (child) {
      return child.tagName && child.tagName.toLowerCase() === 'li';
    });

    return items.map(function (item, index) {
      const nestedLists = Array.prototype.slice.call(item.children).filter(function (child) {
        const tag = child.tagName ? child.tagName.toLowerCase() : '';
        return tag === 'ul' || tag === 'ol';
      });
      const clone = item.cloneNode(true);
      Array.prototype.slice.call(clone.children).forEach(function (child) {
        const tag = child.tagName ? child.tagName.toLowerCase() : '';
        if (tag === 'ul' || tag === 'ol') {
          child.remove();
        }
      });

      const marker = ordered ? `${index + 1}. ` : '- ';
      const indent = '  '.repeat(depth);
      const body = inlineMarkdown(clone).trim();
      const nested = nestedLists.map(function (nestedList) {
        return listToMarkdown(nestedList, depth + 1);
      }).filter(Boolean).join('\n');

      return `${indent}${marker}${body}${nested ? '\n' + nested : ''}`;
    }).join('\n');
  }

  function blockMarkdown(node) {
    if (node.nodeType === Node.TEXT_NODE) {
      return normalizeText(node.textContent).trim();
    }

    if (node.nodeType !== Node.ELEMENT_NODE) {
      return '';
    }

    const tag = node.tagName.toLowerCase();

    if (['script', 'style', 'button', 'svg'].includes(tag)) {
      return '';
    }

    if (/^h[1-6]$/.test(tag)) {
      const level = Number(tag.slice(1));
      return `${'#'.repeat(level)} ${inlineMarkdown(node).trim()}`;
    }

    if (tag === 'p') {
      return inlineMarkdown(node).trim();
    }

    if (tag === 'pre') {
      const code = node.querySelector('code');
      const codeNode = code || node;
      const className = code ? code.className : '';
      const languageMatch = className.match(/language-([a-z0-9_-]+)/i);
      const language = languageMatch ? languageMatch[1] : '';
      return '```' + language + '\n' + codeNode.textContent.replace(/\n+$/, '') + '\n```';
    }

    if (tag === 'ul' || tag === 'ol') {
      return listToMarkdown(node, 0);
    }

    if (tag === 'blockquote') {
      return markdownChildren(node).split('\n').map(function (line) {
        return line.trim() ? `> ${line}` : '>';
      }).join('\n');
    }

    if (tag === 'table') {
      return tableToMarkdown(node);
    }

    if (tag === 'hr') {
      return '---';
    }

    if (tag === 'img') {
      return inlineMarkdown(node);
    }

    return markdownChildren(node);
  }

  function markdownChildren(parent) {
    return Array.prototype.map.call(parent.childNodes, blockMarkdown)
      .map(function (part) { return part.trim(); })
      .filter(Boolean)
      .join('\n\n');
  }

  function pageMarkdown() {
    const markdown = markdownChildren(pageContent).trim();
    return `${markdown}\n\n---\n\nSource: ${window.location.href}\n`;
  }

  function copyText(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text);
    }

    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.setAttribute('readonly', '');
    textarea.style.position = 'fixed';
    textarea.style.top = '-9999px';
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    textarea.remove();
    return Promise.resolve();
  }

  function setCopied(message) {
    const original = label ? label.getAttribute('data-original-label') || label.textContent : '';
    if (label && !label.getAttribute('data-original-label')) {
      label.setAttribute('data-original-label', original);
    }

    if (label) {
      label.textContent = message || 'Copied';
    }
    copyMenu.classList.add('is-copied');

    window.clearTimeout(copyMenu._copiedTimer);
    copyMenu._copiedTimer = window.setTimeout(function () {
      if (label) {
        label.textContent = label.getAttribute('data-original-label');
      }
      copyMenu.classList.remove('is-copied');
    }, 1800);
  }

  function closeMenu() {
    copyMenu.classList.remove('is-open');
    if (toggle) {
      toggle.setAttribute('aria-expanded', 'false');
    }
    if (dropdown) {
      dropdown.style.display = 'none';
      dropdown.style.visibility = '';
    }
  }

  function portalDropdown() {
    if (!dropdown || dropdownPortaled) {
      return;
    }

    dropdown.parentNode.insertBefore(dropdownPlaceholder, dropdown);
    document.body.appendChild(dropdown);
    dropdownPortaled = true;
  }

  function positionDropdown() {
    if (!dropdown || !toggle || !copyMenu.classList.contains('is-open')) {
      return;
    }

    portalDropdown();
    dropdown.style.display = 'block';
    dropdown.style.visibility = 'hidden';
    dropdown.style.maxHeight = '';

    const viewportPadding = 12;
    const gap = 8;
    const triggerRect = toggle.getBoundingClientRect();
    const menuRect = dropdown.getBoundingClientRect();
    const menuHeight = menuRect.height;
    const spaceBelow = window.innerHeight - triggerRect.bottom - viewportPadding;
    const spaceAbove = triggerRect.top - viewportPadding;
    const openAbove = menuHeight + gap > spaceBelow && spaceAbove > spaceBelow;
    const maxHeight = Math.max(160, (openAbove ? spaceAbove : spaceBelow) - gap);
    const renderedHeight = Math.min(menuHeight, maxHeight);

    let top = openAbove
      ? triggerRect.top - renderedHeight - gap
      : triggerRect.bottom + gap;
    let left = triggerRect.right - menuRect.width;

    left = Math.min(left, window.innerWidth - menuRect.width - viewportPadding);
    left = Math.max(left, viewportPadding);

    if (openAbove) {
      top = Math.max(top, viewportPadding);
    } else {
      top = Math.min(top, window.innerHeight - renderedHeight - viewportPadding);
      top = Math.max(top, viewportPadding);
    }

    dropdown.style.left = `${Math.round(left)}px`;
    dropdown.style.top = `${Math.round(top)}px`;
    dropdown.style.maxHeight = `${Math.round(maxHeight)}px`;
    dropdown.style.visibility = 'visible';
  }

  function openMenu() {
    copyMenu.classList.add('is-open');
    if (toggle) {
      toggle.setAttribute('aria-expanded', 'true');
    }
    positionDropdown();
  }

  function openMarkdownPreview(markdown) {
    const blob = new Blob([markdown], { type: 'text/markdown;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank', 'noopener');
    window.setTimeout(function () {
      URL.revokeObjectURL(url);
    }, 30000);
  }

  if (toggle) {
    toggle.addEventListener('click', function (event) {
      event.stopPropagation();
      const open = !copyMenu.classList.contains('is-open');
      if (open) {
        openMenu();
      } else {
        closeMenu();
      }
    });
  }

  actions.forEach(function (button) {
    button.addEventListener('click', function () {
      const action = button.getAttribute('data-doc-copy-action');
      const markdown = pageMarkdown();

      if (action === 'view') {
        openMarkdownPreview(markdown);
        closeMenu();
        return;
      }

      copyText(markdown).then(function () {
        if (llmUrls[action]) {
          window.open(llmUrls[action], '_blank', 'noopener');
          setCopied('Copied for ' + button.querySelector('strong').textContent.replace('Ask ', ''));
        } else {
          setCopied('Copied');
        }
        closeMenu();
      }).catch(function () {
        setCopied('Copy failed');
      });
    });
  });

  document.addEventListener('click', function (event) {
    if (!copyMenu.contains(event.target) && (!dropdown || !dropdown.contains(event.target))) {
      closeMenu();
    }
  });

  document.addEventListener('keydown', function (event) {
    if (event.key === 'Escape') {
      closeMenu();
    }
  });

  window.addEventListener('resize', positionDropdown, { passive: true });
  window.addEventListener('scroll', positionDropdown, { passive: true });
});

function setupTabs(tabContainerSelector, defaultTabTarget) {
  const tabContainer = tabContainerSelector;

  if (!tabContainer) {
    console.error(`Tab container with selector "${tabContainerSelector}" not found.`);
    return;
  }

  const tabs = tabContainer.querySelectorAll('[data-target]');
  const tabBlocks = tabContainer.querySelectorAll('[data-type="tab-block"]');

  // showing the first tab as default unless default is set
  if (tabs.length > 0) {
    var tab
    if (defaultTabTarget) {
      for (let i = 0; i < tabs.length; i++) {
        if (tabs[i].getAttribute('data-target') === defaultTabTarget) {
          tab = tabs[i];
          break;
        }
      }
    } else {
      tab = tabs[0];
    }

    if (tab) {
      const firstTabTarget = tab.getAttribute('data-target');
    const tabBlock = tabContainer.querySelector(`#${firstTabTarget}`);
      // Set the first tab as active
      tab.classList.add('active');
      tab.parentElement.classList.add('is-active');

      if (tabBlock) {
        tabBlock.classList.add('is-selected');
      }

    }
  }

  // Add click event listeners to each tab
  tabs.forEach(tab => {
    tab.addEventListener('click', event => {
      event.preventDefault();

      // Hide all tab blocks
      tabBlocks.forEach(block => {
        block.classList.remove('is-selected')
      });

      // Remove active class from all tabs
      tabs.forEach(tab => {
        tab.classList.remove('active')
        tab.parentElement.classList.remove('is-active');
      });

      // Show the target tab block
      const target = tab.getAttribute('data-target');
      const targetBlock = tabContainer.querySelector(`#${target}`);
      if (targetBlock) {
        targetBlock.classList.add('is-selected');
      }

      // Add active class to the clicked tab
      tab.classList.add('active');
      tab.parentElement.classList.add('is-active');
    });
  });
}

jQuery(function() {
    var $el;

	$("section > div.highlighter-rouge:first-of-type").each(function(i) {
		var $this = $(this).before("<ul></ul>"),
		$languages = $this.prev(),
		$notFirst = $this.nextUntil(":not(div.highlighter-rouge)"),
		$all = $this.add($notFirst);
        $languages.wrapAll("<div class=\"tabs is-boxed\"></div>");

        listLanguages($all, $languages);

		$this.css('display', 'block');
		$notFirst.css('display', 'none');

		$languages.find('li').first().addClass('is-active');
		$languages.find('a').first().addClass('active');

		$languages.find('li').click(function() {
			$all.css('display', 'none');
			$all.eq($(this).index()).css('display', 'block');

			$languages.find('li').removeClass('is-active');
			$languages.find('a').removeClass('active');
			$(this).addClass('is-active');
            $(this).find('a').addClass('active');
			return false;
		});

		if ($languages.children().length === 0) {
			$languages.remove();
		}
	});

	function listLanguages($el, $insert) {
		$el.each(function(i) {
			var title = $(this).attr('title');
			if (title) {
				$insert.append("<li class=\"tab\"><a href=\"#\">" + title + "</a></li>");
			}
		});
	}
});

// jQuery(function() {
//     var $el;

// 	$("section > div.highlighter-rouge:first-of-type").each(function(i) {
// 		var $this = $(this).before("<ul class=\"languages\"></ul>"),
// 		$languages = $this.prev(),
// 		$notFirst = $this.nextUntil(":not(div.highlighter-rouge)"),
// 		$all = $this.add($notFirst);

// 		$all.add($languages).wrapAll("<div class=\"code-viewer\"></div>");
        
// 		listLanguages($all, $languages);

// 		$this.css('display', 'block');
// 		$notFirst.css('display', 'none');

// 		$languages.find('a').first().addClass('active');

// 		$languages.find('a').click(function() {
// 			$all.css('display', 'none');
// 			$all.eq($(this).parent().index()).css('display', 'block');

// 			$languages.find('a').removeClass('active');
// 			$(this).addClass('active');
// 			return false;
// 		});

// 		if ($languages.children().length === 0) {
// 			$languages.remove();
// 		}
// 	});

// 	function listLanguages($el, $insert) {
// 		$el.each(function(i) {
// 			var title = $(this).attr('title');
// 			if (title) {
// 				$insert.append("<li><a href=\"#\">" + title + "</a></li>");
// 			}
// 		});
// 	}
// });

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
};

document.addEventListener("DOMContentLoaded", function () {
  const tabContainers = document.querySelectorAll('[data-type="tabs"]');
  for (let i = 0; i < tabContainers.length; i++) {
    
    const tabContainer = tabContainers[i];
    const defaultTab = tabContainer.getAttribute('data-default');
    setupTabs(tabContainer, defaultTab);
  }
})

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

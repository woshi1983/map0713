(function () {
  var logoWrap = document.querySelector('.theme-navbar__logo-wrap');
  if (!logoWrap) { return; }

  fetch('/api/branding', { cache: 'no-store' })
    .then(function (response) {
      if (!response.ok) { throw new Error('branding unavailable'); }
      return response.json();
    })
    .then(function (branding) {
      var language = 'en-us';
      try { language = localStorage.getItem('map-language') || language; } catch (_) { /* storage can be disabled */ }
      var selected = language.toLowerCase().indexOf('zh') === 0 ? branding.nameZh : branding.nameEn;
      var name = typeof selected === 'string' ? selected.trim() : '';
      if (!name) { return; }

      document.title = name;
      logoWrap.style.display = 'inline-flex';
      logoWrap.style.alignItems = 'center';
      logoWrap.style.gap = '10px';
      logoWrap.style.width = 'auto';
      logoWrap.style.maxWidth = '100%';
      logoWrap.style.boxSizing = 'border-box';
      var logo = logoWrap.querySelector('img');
      if (logo) { logo.alt = ''; }

      var label = document.createElement('span');
      label.textContent = name;
      label.style.color = '#333';
      label.style.fontSize = '16px';
      label.style.fontWeight = '600';
      label.style.minWidth = '0';
      label.style.overflow = 'hidden';
      label.style.textOverflow = 'ellipsis';
      label.style.whiteSpace = 'nowrap';
      logoWrap.appendChild(label);
    })
    .catch(function () { /* the login form remains available if branding cannot load */ });
})();

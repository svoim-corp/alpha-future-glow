/* Design review interactions only. No authentication, search, API or business logic. */
(() => {
  const root = document.documentElement;
  const theme = document.querySelector('#theme');
  let saved;
  try { saved = localStorage.getItem('credit-ui-theme'); } catch {}
  const darkOS = matchMedia('(prefers-color-scheme: dark)');
  let manual = saved === 'light' || saved === 'dark';
  function setTheme(value) {
    root.dataset.theme = value;
    const dark = value === 'dark';
    if (!theme) return;
    theme.setAttribute('aria-pressed', String(dark));
    theme.innerHTML = `<svg class="icon" aria-hidden="true"><use href="#icon-${dark ? 'sun' : 'moon'}"/></svg>`;
  }
  setTheme(manual ? saved : darkOS.matches ? 'dark' : 'light');
  theme?.addEventListener('click', () => {
    manual = true;
    const value = root.dataset.theme === 'light' ? 'dark' : 'light';
    setTheme(value);
    try { localStorage.setItem('credit-ui-theme', value); } catch {}
  });
  darkOS.addEventListener('change', e => { if (!manual) setTheme(e.matches ? 'dark' : 'light'); });
  const triggers = ['profile'].map(id => document.getElementById(id)).filter(Boolean);
  function closePopovers() {
    triggers.forEach(button => { button.setAttribute('aria-expanded', 'false'); document.getElementById(button.getAttribute('aria-controls')).hidden = true; });
    const search = document.querySelector('#search');
    if (search) { search.setAttribute('aria-expanded', 'false'); document.querySelector('#search-preview').hidden = true; }
  }
  triggers.forEach(button => button.addEventListener('click', () => {
    const open = button.getAttribute('aria-expanded') !== 'true';
    closePopovers();
    button.setAttribute('aria-expanded', String(open));
    document.getElementById(button.getAttribute('aria-controls')).hidden = !open;
  }));
  document.addEventListener('click', e => { if (!e.target.closest('.dropdown, .search-wrap')) closePopovers(); if (!e.target.closest('.customize-preview')) document.querySelector('.customize-preview')?.removeAttribute('open'); });
  document.addEventListener('focusin', e => { if (!e.target.closest('.dropdown, .search-wrap')) closePopovers(); });
  const search = document.querySelector('#search');
  search?.addEventListener('focus', () => { closePopovers(); search.setAttribute('aria-expanded', 'true'); document.querySelector('#search-preview').hidden = false; });
  const sidebar = document.querySelector('#sidebar');
  const mobile = document.querySelector('#mobile-menu');
  const backdrop = document.querySelector('.sidebar-backdrop');
  function setSidebar(open) {
    if (!sidebar) return;
    sidebar.classList.toggle('is-open', open);
    backdrop.hidden = !open;
    mobile.setAttribute('aria-expanded', String(open));
    if (open) { closePopovers(); sidebar.querySelector('.mobile-close').focus(); }
  }
  mobile?.addEventListener('click', () => setSidebar(!sidebar.classList.contains('is-open')));
  backdrop?.addEventListener('click', () => { setSidebar(false); mobile.focus(); });
  document.querySelector('.mobile-close')?.addEventListener('click', () => { setSidebar(false); mobile.focus(); });
  const descriptions = {
    home: ['Главная', 'Ваш бизнес. Ваш следующий шаг.'], calendar: ['Планер', 'Важные даты и задачи — в одном месте.'],
    applications: ['Мои заявки', 'Понятный путь от заявки до решения.'], loans: ['Мои кредиты', 'Условия и обязательства под контролем.'],
    catalog: ['Каталог', 'Финансирование для задач вашего бизнеса.'],
    support: ['Поддержка', 'Помощь на каждом этапе.'], settings: ['Настройки', 'Кабинет, в котором вам удобно.'],
  };
  function syncPage() {
    if (!sidebar) return;
    const key = document.body.dataset.designPage || location.hash.slice(1) || 'home';
    const page = document.body.classList.contains('home-page') ? ['Добрый день, Олег', 'Всё важное о финансировании — перед вами.'] : document.body.dataset.designPage ? [document.querySelector('#page-title').textContent, document.querySelector('#page-description').textContent] : descriptions[key] || descriptions.home;
    document.querySelector('#page-title').textContent = page[0];
    document.querySelector('#page-description').textContent = page[1];
    document.title = document.body.classList.contains('home-page') ? 'Главная · Кредитный кабинет' : `${page[0]} · Кредитный кабинет`;
    document.querySelectorAll('[data-page]').forEach(link => {
      if (link.dataset.page === (descriptions[key] ? key : 'home')) link.setAttribute('aria-current', 'page');
      else link.removeAttribute('aria-current');
    });
  }
  syncPage();
  window.addEventListener('hashchange', syncPage);
  document.querySelectorAll('[data-page], .sidebar-brand').forEach(link => link.addEventListener('click', () => { if (sidebar.classList.contains('is-open')) { setSidebar(false); mobile.focus(); } }));
  let toastTimer;
  document.querySelector('#logout')?.addEventListener('click', () => {
    const toast = document.querySelector('#design-toast');
    toast.hidden = false;
    clearTimeout(toastTimer); toastTimer = setTimeout(() => { toast.hidden = true; }, 4000);
  });
  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') {
      const customize = document.querySelector('.customize-preview[open]');
      if (customize) { customize.removeAttribute('open'); customize.querySelector('summary').focus(); }
      const activeTrigger = triggers.find(button => button.getAttribute('aria-expanded') === 'true');
      closePopovers();
      activeTrigger?.focus();
      if (sidebar?.classList.contains('is-open')) { setSidebar(false); mobile.focus(); }
    }
    if (e.key === '/' && search && !e.target.closest('input, textarea, select, [contenteditable]')) { e.preventDefault(); search.focus(); }
    if (e.key === 'Tab' && sidebar?.classList.contains('is-open')) {
      const focusable = [...sidebar.querySelectorAll('a,button')].filter(el => el.getClientRects().length);
      const first = focusable[0], last = focusable.at(-1);
      if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
      else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
    }
  });
})();

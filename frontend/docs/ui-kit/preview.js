/* Standalone demo only. No API calls, remote fonts or dependencies. */
(() => {
  const root = document.documentElement;
  const toggle = document.querySelector('#theme');
  const swatches = [
    ['background', 'Фон страницы'], ['surface', 'Поверхность'], ['raised', 'Вложенный блок'],
    ['text', 'Основной текст'], ['text-muted', 'Вторичный текст'], ['primary', 'Главный акцент'],
    ['success', 'Успех'], ['warning', 'Внимание'], ['danger', 'Ошибка'], ['info', 'Информация'],
  ];
  const osTheme = window.matchMedia('(prefers-color-scheme: dark)');
  let manual = false;
  function setTheme(theme) {
    root.dataset.theme = theme;
    const dark = theme === 'dark';
    toggle.textContent = dark ? 'Светлая тема' : 'Тёмная тема';
    toggle.setAttribute('aria-label', 'Тёмная тема');
    toggle.setAttribute('aria-pressed', String(dark));
    const style = getComputedStyle(root);
    const target = document.querySelector('#swatches');
    target.replaceChildren(...swatches.map(([token, label]) => {
      const card = document.createElement('div'); card.className = 'swatch';
      const sample = document.createElement('div'); sample.className = 'swatch-color';
      sample.style.background = `var(--color-${token})`;
      const meta = document.createElement('div'); meta.className = 'swatch-meta';
      meta.textContent = label;
      const code = document.createElement('code'); code.textContent = style.getPropertyValue(`--color-${token}`).trim().toUpperCase();
      meta.append(code); card.append(sample, meta); return card;
    }));
  }
  let preferred;
  try { preferred = localStorage.getItem('credit-ui-theme'); } catch { /* Storage can be blocked. */ }
  manual = preferred === 'light' || preferred === 'dark';
  setTheme(manual ? preferred : osTheme.matches ? 'dark' : 'light');
  toggle.addEventListener('click', () => {
    manual = true;
    const theme = root.dataset.theme === 'light' ? 'dark' : 'light';
    setTheme(theme);
    try { localStorage.setItem('credit-ui-theme', theme); } catch { /* Theme still works. */ }
  });
  osTheme.addEventListener('change', event => { if (!manual) setTheme(event.matches ? 'dark' : 'light'); });
  const note = document.querySelector('#action-note');
  document.querySelectorAll('[data-note]').forEach(button => button.addEventListener('click', () => { note.textContent = button.dataset.note; }));
  document.querySelector('#demo-action').addEventListener('click', async event => {
    const button = event.currentTarget;
    button.disabled = true; button.setAttribute('aria-busy', 'true');
    button.innerHTML = '<span class="ui-spinner" aria-hidden="true"></span>Проверяем данные';
    note.textContent = 'Демонстрация: проверяем данные…';
    await new Promise(resolve => setTimeout(resolve, 1000));
    button.disabled = false; button.removeAttribute('aria-busy');
    button.innerHTML = 'Продолжить <span aria-hidden="true">↗</span>';
    note.textContent = 'Проверка завершена — пример успешного ответа. Запросы в банк не выполнялись.';
  });
  document.querySelector('#sample-form').addEventListener('submit', event => {
    event.preventDefault();
    const input = document.querySelector('#amount');
    const raw = input.value.replace(/\s/g, '');
    const amount = Number(raw);
    const valid = /^\d+$/.test(raw) && Number.isSafeInteger(amount) && amount >= 100000 && amount <= 5000000;
    input.setAttribute('aria-invalid', String(!valid));
    const error = document.querySelector('#amount-error');
    error.hidden = valid; error.textContent = valid ? '' : 'Введите целую сумму от 100 000 до 5 000 000 ₽';
    document.querySelector('#form-result').textContent = valid ? 'Сумма в демонстрационном диапазоне. Это не проверка заявки и не решение банка.' : 'Проверьте сумму кредита.';
    if (!valid) input.focus();
    else input.value = new Intl.NumberFormat('ru-RU').format(amount);
  });
  document.querySelector('#statement').addEventListener('change', event => {
    const input = event.currentTarget;
    const file = input.files[0];
    const status = document.querySelector('#file-status');
    const invalid = file && (!/\.txt$/i.test(file.name) || file.size > 5 * 1024 * 1024);
    input.setAttribute('aria-invalid', String(Boolean(invalid)));
    status.classList.toggle('ui-error', Boolean(invalid));
    status.textContent = !file ? 'Файл не выбран.' : invalid ? 'Выберите файл .txt размером до 5 МБ.' : `Выбран: ${file.name}. Файл остаётся на вашем устройстве; загрузка на сервер не выполняется.`;
    if (invalid) input.value = '';
  });
  if ('IntersectionObserver' in window) {
    const links = [...document.querySelectorAll('nav a')];
    const observer = new IntersectionObserver(entries => {
      const visible = entries.filter(entry => entry.isIntersecting).sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0];
      if (!visible) return;
      links.forEach(link => {
        if (link.hash === `#${visible.target.id}`) link.setAttribute('aria-current', 'location');
        else link.removeAttribute('aria-current');
      });
    }, { rootMargin: '-5% 0px -65% 0px' });
    document.querySelectorAll('main section').forEach(section => observer.observe(section));
  }
})();

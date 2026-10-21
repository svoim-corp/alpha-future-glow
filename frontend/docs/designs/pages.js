/* View-only interactions. All data is synthetic; there are no requests to a backend. */
(() => {
  const money = value => new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(value);
  const escape = value => String(value).replace(/[&<>"']/g, char => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[char]));
  let toastTimer;
  function demo(message, source) {
    const dialog = source?.closest('dialog');
    if (dialog) {
      let feedback = dialog.querySelector('.dialog-feedback');
      if (!feedback) { feedback = document.createElement('p'); feedback.className = 'dialog-notice dialog-feedback'; feedback.setAttribute('role', 'status'); dialog.append(feedback); }
      feedback.textContent = message;
      return;
    }
    const toast = document.querySelector('#design-toast');
    toast.textContent = message; toast.hidden = false;
    clearTimeout(toastTimer); toastTimer = setTimeout(() => { toast.hidden = true; }, 5500);
  }
  document.querySelectorAll('[data-demo]').forEach(button => button.addEventListener('click', () => demo(button.dataset.demo, button)));
  function openDialog(id, trigger) {
    const dialog = document.getElementById(id);
    if (!dialog) return;
    dialog.querySelector('.dialog-feedback')?.remove();
    dialog.showModal();
    dialog.addEventListener('close', () => trigger?.focus(), { once: true });
  }
  document.querySelectorAll('[data-open-dialog]').forEach(button => button.addEventListener('click', () => openDialog(button.dataset.openDialog, button)));
  document.querySelectorAll('dialog').forEach(dialog => dialog.addEventListener('click', event => {
    if (event.target !== dialog) return;
    const bounds = dialog.getBoundingClientRect();
    if (event.clientX < bounds.left || event.clientX > bounds.right || event.clientY < bounds.top || event.clientY > bounds.bottom) dialog.close();
  }));
  document.querySelectorAll('[role="tablist"]').forEach(list => {
    const tabs = [...list.querySelectorAll('[role="tab"]')];
    function select(tab) {
      tabs.forEach(item => {
        const active = item === tab;
        item.setAttribute('aria-selected', String(active)); item.tabIndex = active ? 0 : -1;
        document.getElementById(item.getAttribute('aria-controls')).hidden = !active;
      });
    }
    tabs.forEach((tab, index) => {
      tab.addEventListener('click', () => select(tab));
      tab.addEventListener('keydown', event => {
        if (!['ArrowLeft','ArrowRight','Home','End'].includes(event.key)) return;
        event.preventDefault();
        const next = event.key === 'Home' ? 0 : event.key === 'End' ? tabs.length - 1 : (index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length;
        select(tabs[next]); tabs[next].focus();
      });
    });
  });
  // Applications: query, filters, sorting, empty state and a detail preview.
  const appList = document.querySelector('#application-list');
  if (appList) {
    const rows = [...appList.querySelectorAll('.filter-row')];
    const query = document.querySelector('#application-search');
    const status = document.querySelector('#application-status');
    const product = document.querySelector('#application-product');
    const period = document.querySelector('#application-period');
    const sort = document.querySelector('#application-sort');
    function filterApplications() {
      let count = 0;
      rows.forEach(row => {
        const visible = row.dataset.name.toLocaleLowerCase('ru').includes(query.value.trim().toLocaleLowerCase('ru')) && (!status.value || row.dataset.status === status.value) && (!product.value || row.dataset.product === product.value) && Number(row.dataset.date) >= Number(period.value);
        row.hidden = !visible; if (visible) count++;
      });
      [...rows].sort((a,b) => sort.value === 'amount-desc' ? Number(b.dataset.amount)-Number(a.dataset.amount) : sort.value === 'amount-asc' ? Number(a.dataset.amount)-Number(b.dataset.amount) : Number(b.dataset.date)-Number(a.dataset.date)).forEach(row => appList.append(row));
      document.querySelector('#application-count').textContent = `Найдено: ${count}`;
      document.querySelector('#application-empty').hidden = count > 0;
      const bottom = document.querySelector('.list-bottom');
      bottom.firstElementChild.textContent = count === rows.length ? 'Показаны все заявки' : 'С учётом выбранных фильтров';
      bottom.lastElementChild.textContent = count ? `1–${count} из ${count}` : '0 из 0';
    }
    [query,status,product,period,sort].forEach(input => input.addEventListener(input === query ? 'input' : 'change', filterApplications));
    document.querySelector('#reset-application-filters').addEventListener('click', () => { query.value = ''; status.value = ''; product.value = ''; period.value = '0'; sort.value = 'date'; filterApplications(); query.focus(); });
    document.querySelectorAll('[data-application]').forEach(button => button.addEventListener('click', () => {
      const row = button.closest('.application-row');
      const status = row.dataset.status;
      document.querySelector('#application-detail-body').innerHTML = `<span class="row-eyebrow">${escape(row.querySelector('.row-eyebrow').textContent)}</span><h3>${escape(row.querySelector('h3').textContent)}</h3><div class="detail-values"><div><span>Сумма запроса</span><strong>${money(Number(row.dataset.amount))}</strong></div><div><span>Статус</span><strong>${escape(status)}</strong></div></div><div class="soft-info"><p>${escape(row.querySelector('.application-info p').textContent)}</p></div>${button.dataset.application === '1018' ? '<p class="dialog-description">Запрошено 1 200 000 ₽. Предложено 1 000 000 ₽. Полные условия нужно изучить перед принятием.</p>' : ''}<p class="dialog-notice">Пример карточки заявки. Форму дополнения, подписание и полную историю спроектируем отдельно.</p>`;
      openDialog('application-detail', button);
    }));
  }
  // Catalog: goal filter and a local illustrative annuity calculation.
  document.querySelectorAll('.catalog-goals [data-goal]').forEach(button => button.addEventListener('click', () => {
    document.querySelectorAll('.catalog-goals [data-goal]').forEach(item => { const active = item === button; item.classList.toggle('is-selected', active); item.setAttribute('aria-pressed', String(active)); });
    document.querySelectorAll('.product-card').forEach(card => { card.hidden = button.dataset.goal !== 'all' && card.dataset.goal !== button.dataset.goal; });
  }));
  const amount = document.querySelector('#calc-amount');
  if (amount) {
    const term = document.querySelector('#calc-term'), rate = document.querySelector('#calc-rate');
    function calculate() {
      const principal = Number(amount.value), months = Number(term.value), monthlyRate = Number(rate.value)/100/12;
      const payment = monthlyRate === 0 ? principal/months : principal*monthlyRate/(1-Math.pow(1+monthlyRate,-months));
      document.querySelector('#calc-amount-output').textContent = money(principal);
      const ending = months%10 === 1 && months%100 !== 11 ? 'месяц' : [2,3,4].includes(months%10) && ![12,13,14].includes(months%100) ? 'месяца' : 'месяцев';
      document.querySelector('#calc-term-output').textContent = `${months} ${ending}`;
      document.querySelector('#calc-payment').textContent = money(payment);
      document.querySelector('#calc-total').textContent = money(payment*months);
      document.querySelector('#calc-interest').textContent = money(payment*months-principal);
    }
    [amount,term,rate].forEach(input => input.addEventListener('input',calculate)); calculate();
  }
  const products = {
    working: {name:'Оборотный кредит',amount:'300 тыс. — 10 млн ₽',term:'3–36 месяцев',rate:'20–28 % годовых',note:'Закупки, зарплаты, налоги и аренда. Условия из демонстрационного API-контракта.'},
    invest: {name:'Инвестиционный кредит',amount:'1–30 млн ₽',term:'12–84 месяца',rate:'18–26 % годовых',note:'Оборудование, помещение и расширение. Для ООО в примере предусмотрен поручитель; выписка нужна всегда.'},
    express: {name:'Экспресс-кредит',amount:'100 тыс. — 1 млн ₽',term:'3–24 месяца',rate:'24,9–29,9 % годовых',note:'Предложенные параметры для макета. Окончательные рамки продукта ещё нужно согласовать.'},
    overdraft: {name:'Овердрафт',amount:'Лимит до 3 млн ₽',term:'Лимит на 12 месяцев',rate:'23–29 % годовых',note:'Пример параметров для макета. Только для клиентов банка. Срок отдельной задолженности определяется условиями продукта.'}
  };
  document.querySelectorAll('[data-product-preview]').forEach(button => button.addEventListener('click', () => {
    const item = products[button.dataset.productPreview];
    document.querySelector('#product-detail-title').textContent = item.name;
    document.querySelector('#product-detail-body').innerHTML = `<div class="detail-values"><div><span>Сумма</span><strong>${item.amount}</strong></div><div><span>Срок</span><strong>${item.term}</strong></div><div><span>Диапазон ставок</span><strong>${item.rate}</strong></div></div><p class="dialog-description">${item.note}</p><div class="soft-info"><p>Итоговые условия определяются после рассмотрения. Подбор продукта не гарантирует одобрения.</p></div><p class="dialog-notice">Это просмотр дизайна. Оформление заявки не запускается.</p>`;
    openDialog('product-detail',button);
  }));
  document.querySelector('#faq-search')?.addEventListener('input',event => {
    let count = 0;
    document.querySelectorAll('.faq-item').forEach(item => { item.hidden = !item.dataset.name.toLocaleLowerCase('ru').includes(event.target.value.trim().toLocaleLowerCase('ru')); if (!item.hidden) count++; });
    document.querySelector('#faq-empty').hidden = count > 0;
  });
  document.querySelectorAll('[data-chat-prompt]').forEach(button => button.addEventListener('click', () => { const field = document.querySelector('#chat-message'); field.value = button.dataset.chatPrompt; field.focus(); }));
  document.querySelector('#chat-form')?.addEventListener('submit',event => {
    event.preventDefault();
    const field = document.querySelector('#chat-message'), value = field.value.trim();
    if (!value) return;
    const bubble = document.createElement('div'); bubble.className = 'chat-bubble chat-bubble--user';
    const text = document.createElement('p'); text.textContent = value;
    const caption = document.createElement('span'); caption.textContent = 'Только в макете · Не отправлено';
    bubble.append(text,caption);
    const log = document.querySelector('#chat-messages'); log.append(bubble); log.scrollTop = log.scrollHeight;
    field.value = ''; field.focus();
    document.querySelector('#chat-status').textContent = 'Сообщение показано в макете. Оно не отправлено в поддержку и исчезнет после обновления страницы.';
  });
})();

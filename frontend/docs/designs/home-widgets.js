/* Только интерактивный дизайн. Данные синтетические, запросов к API нет. */
(() => {
  const board = document.querySelector('#db-board');
  if (!board) return;
  const $ = selector => document.querySelector(selector);
  const esc = value => String(value).replace(/[&<>"']/g, char => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[char]));
  const icon = name => `<svg class="icon" aria-hidden="true"><use href="#icon-${name}"/></svg>`;
  const badge = (text, tone = '') => `<span class="db-badge ${tone}">${text}</span>`;
  const money = value => new Intl.NumberFormat('ru-RU').format(value) + ' ₽';
  const loans = {
    working: {name:'Оборотный кредит',number:'КД-2026-0042',balance:1125000,amount:1500000,payment:77800,rate:'21,5%',term:'24 месяца'},
    equipment: {name:'Кредит на оборудование',number:'КД-2026-0051',balance:2240000,amount:2400000,payment:113400,rate:'19,8%',term:'24 месяца'},
  };
  const applications = {
    warehouse: {name:'Оборотный кредит',number:'КЗ-2026-001024',amount:1500000,status:'Нужна информация'},
    express: {name:'Экспресс-кредит',number:'КЗ-2026-001031',amount:500000,status:'На рассмотрении'},
    investment: {name:'Инвестиционный кредит',number:'КЗ-2026-001035',amount:3200000,status:'Черновик'},
  };
  const types = {
    applications: {title:'Мои заявки',icon:'applications',description:'Карточки заявок, их статусы и следующие действия.',sizes:[[8,3],[8,4],[12,3]],preview:'3 заявки',binding:'filter'},
    loan: {title:'Мой кредит',icon:'loans',description:'Выбранный договор: остаток, условия и следующий платёж.',sizes:[[4,3],[4,4],[6,3]],preview:'1 125 000 ₽',binding:'loan'},
    payment: {title:'Ближайший платёж',icon:'calendar',description:'Ближайшая дата и сумма по всем кредитам или одному.',sizes:[[4,2],[4,3],[6,2]],preview:'191 200 ₽',binding:'scope'},
    debt: {title:'Основной долг',icon:'loans',description:'Остаток основного долга по действующим кредитам.',sizes:[[4,2],[4,3]],preview:'3 365 000 ₽'},
    quick: {title:'Быстрые действия',icon:'catalog',description:'Переходы к частым действиям в кредитном кабинете.',sizes:[[4,2],[4,3],[8,2]],preview:'Заявка · План · Банк',binding:'actions'},
    calendar: {title:'Календарь событий',icon:'calendar',description:'Даты платежей и планов в одном компактном календаре.',sizes:[[6,4],[8,4],[12,4]],preview:'Ноябрь 2026',binding:'events'},
    plans: {title:'Мои планы',icon:'calendar',description:'Ближайшие планы, бюджеты и идеи без даты.',sizes:[[6,4],[6,5],[12,4]],preview:'3 плана · 1 идея',binding:'plans'},
    application: {title:'Конкретная заявка',icon:'documents',description:'Отдельная заявка: статус и необходимое действие.',sizes:[[4,3],[6,3],[8,3]],preview:'Нужна информация',binding:'application'},
    loans: {title:'Все кредиты',icon:'loans',description:'Обзор действующих или закрытых договоров.',sizes:[[8,3],[8,4],[12,3]],preview:'2 договора',binding:'portfolio'},
    payments: {title:'Платежи на период',icon:'calendar',description:'Запланированные платежи за ближайшие 7 или 30 дней.',sizes:[[8,3],[4,3],[8,4]],preview:'191 200 ₽',binding:'period'},
    offers: {title:'Предложения банка',icon:'check',description:'Условия банка, срок предложения и следующий шаг.',sizes:[[6,4],[4,4],[8,4]],preview:'Кредит одобрен',binding:'offers'},
    drafts: {title:'Черновики',icon:'applications',description:'Последние незавершённые заявки и продолжение оформления.',sizes:[[4,3],[6,3],[8,3]],preview:'Продолжить заявку'},
    planFunding: {title:'План и финансирование',icon:'catalog',description:'Выбранный план и рекомендация продукта.',sizes:[[6,4],[4,4],[8,4]],preview:'Печь для выпечки',binding:'planFunding'},
    notifications: {title:'Уведомления',icon:'documents',description:'Последние события кабинета и непрочитанные сообщения.',sizes:[[4,4],[6,4],[8,4]],preview:'2 новых события',binding:'notifications'},
    tickets: {title:'Обращения в банк',icon:'support',description:'Открытые обращения и ответы специалистов.',sizes:[[6,3],[4,3],[6,4]],preview:'Есть ответ',binding:'tickets'},
    funding: {title:'Подобрать финансирование',icon:'catalog',description:'Продукты для оборота, оборудования и развития бизнеса.',sizes:[[4,3],[6,3],[8,3]],preview:'Для вашего бизнеса'},
  };
  const defaults = [
    {id:'applications-1',type:'applications',x:0,y:0,w:8,h:3,binding:'all'},
    {id:'loan-1',type:'loan',x:8,y:0,w:4,h:3,binding:'working'},
    {id:'payment-1',type:'payment',x:0,y:3,w:4,h:2,binding:'all'},
    {id:'debt-1',type:'debt',x:4,y:3,w:4,h:2},
    {id:'quick-1',type:'quick',x:8,y:3,w:4,h:2,binding:'standard'},
    {id:'calendar-1',type:'calendar',x:0,y:5,w:6,h:4,binding:'all'},
    {id:'plans-1',type:'plans',x:6,y:5,w:6,h:4,binding:'all'},
    {id:'application-1',type:'application',x:0,y:9,w:4,h:3,binding:'warehouse'},
  ];
  const clone = value => JSON.parse(JSON.stringify(value));
  const storageKey = 'design-home-widgets-v1';
  let widgets = clone(defaults);
  try {
    const saved = JSON.parse(localStorage.getItem(storageKey));
    if (Array.isArray(saved) && saved.length <= 30 && saved.every(item => item && types[item.type]
      && typeof item.id === 'string' && types[item.type].sizes.some(([w,h]) => w === item.w && h === item.h)
      && Number.isInteger(item.x) && item.x >= 0 && item.x + item.w <= 12 && Number.isInteger(item.y) && item.y >= 0 && item.y <= 200
      && (item.binding === undefined || typeof item.binding === 'string')) && new Set(saved.map(item => item.id)).size === saved.length) widgets = saved;
  } catch { /* Недоступное хранилище не мешает обсуждать макет. */ }
  let editing = false;
  let snapshot = null;
  let selectedId = null;
  let draggedId = null;
  let nextId = Date.now();
  const isTablet = () => matchMedia('(max-width:1250px)').matches;
  const mobile = () => matchMedia('(max-width:600px)').matches;
  const columns = () => isTablet() ? 6 : 12;
  const ordered = () => [...widgets].sort((a,b) => a.y-b.y || a.x-b.x);
  const overlap = (a,b) => a.x < b.x+b.w && a.x+a.w > b.x && a.y < b.y+b.h && a.y+a.h > b.y;
  const firstSpace = (item, occupied, cols) => {
    for (let y=0; ; y++) for (let x=0; x<=cols-item.w; x++) {
      const position = {...item,x,y};
      if (!occupied.some(other => overlap(position,other))) return position;
    }
  };
  // Перемещаемая карточка закрепляется, остальные занимают ближайшие свободные клетки.
  const pack = anchor => {
    const result = anchor ? [anchor] : [];
    ordered().filter(item => item.id !== anchor?.id).forEach(item => {
      result.push(result.some(other => overlap(item,other)) ? firstSpace(item,result,12) : item);
    });
    widgets = result;
  };
  const visualLayout = () => {
    if (!isTablet()) return widgets;
    const result = [];
    ordered().forEach(item => result.push(firstSpace({...item,w:item.w >= 6 ? 6 : 3},result,6)));
    return result;
  };
  const status = message => { $('#db-status').textContent = message; };
  const link = (href,text) => `<a class="db-link" href="${href}">${text} ↗</a>`;
  const content = item => {
    const loan = loans[item.binding] || loans.working;
    const application = applications[item.binding] || applications.warehouse;
    if (item.state === 'loading') return '<div class="db-skeleton" role="status" aria-label="Загрузка"><span></span><span></span><span></span></div>';
    if (item.state === 'error' && item.h === 2) return `<div class="db-state"><strong>Данные недоступны</strong><button class="db-button" data-retry="${esc(item.id)}">Повторить</button></div>`;
    if (item.state === 'error') return `<div class="db-state"><strong>Не удалось обновить данные</strong><span class="db-muted">Остальная главная доступна.</span><button class="db-button" data-retry="${esc(item.id)}">Повторить</button></div>`;
    if (item.state === 'empty' && item.h === 2) return '<div class="db-state"><strong>Пока нет данных</strong></div>';
    if (item.state === 'empty') return `<div class="db-state"><strong>Здесь пока ничего нет</strong><span class="db-muted">Данные появятся после создания заявки, плана или выдачи кредита.</span></div>`;
    switch (item.type) {
      case 'applications': {
        const entries = Object.values(applications).filter(app => item.binding !== 'attention' || app.status === 'Нужна информация');
        return `<div class="db-mini-apps">${entries.map(app => `<div class="db-mini-app"><strong>${app.name}</strong><span class="db-app-amount">${money(app.amount)}</span>${badge(app.status,app.status === 'Нужна информация' ? 'warning' : '')}</div>`).join('')}</div><div class="db-inline db-bottom"><span class="db-muted">${entries.length} ${entries.length === 1 ? 'заявка требует внимания' : 'заявки · 1 требует внимания'}</span>${link('applications.html','Открыть заявки')}</div>${item.h >= 4 ? '<p class="db-muted">У одной заявки банк запросил уточнение. Последняя сохранённая заявка доступна в черновиках.</p>' : ''}`;
      }
      case 'loan': return `<p class="db-loan-label">${loan.name} · ${loan.number}</p><div><p class="db-muted">Остаток основного долга</p><div class="db-number">${money(loan.balance)}</div></div><div class="db-inline db-muted"><span>${loan.rate} годовых</span><span>${loan.term}</span></div>${item.h >= 4 ? `<p class="db-muted">Следующий платёж 15 ноября · ${money(loan.payment)}</p>` : ''}<div class="db-progress db-bottom"><span style="width:${Math.round((1-loan.balance/loan.amount)*100)}%"></span></div><div class="db-inline db-muted"><span>Погашено ${Math.round((1-loan.balance/loan.amount)*100)}% долга</span>${link('loans.html','Детали')}</div>`;
      case 'payment': { const total = item.binding === 'all' || !loans[item.binding]; return `<div class="db-number">${money(total ? 191200 : loan.payment)}</div><div class="db-inline"><span class="db-muted">15 ноября · ${total ? '2 кредита' : loan.name}</span>${link('loans.html','График')}</div>${item.h >= 3 ? '<p class="db-muted">Сумма платежей в ближайшую дату. Просроченные обязательства сюда не включены.</p>' : ''}`; }
      case 'debt': return `<div class="db-number">3 365 000 <small>₽</small></div><span class="db-muted">2 действующих кредита · только основной долг</span>${item.h >= 3 ? '<div class="db-inline db-bottom"><span class="db-muted">Без процентов и дополнительных начислений</span>'+link('loans.html','Мои кредиты')+'</div>' : ''}`;
      case 'quick': { const actions = item.binding === 'plans' ? [['calendar','Создать план','planner.html'],['loans','Мои кредиты','loans.html'],['support','Написать банку','support.html']] : [['catalog','Подать заявку','catalog.html'],['calendar','Создать план','planner.html'],['support','Написать банку','support.html']]; return `<div class="db-actions">${actions.map(([symbol,title,href])=>`<a href="${href}"><span class="db-icon">${icon(symbol)}</span>${title}</a>`).join('')}</div>${item.h >= 3 ? '<p class="db-muted db-bottom">Частые действия — без поиска по разделам.</p>' : ''}`; }
      case 'calendar': {
        const paymentDays = item.binding === 'plans' ? [] : [15,25];
        const planDays = item.binding === 'payments' ? [] : [6,18,25];
        return `<div class="db-inline"><strong style="font-size:13px">Ноябрь 2026</strong><span class="db-muted">${item.binding === 'payments' ? 'Платежи' : item.binding === 'plans' ? 'Планы' : 'Платежи и планы'}</span></div><div class="db-calendar">${['Пн','Вт','Ср','Чт','Пт','Сб','Вс'].map(day=>`<span class="weekday">${day}</span>`).join('')}${'<span></span>'.repeat(6)}${Array.from({length:30},(_,i)=>`<span class="${i+1 === 15 ? 'selected' : ''}">${i+1}<i>${paymentDays.includes(i+1)?'<b></b>':''}${planDays.includes(i+1)?'<b class="plan-dot"></b>':''}</i></span>`).join('')}</div><div class="db-inline db-bottom"><div class="db-legend"><span>Платёж</span><span>План</span></div>${link('planner.html','Открыть')}</div>`;
      }
      case 'plans': return `${(item.binding === 'ideas' ? [['—','ИДЕЯ','Новая точка','900 000 ₽']] : [['06','НОЯ','Обновить витрину','250 000 ₽'],['18','НОЯ','Печь для выпечки','600 000 ₽'],['25','НОЯ','Запас сырья','180 000 ₽']]).map(([day,month,title,amount])=>`<div class="db-plan"><span class="db-plan-date">${day}<small>${month}</small></span><div><strong>${title}</strong><span class="db-muted">${day === '—' ? 'Без даты' : 'План развития'}</span></div><span>${amount}</span></div>`).join('')}<div class="db-inline db-bottom"><span class="db-muted">${item.binding === 'ideas' ? 'Идеи без даты' : 'Ещё 1 идея без даты'}</span>${link('planner.html','Все планы')}</div>`;
      case 'application': return `<span class="db-muted">${application.number}</span><div class="db-number">${money(application.amount)}</div>${badge(application.status,application.status === 'Нужна информация'?'warning':'')}<div class="db-inline db-bottom"><span class="db-muted">${application.status === 'Нужна информация' ? 'Уточнить адрес склада' : 'Открыть подробности'}</span>${link('applications.html','Открыть')}</div>`;
      case 'loans': return item.binding === 'closed' ? '<p class="db-muted">В демонстрационном наборе закрытых кредитов нет.</p>' : `<div class="db-mini-apps">${Object.values(loans).map(credit=>`<div class="db-mini-app"><strong>${credit.name}</strong><span class="db-muted">${credit.number}</span><span class="db-app-amount">${money(credit.balance)}</span><span class="db-muted">Основной долг</span></div>`).join('')}</div>${link('loans.html','Открыть кредиты')}`;
      case 'payments': return item.binding === '7' ? '<p class="db-muted">С 22 по 28 октября платежей не запланировано.</p>' : `<div class="db-number">191 200 ₽</div><p class="db-muted">22 октября — 20 ноября · по всем кредитам</p><div class="db-schedule-row"><span>15 ноября</span><i></i><strong>191 200 ₽</strong></div><p class="db-muted">2 платежа в эту дату</p>${link('loans.html','Открыть графики')}`;
      case 'offers': return `<div class="db-mini-app"><div class="db-inline"><strong>Оборотный кредит</strong>${badge('Одобрено','success')}</div><span class="db-number">1 500 000 ₽</span><span class="db-muted">21,5% годовых · 24 месяца</span><span class="db-muted">Предложение до 30 октября</span>${link('applications.html','Рассмотреть условия')}</div>${item.binding === 'all' ? `<p class="db-muted">Ещё одно предложение истекло 18 октября.</p>` : ''}<p class="db-muted">Условия банка, деньги ещё не выданы.</p>`;
      case 'drafts': return `<div class="db-mini-app"><strong>Инвестиционный кредит</strong><span class="db-app-amount">3 200 000 ₽</span>${badge('Черновик')}<span class="db-muted">Последнее изменение 22 октября</span>${link('applications.html','Продолжить заявку')}</div>`;
      case 'planFunding': return `<strong>${item.binding === 'store' ? 'Новая точка' : 'Печь для выпечки'}</strong><div class="db-number">${item.binding === 'store' ? '900 000' : '600 000'} ₽</div><p class="db-muted">Ориентир бюджета плана</p>${item.binding === 'store' ? '<p class="db-muted">Для этого плана пока нет рекомендации.</p>' : '<div class="db-mini-app"><strong>Кредит на оборудование</strong><p class="db-muted">Вариант по цели и бюджету. Это не одобрение банка.</p></div>'}${link('catalog.html','Изучить каталог продуктов')}`;
      case 'notifications': return `<div class="db-mini-app"><strong>Банку нужны пояснения</strong><span class="db-muted">Уточните адрес склада в заявке</span>${badge('Не прочитано','warning')}${link('applications.html','Открыть заявку')}</div><div class="db-mini-app"><strong>Скоро платёж по кредитам</strong><span class="db-muted">15 ноября · 191 200 ₽</span>${link('loans.html','Посмотреть кредиты')}</div>${item.binding === 'all' ? '<p class="db-muted">Также показаны ранее прочитанные события.</p>' : ''}`;
      case 'tickets': return `<div class="db-mini-app"><div class="db-inline"><strong>Уточнение адреса склада</strong>${badge('Есть ответ','success')}</div><p class="db-muted">Банк: адрес обновлён. Можно продолжить заявку.</p>${link('support.html','Открыть поддержку')}</div>${item.binding !== 'answered' ? '<p class="db-muted">Ещё одно открытое обращение ожидает ответа.</p>' : ''}`;
      case 'funding': return `<strong class="db-funding-title">Большие планы?<br>Найдём решение.</strong><p class="db-muted">Продукты для оборота, оборудования и развития.</p>${link('catalog.html','Подобрать финансирование')}<p class="db-muted">Условия и доступность зависят от продукта и решения банка.</p>`;
      default: return '';
    }
  };
  const render = () => {
    board.dataset.editing = String(editing);
    $('#db-editbar').hidden = !editing;
    $('#db-start').hidden = editing;
    $('#db-view-add').hidden = editing;
    const positions = visualLayout();
    board.innerHTML = positions.sort((a,b)=>a.y-b.y || a.x-b.x).map(item => {
      const type = types[item.type];
      return `<section class="db-tile ${item.type}" data-id="${esc(item.id)}" data-type="${item.type}" data-h="${item.h}" data-editing="${editing}" style="grid-column:${item.x+1}/span ${item.w};grid-row:${item.y+1}/span ${item.h}" aria-label="${type.title}">
        ${editing ? `<div class="db-tools" aria-label="Настройка: ${type.title}"><button draggable="true" data-drag="${esc(item.id)}" title="Перетащить виджет" aria-label="Перетащить: ${type.title}">⠿</button><button data-settings="${esc(item.id)}" title="Содержимое и размер" aria-label="Настроить: ${type.title}">${icon('settings')}</button><button data-remove="${esc(item.id)}" title="Убрать виджет" aria-label="Убрать: ${type.title}">×</button></div>` : ''}
        <div class="db-tile-head"><span class="db-icon">${icon(type.icon)}</span><h2>${type.title}</h2></div><div class="db-body">${content(item)}</div></section>`;
    }).join('');
    const last = Math.max(0,...positions.map(item=>item.y+item.h));
    board.insertAdjacentHTML('beforeend',`<button class="db-add-slot" data-catalog style="grid-column:span ${columns()};grid-row:${last+1}/span 2"><span aria-hidden="true">＋</span>Добавить виджет<small>Выберите данные, которые важны именно вам</small></button>`);
    if (mobile()) board.style.gridTemplateRows = '';
  };
  const begin = () => { if (!editing) {snapshot = clone(widgets); editing = true; render();} };
  const catalog = $('#db-catalog-dialog');
  const settings = $('#db-settings-dialog');
  const catalogRender = () => {
    const query = $('#db-search').value.toLocaleLowerCase('ru-RU').trim();
    const entries = Object.entries(types).filter(([,type])=>(type.title+' '+type.description).toLocaleLowerCase('ru-RU').includes(query));
    $('#db-catalog-items').innerHTML = entries.map(([key,type])=>`<article class="db-catalog-card"><div class="db-catalog-preview ${key}">${icon(type.icon)}<strong>${type.preview}</strong></div><h3>${type.title}</h3><p>${type.description}</p><span class="db-muted">Размеры: ${type.sizes.map(([w,h])=>`${w} × ${h}`).join(' · ')}</span><button class="db-button" data-add="${key}">＋ Добавить</button></article>`).join('') || '<p class="db-muted">Ничего не найдено. Попробуйте другое название.</p>';
  };
  const showCatalog = () => {begin(); $('#db-search').value=''; catalogRender(); catalog.showModal();};
  const bindingOptions = item => {
    switch (types[item.type].binding) {
      case 'portfolio': return [['active','Действующие'],['closed','Закрытые']];
      case 'period': return [['30','Ближайшие 30 дней'],['7','Ближайшие 7 дней']];
      case 'offers': return [['current','Действующие и принятые'],['all','Включая истёкшие']];
      case 'planFunding': return [['oven','Печь для выпечки'],['store','Новая точка']];
      case 'notifications': return [['unread','Непрочитанные'],['all','Последние события']];
      case 'tickets': return [['active','Открытые и с ответом'],['answered','С ответом банка'],['all','Последние обращения']];
      case 'loan': return Object.entries(loans).map(([id,loan])=>[id,loan.name+' · '+loan.number]);
      case 'scope': return [['all','Все действующие кредиты'],...Object.entries(loans).map(([id,loan])=>[id,loan.name+' · '+loan.number])];
      case 'application': return Object.entries(applications).map(([id,app])=>[id,app.name+' · '+app.number]);
      case 'filter': return [['all','Все заявки'],['attention','Требуют внимания']];
      case 'events': return [['all','Платежи и планы'],['payments','Только платежи'],['plans','Только планы']];
      case 'plans': return [['all','Ближайшие планы'],['ideas','Идеи без даты']];
      case 'actions': return [['standard','Заявка · План · Написать банку'],['plans','План · Кредиты · Написать банку']];
      default: return [];
    }
  };
  const openSettings = id => {
    const item = widgets.find(value=>value.id === id);
    if (!item) return;
    selectedId=id;
    $('#db-settings-title').textContent=types[item.type].title;
    $('#db-size').innerHTML=types[item.type].sizes.map(([w,h])=>`<option value="${w},${h}" ${item.w===w && item.h===h ? 'selected' : ''}>${w} × ${h} клеток</option>`).join('');
    const options=bindingOptions(item);
    $('#db-binding-field').hidden=!options.length;
    $('#db-binding').innerHTML=options.map(([value,label])=>`<option value="${value}" ${item.binding===value ? 'selected' : ''}>${esc(label)}</option>`).join('');
    $('#db-position').innerHTML=ordered().map((entry,index)=>`<option value="${index}" ${entry.id===id ? 'selected' : ''}>${index+1} / ${widgets.length}</option>`).join('');
    $('#db-state').value=item.state || 'normal';
    settings.showModal();
  };
  $('#db-start').addEventListener('click',begin);
  $('#db-cancel').addEventListener('click',()=>{widgets=clone(snapshot);editing=false;render();status('Настройка отменена.');});
  $('#db-save').addEventListener('click',()=>{
    editing=false; render();
    try {localStorage.setItem(storageKey,JSON.stringify(widgets));status('Пример раскладки сохранён в этом браузере.');}
    catch {status('Раскладка применена на время просмотра. Локальное сохранение недоступно.');}
  });
  $('#db-reset').addEventListener('click',()=>{widgets=clone(defaults);render();status('Восстановлен исходный пример. Сохраните или отмените настройку.');});
  $('#db-edit-add').addEventListener('click',showCatalog);
  $('#db-view-add').addEventListener('click',showCatalog);
  $('#db-search').addEventListener('input',catalogRender);
  document.addEventListener('click',event=>{
    const target=event.target.closest('button');
    if (!target) return;
    if (target.hasAttribute('data-catalog')) showCatalog();
    if (target.dataset.settings) openSettings(target.dataset.settings);
    if (target.dataset.remove) {widgets=widgets.filter(item=>item.id !== target.dataset.remove);render();status('Виджет убран. Отмена вернёт исходный набор.');}
    if (target.dataset.retry) {const item=widgets.find(value=>value.id === target.dataset.retry);if (item) {item.state='normal';render();status('Показан пример успешной загрузки. Запрос к серверу не выполнялся.');}}
    if (target.dataset.add) {
      if (widgets.length >= 30) {status('В примере можно разместить до 30 виджетов.');return;}
      const type=types[target.dataset.add];
      const item={id:'demo-'+(++nextId),type:target.dataset.add,w:type.sizes[0][0],h:type.sizes[0][1]};
      const options=bindingOptions(item);item.binding=options[0]?.[0];
      widgets.push(firstSpace(item,widgets,12));catalog.close();render();openSettings(item.id);
    }

  });
  $('#db-settings-form').addEventListener('submit',event=>{
    event.preventDefault();const item=widgets.find(value=>value.id === selectedId);if (!item) return;
    const originalIndex=ordered().findIndex(entry=>entry.id===item.id);
    const nextIndex=Number($('#db-position').value);
    const [w,h]=$('#db-size').value.split(',').map(Number);
    item.w=w;item.h=h;item.x=Math.min(item.x,12-w);item.binding=$('#db-binding').value;item.state=$('#db-state').value;
    pack({...item});
    if (originalIndex!==nextIndex) {
      const sequence=ordered().filter(entry=>entry.id!==item.id);sequence.splice(nextIndex,0,item);
      const placed=[];sequence.forEach(entry=>placed.push(firstSpace(entry,placed,12)));widgets=placed;
    }
    settings.close();render();status('Настройки виджета применены.');
  });
  document.querySelectorAll('[data-close-dialog]').forEach(button=>button.addEventListener('click',()=>button.closest('dialog').close()));
  [catalog,settings].forEach(dialog=>dialog.addEventListener('close',()=>{
    if (!catalog.open && !settings.open) $('#db-edit-add').focus();
  }));
  board.addEventListener('dragstart',event=>{
    const handle=event.target.closest('[data-drag]');if (!editing || !handle) {event.preventDefault();return;}
    draggedId=handle.dataset.drag;event.dataTransfer.setData('text/plain',draggedId);event.dataTransfer.effectAllowed='move';handle.closest('.db-tile').classList.add('is-moving');
  });
  board.addEventListener('dragover',event=>{
    if (!draggedId) return;event.preventDefault();event.dataTransfer.dropEffect='move';
    board.querySelectorAll('.is-target').forEach(tile=>tile.classList.remove('is-target'));
    event.target.closest('.db-tile')?.classList.add('is-target');
  });
  board.addEventListener('drop',event=>{
    if (!draggedId) return;event.preventDefault();const item=widgets.find(value=>value.id === draggedId);if (!item) return;
    if (isTablet()) {
      const target=event.target.closest('[data-id]');const sequence=ordered().filter(value=>value.id !== item.id);const index=sequence.findIndex(value=>value.id === target?.dataset.id);
      sequence.splice(index<0?sequence.length:index,0,item);const placed=[];sequence.forEach(value=>placed.push(firstSpace(value,placed,12)));widgets=placed;
    } else {
      const rect=board.getBoundingClientRect();const col=(rect.width+16)/12;
      item.x=Math.max(0,Math.min(12-item.w,Math.floor((event.clientX-rect.left)/col)));
      item.y=Math.max(0,Math.min(200,Math.floor((event.clientY-rect.top)/80)));pack({...item});
    }
    draggedId=null;render();status('Виджет перемещён. Соседние карточки заняли свободные клетки.');
  });
  board.addEventListener('dragend',()=>{draggedId=null;board.querySelectorAll('.is-moving,.is-target').forEach(tile=>tile.classList.remove('is-moving','is-target'));});
  let resizeTimer;window.addEventListener('resize',()=>{clearTimeout(resizeTimer);resizeTimer=setTimeout(render,120);});
  pack();render();
})();

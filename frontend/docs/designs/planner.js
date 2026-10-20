/* Calendar and planning previews only. No recommendations service or notification delivery. */
(() => {
  const root = document.querySelector('#planner-calendar');
  if (!root) return;
  const icon = name => `<svg class="icon" aria-hidden="true"><use href="#icon-${name}"/></svg>`;
  const safe = value => String(value).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const money = value => new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(value);
  const plans = {
    repair: {name:'Ремонт помещения',budget:2400000,start:'2027-06-01',end:'2027-08-31',period:'1 июня — 31 августа 2027',note:'Обновить торговый зал и входную группу. Заложить расходы на материалы, работы и резерв на время ремонта.',goal:'Развитие бизнеса',icon:'building',tracking:'Подбор начнётся 1 июня 2027',explanation:'Пока план впереди. Когда начнётся летний период, на главной смогут появиться варианты финансирования ремонта.',badge:'На будущее',email:true},
    stock: {name:'Запасы к зимнему сезону',budget:650000,start:'2026-11-01',end:'2026-11-30',period:'1–30 ноября 2026',note:'Закупить муку, упаковку и сезонные ингредиенты. Увеличить запас перед праздничным спросом.',goal:'Оборотные средства',icon:'applications',tracking:'Подбор начнётся 1 ноября 2026',explanation:'С начала периода система сможет показывать варианты для оборотных средств, подходящие по цели и примерной сумме.',badge:'Скоро',email:false},
    oven: {name:'Новая печь для пекарни',budget:1200000,start:null,end:null,period:'Без даты · идея для развития',note:'Увеличить выпуск и добавить новую линейку хлеба. Пока изучаю оборудование и оцениваю окупаемость.',goal:'Оборудование',icon:'catalog',tracking:'Подбор можно показывать уже сейчас',explanation:'Идея не привязана к календарю. Цель и бюджет помогают подбирать варианты; период можно добавить позже.',badge:'Идея без даты',email:false}
  };
  let selected = 'repair', year = 2027, month = 5;
  const today = '2026-10-20';
  function iso(date) { return date.toISOString().slice(0,10); }
  function planCard(key) {
    const p = plans[key];
    return `<article class="plan-list-card"><div class="selected-plan-head"><span class="plan-category-icon">${icon(p.icon)}</span><span class="ui-badge ${p.start ? 'ui-badge--info':'ui-badge--neutral'}">${p.badge}</span></div><h2>${p.name}</h2><div class="plan-period">${icon('calendar')}${p.period}</div><p class="plan-note">${p.note}</p><div class="plan-budget"><small>Ориентир бюджета</small><strong class="ui-number">${money(p.budget)}</strong></div><button type="button" class="inline-action" data-plan-card="${key}">Посмотреть план ${icon('arrow')}</button></article>`;
  }
  document.querySelector('#all-plan-cards').innerHTML = Object.keys(plans).map(planCard).join('');
  document.querySelector('#idea-plan-cards').innerHTML = planCard('oven');
  function renderSelected() {
    const p = plans[selected];
    document.querySelector('#selected-plan').innerHTML = `<div class="selected-plan-head"><span class="plan-category-icon">${icon(p.icon)}</span><span class="ui-badge ${p.start ? 'ui-badge--info' : 'ui-badge--neutral'}">${p.badge}</span></div><h2>${p.name}</h2><div class="plan-period">${icon('calendar')}${p.period}</div><div class="plan-budget"><small>Ориентир бюджета · ${p.goal}</small><strong class="ui-number">${money(p.budget)}</strong></div><p class="plan-note">${p.note}</p><div class="plan-tracking"><span>${p.tracking}</span><p>${p.explanation}</p></div><div class="plan-notify-line"><span>В кабинете</span>${p.email?'<span>Письмо · настройка-пример</span>':''}</div><a class="text-link" href="home.html">Подбор на главной ${icon('arrow')}</a>`;
    document.querySelectorAll('[data-select-plan]').forEach(button => { const active = button.dataset.selectPlan === selected; button.classList.toggle('is-selected',active); button.setAttribute('aria-pressed',String(active)); });
  }
  function renderCalendar() {
    const first = new Date(Date.UTC(year,month,1));
    const shift = (first.getUTCDay()+6)%7;
    const count = new Date(Date.UTC(year,month+1,0)).getUTCDate();
    const cells = Math.ceil((count+shift)/7)*7;
    document.querySelector('#calendar-title').textContent = new Intl.DateTimeFormat('ru-RU',{month:'long',year:'numeric',timeZone:'UTC'}).format(first).replace(' г.','');
    const fragment = document.createDocumentFragment();
    let covered = 0;
    for(let n=0;n<cells;n++) {
      const date = new Date(Date.UTC(year,month,n-shift+1)), value = iso(date);
      const outside = date.getUTCMonth() !== month;
      const event = Object.entries(plans).find(([,p])=>p.start && value>=p.start && value<=p.end);
      if(event && !outside) covered++;
      const button = document.createElement('button');button.type='button';button.className=`calendar-day${outside?' is-outside':''}${event?' has-plan':''}${value===today?' is-today':''}`;
      const number = document.createElement('span');number.textContent=date.getUTCDate();button.append(number);
      button.setAttribute('aria-label',new Intl.DateTimeFormat('ru-RU',{day:'numeric',month:'long',year:'numeric',timeZone:'UTC'}).format(date)+(event?` · ${event[1].name}`:' · Добавить план'));
      if(event && (date.getUTCDate()===1 || date.getUTCDay()===1)) { const label=document.createElement('small');label.className='calendar-event';label.textContent=event[0]==='repair'?'Ремонт':'Запасы';button.append(label); }
      button.addEventListener('click',()=>{
        if(event) { selected=event[0];renderSelected(); }
        else { document.querySelector('#plan-start').value=value;document.querySelector('#plan-end').value=value;document.querySelector('[name="plan-date-mode"][value="period"]').checked=true;syncDateMode();const dialog=document.querySelector('#plan-dialog');dialog.showModal();dialog.addEventListener('close',()=>button.focus(),{once:true}); }
      });
      fragment.append(button);
    }
    root.replaceChildren(fragment);
    document.querySelector('#calendar-summary').textContent = covered ? `${covered} дней с планами` : 'На этот месяц планов нет';
  }
  function choosePlan(key) {
    selected=key;
    if(plans[key].start) { const date=new Date(plans[key].start+'T00:00:00Z');year=date.getUTCFullYear();month=date.getUTCMonth(); }
    renderSelected();renderCalendar();
  }
  document.querySelectorAll('[data-select-plan]').forEach(button=>button.addEventListener('click',()=>choosePlan(button.dataset.selectPlan)));
  document.querySelectorAll('[data-plan-card]').forEach(button=>button.addEventListener('click',()=>{choosePlan(button.dataset.planCard);document.querySelector('#tab-calendar').click();document.querySelector('#tab-calendar').focus();}));
  document.querySelector('#calendar-prev').addEventListener('click',()=>{month--;if(month<0){month=11;year--;}renderCalendar();});
  document.querySelector('#calendar-next').addEventListener('click',()=>{month++;if(month>11){month=0;year++;}renderCalendar();});
  document.querySelector('#calendar-today').addEventListener('click',()=>{year=2026;month=9;renderCalendar();});
  const start=document.querySelector('#plan-start'), end=document.querySelector('#plan-end');
  function syncDateMode() {
    const idea=document.querySelector('[name="plan-date-mode"]:checked').value==='idea';
    document.querySelector('#plan-date-fields').hidden=idea;start.required=!idea;start.disabled=idea;end.disabled=idea;
    end.setCustomValidity('');document.querySelector('#plan-preview').hidden=true;
  }
  document.querySelectorAll('[name="plan-date-mode"]').forEach(radio=>radio.addEventListener('change',syncDateMode));
  [start,end].forEach(field=>field.addEventListener('input',()=>end.setCustomValidity('')));
  document.querySelector('#plan-form').addEventListener('input',()=>{document.querySelector('#plan-preview').hidden=true;});
  document.querySelector('#plan-form').addEventListener('submit',event=>{
    event.preventDefault();
    const idea=document.querySelector('[name="plan-date-mode"]:checked').value==='idea';
    if(!idea && end.value && end.value<start.value){end.setCustomValidity('Окончание не может быть раньше начала.');end.reportValidity();return;}
    const name=document.querySelector('#plan-name').value.trim();
    if(!name){document.querySelector('#plan-name').focus();return;}
    const period=idea?'Без даты · подбор может появляться сразу':`${start.value}${end.value&&end.value!==start.value?' — '+end.value:''} · подбор с начала периода`;
    const output=document.querySelector('#plan-preview');
    output.innerHTML=`<span class="tile-label">ПРЕДПРОСМОТР · ПЛАН НЕ СОХРАНЁН</span><h3>${safe(name)}</h3><p>${safe(period)}</p><div class="plan-preview-budget ui-number">${money(Number(document.querySelector('#plan-budget').value))}</div><p>${safe(document.querySelector('#plan-note').value)}</p><p>Уведомления: в кабинете${document.querySelector('#plan-email').checked?' и письмом':''}. В макете ничего не отправляется.</p>`;
    output.hidden=false;output.scrollIntoView({block:'nearest'});
  });
  renderSelected();renderCalendar();
})();

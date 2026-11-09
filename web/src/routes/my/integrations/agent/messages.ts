import { defineMessages } from '#lib/i18n/t.js';

export const messages = defineMessages(
  {
    intro:
      'Use the Claude Code or Codex subscription you already pay for. Your agent reads jobs, tailors your CV and fills applications through the tools below; this server makes no AI calls of its own.',
    serverLabel: 'Your MCP server',
    copy: 'Copy',
    copied: 'Copied',
    connectTitle: '1. Connect your agent',
    claudeLabel: 'Claude Code',
    claudeAfter:
      'Then run /mcp inside Claude Code, choose keenu and Authenticate. A browser tab opens here; approve it.',
    codexLabel: 'Codex',
    codexAfter: 'The login opens a browser tab here; approve it.',
    cvTitle: '2. Give it your CV',
    cvBody:
      'Upload your CV as a PDF on your profile page, then ask your agent to import it. Only what the file actually says is kept; anything else is dropped and listed back to you.',
    tailorTitle: '3. Tailor a CV for a job',
    tailorBody:
      'Ask in your own words, for example the prompt below. Every claim the agent writes must point at your own experience. What it adds by itself stays unprintable until you press Confirm on your Experience page.',
    applyTitle: '4. Apply from your browser',
    applyBody:
      'Build the browser extension for this machine, load it in Chrome (chrome://extensions → Developer mode → Load unpacked → the folder below), sign in from its side panel, then open the application page and ask your agent to fill it. Nothing is ever submitted for you: you review the form and press Submit.',
    promptLabel: 'Example prompt',
    openProfile: 'Open your profile →',
  },
  {
    ru: {
      intro:
        'Используйте подписку Claude Code или Codex, за которую вы уже платите. Ваш агент читает вакансии, адаптирует резюме и заполняет отклики через инструменты ниже; этот сервер сам не обращается к ИИ.',
      serverLabel: 'Ваш MCP-сервер',
      copy: 'Копировать',
      copied: 'Скопировано',
      connectTitle: '1. Подключите агента',
      claudeLabel: 'Claude Code',
      claudeAfter:
        'Затем выполните /mcp в Claude Code, выберите keenu и Authenticate. Здесь откроется вкладка браузера — подтвердите доступ.',
      codexLabel: 'Codex',
      codexAfter: 'Вход откроет здесь вкладку браузера — подтвердите доступ.',
      cvTitle: '2. Передайте резюме',
      cvBody:
        'Загрузите резюме в PDF на странице профиля и попросите агента импортировать его. Сохраняется только то, что действительно написано в файле; остальное отбрасывается, и агент вам об этом скажет.',
      tailorTitle: '3. Адаптируйте резюме под вакансию',
      tailorBody:
        'Попросите своими словами, например как ниже. Каждое утверждение агента должно опираться на ваш опыт. То, что агент добавил сам, не попадёт в резюме, пока вы не нажмёте «Подтвердить» на странице опыта.',
      applyTitle: '4. Откликайтесь из браузера',
      applyBody:
        'Соберите расширение для этого компьютера, загрузите его в Chrome (chrome://extensions → Режим разработчика → Загрузить распакованное → папка ниже), войдите в боковой панели, откройте страницу отклика и попросите агента заполнить её. Ничего не отправляется за вас: вы проверяете форму и нажимаете «Отправить».',
      promptLabel: 'Пример запроса',
      openProfile: 'Открыть профиль →',
    },
  },
);

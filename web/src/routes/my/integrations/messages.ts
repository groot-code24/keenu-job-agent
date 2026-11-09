import { defineMessages } from '#lib/i18n/t.js';

export const messages = defineMessages(
  {
    title: 'Integrations',
    description: 'Connect your other accounts, and let your own tools reach Keenu.',
    tabsLabel: 'Integration sections',
    tabs: { connections: 'Connections', agent: 'Your AI agent', 'api-keys': 'API keys', webhook: 'Webhook' },
  },
  {
    ru: {
      title: 'Интеграции',
      description: 'Подключите другие аккаунты и дайте своим инструментам доступ к Keenu.',
      tabsLabel: 'Разделы интеграций',
      tabs: { connections: 'Подключения', agent: 'Ваш ИИ-агент', 'api-keys': 'API-ключи', webhook: 'Вебхук' },
    },
  },
);

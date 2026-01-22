/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    {
      type: 'doc',
      id: 'intro',
      label: '🚀 Introducción',
    },
    {
      type: 'category',
      label: '🏁 Primeros Pasos',
      collapsed: false,
      items: [
        'installation',
        'licensing',
      ],
    },
    {
      type: 'category',
      label: '💵 Gestión Financiera',
      collapsed: false,
      items: [
        'dashboard',
        'accounts',
        'categories',
        'transactions',
        'recurring',
        'reconciliation',
      ],
    },
    {
      type: 'category',
      label: '🏛️ Facturación Electrónica',
      collapsed: false,
      items: [
        'taxpayers',
        'sri-setup',
        'issuing-invoices',
        'credit-notes',
      ],
    },
    {
      type: 'category',
      label: '🛠️ Administración y Control',
      collapsed: true,
      items: [
        'reports-overview',
        'users',
      ],
    },
  ],
};

module.exports = sidebars;
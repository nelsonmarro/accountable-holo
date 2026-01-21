/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'Comenzando',
      items: ['intro', 'installation', 'dashboard', 'licensing'],
      collapsed: false,
    },
    {
      type: 'category',
      label: 'Gestión Financiera',
      items: ['accounts', 'categories', 'transactions', 'recurring', 'reconciliation'],
    },
    {
      type: 'category',
      label: 'Facturación Electrónica (SRI)',
      items: ['taxpayers', 'sri-setup', 'issuing-invoices', 'credit-notes'],
    },
    {
      type: 'category',
      label: 'Reportes y Análisis',
      items: ['reports-overview'],
    },
    {
      type: 'category',
      label: 'Administración',
      items: ['users'],
    },
  ],
};

module.exports = sidebars;

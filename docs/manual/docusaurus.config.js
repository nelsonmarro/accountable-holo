// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const {themes} = require('prism-react-renderer');
const lightCodeTheme = themes.github;
const darkCodeTheme = themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Verith',
  tagline: 'Gestión Financiera y Facturación SRI Simplificada',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://docs.verith.naphsoft.dev',
  baseUrl: '/',

  // GitHub pages deployment config.
  organizationName: 'nelsonmarro', 
  projectName: 'verith', 

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Habilitar Mermaid en Markdown
  markdown: {
    mermaid: true,
  },

  // Activar el tema de Mermaid
  themes: ['@docusaurus/theme-mermaid'],

  i18n: {
    defaultLocale: 'es',
    locales: ['es'],
  },

  plugins: [
    [
      require.resolve("@easyops-cn/docusaurus-search-local"),
      /** @type {import("@easyops-cn/docusaurus-search-local").Options} */
      ({
        hashed: true,
        language: ["es"],
        docsRouteBasePath: "/",
      }),
    ],
  ],

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          routeBasePath: '/', // Serve the docs at the site's root
        },
        blog: false, // Desactivamos el blog para un manual de usuario
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: {
        defaultMode: 'dark',
        disableSwitch: true,
        respectPrefersColorScheme: false,
      },
      navbar: {
        title: 'Verith Manual',
        logo: {
          alt: 'Verith Logo',
          src: 'img/logo.png',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'tutorialSidebar',
            position: 'left',
            label: 'Documentación',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Manual de Usuario',
                to: '/',
              },
            ],
          },
          {
            title: 'Comunidad',
            items: [
              {
                label: 'Soporte',
                href: 'mailto:soporte@verith.app',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Verith. Construido con Docusaurus.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      // Configuración visual de Mermaid
      mermaid: {
        theme: {light: 'neutral', dark: 'dark'},
      },
      tableOfContents: {
        minHeadingLevel: 2,
        maxHeadingLevel: 4,
      },
    }),
};

module.exports = config;

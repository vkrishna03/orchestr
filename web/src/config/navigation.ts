export interface NavItem {
  title: string
  href: string
  items?: NavItem[]
}

export const docsNav: NavItem[] = [
  {
    title: "Getting Started",
    href: "/docs",
    items: [
      { title: "Introduction", href: "/docs" },
      { title: "Installation", href: "/docs/installation" },
      { title: "Quick Start", href: "/docs/quickstart" },
    ],
  },
  {
    title: "Guide",
    href: "/docs/guide",
    items: [
      { title: "Orchestrfile", href: "/docs/guide/orchestrfile" },
      { title: "Service Detection", href: "/docs/guide/detection" },
      { title: "Test Patterns", href: "/docs/guide/patterns" },
    ],
  },
  {
    title: "Reference",
    href: "/docs/reference",
    items: [
      { title: "CLI Commands", href: "/docs/reference/cli" },
      { title: "Configuration", href: "/docs/reference/config" },
    ],
  },
  {
    title: "Contributing",
    href: "/docs/contributing",
  },
]

export const mainNav = [
  { title: "Docs", href: "/docs" },
  { title: "GitHub", href: "https://github.com/vkrishna03/orchestr" },
]

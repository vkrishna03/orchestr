# Orchestr Web

Documentation and marketing site for Orchestr CLI.

## Project Structure

```
web/
├── src/
│   ├── components/
│   │   ├── ui/           # shadcn components
│   │   └── layout/       # Header, Footer (shared)
│   ├── modules/
│   │   ├── landing/      # Landing page
│   │   │   ├── components/
│   │   │   └── index.tsx
│   │   └── docs/         # Documentation
│   │       ├── components/
│   │       ├── content/  # MDX files
│   │       └── index.tsx
│   ├── config/           # Navigation config
│   ├── lib/              # Utilities
│   ├── types/            # TypeScript types
│   ├── App.tsx           # Routes
│   └── main.tsx          # Entry point
├── public/
├── vite.config.ts
└── package.json
```

## Tech Stack

- React 19 + Vite 7
- Tailwind CSS v4 + shadcn/ui
- react-router-dom v7
- MDX with remark/rehype plugins
- JetBrains Mono font

## Commands

```bash
npm run dev       # Development server
npm run build     # Production build
npm run preview   # Preview build
```

## Adding Pages

1. Create MDX file in `src/modules/docs/content/`
2. Import in `App.tsx`
3. Add route
4. Update navigation in `src/config/navigation.ts`

## shadcn Components

```bash
npx shadcn@latest add <component>
```

**Installed:** button, card, badge, input, textarea, dropdown-menu, label, separator, select, alert-dialog, combobox, field, input-group

## Current State

- [x] Vite + React setup
- [x] shadcn configured
- [x] Tailwind v4
- [x] MDX configured
- [x] Routing setup
- [x] Layout components
- [x] Docs components
- [x] Landing page
- [x] Basic docs content

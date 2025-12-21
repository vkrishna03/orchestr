import { Outlet } from "react-router-dom"
import { MDXProvider } from "@mdx-js/react"
import { Sidebar } from "./components/Sidebar"
import { mdxComponents } from "./components/MDXComponents"

export function DocsLayout() {
  return (
    <div className="container max-w-6xl px-4 py-8">
      <div className="flex gap-8">
        <Sidebar />
        <article className="flex-1 min-w-0 max-w-3xl">
          <MDXProvider components={mdxComponents}>
            <Outlet />
          </MDXProvider>
        </article>
      </div>
    </div>
  )
}

export { Sidebar } from "./components/Sidebar"
export { CodeBlock } from "./components/CodeBlock"
export { Callout } from "./components/Callout"

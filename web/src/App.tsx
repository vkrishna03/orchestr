import { BrowserRouter, Routes, Route } from "react-router-dom"
import { RootLayout } from "@/components/layout"
import { Landing } from "@/modules/landing"
import { DocsLayout } from "@/modules/docs"

// MDX content imports
import DocsIndex from "@/modules/docs/content/index.mdx"
import Installation from "@/modules/docs/content/installation.mdx"
import QuickStart from "@/modules/docs/content/quickstart.mdx"
import Contributing from "@/modules/docs/content/contributing.mdx"

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<RootLayout />}>
          <Route path="/" element={<Landing />} />
          <Route path="/docs" element={<DocsLayout />}>
            <Route index element={<DocsIndex />} />
            <Route path="installation" element={<Installation />} />
            <Route path="quickstart" element={<QuickStart />} />
            <Route path="contributing" element={<Contributing />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App

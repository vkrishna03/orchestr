import { Link } from "react-router-dom"
import { mainNav } from "@/config/navigation"
import { Button } from "@/components/ui/button"
import { AnimatedThemeToggler } from "@/components/ui/animated-theme-toggler"
import { Terminal, Github, Menu, X } from "lucide-react"
import { useState } from "react"

export function Header() {
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur">
      <div className="container flex h-14 max-w-6xl items-center px-4">
        <Link to="/" className="flex items-center gap-2 mr-8">
          <Terminal className="h-5 w-5 text-primary" />
          <span className="font-bold">orchestr</span>
        </Link>

        <nav className="hidden md:flex items-center gap-6 text-sm flex-1">
          {mainNav.map((item) =>
            item.href.startsWith("http") ? (
              <a
                key={item.href}
                href={item.href}
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                {item.title}
              </a>
            ) : (
              <Link
                key={item.href}
                to={item.href}
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                {item.title}
              </Link>
            )
          )}
        </nav>

        <div className="hidden md:flex items-center gap-3">
          <AnimatedThemeToggler className="inline-flex items-center justify-center size-8 rounded-md hover:bg-muted transition-colors [&_svg]:size-4" />
          <a
            href="https://github.com/vkrishna03/orchestr"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center size-8 rounded-md hover:bg-muted transition-colors"
          >
            <Github className="h-4 w-4" />
          </a>
          <Link to="/docs">
            <Button size="sm">Get Started</Button>
          </Link>
        </div>

        <Button
          variant="ghost"
          size="icon"
          className="md:hidden ml-auto"
          onClick={() => setMobileOpen(!mobileOpen)}
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </Button>
      </div>

      {mobileOpen && (
        <nav className="md:hidden border-t border-border/40 p-4 flex flex-col gap-3">
          {mainNav.map((item) => (
            <Link
              key={item.href}
              to={item.href}
              className="text-muted-foreground hover:text-foreground"
              onClick={() => setMobileOpen(false)}
            >
              {item.title}
            </Link>
          ))}
          <Link to="/docs" onClick={() => setMobileOpen(false)}>
            <Button size="sm">Get Started</Button>
          </Link>
        </nav>
      )}
    </header>
  )
}

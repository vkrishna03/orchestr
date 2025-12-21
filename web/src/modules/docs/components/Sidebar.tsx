import { Link, useLocation } from "react-router-dom"
import { docsNav } from "@/config/navigation"
import { cn } from "@/lib/utils"
import { ChevronRight } from "lucide-react"

export function Sidebar() {
  const location = useLocation()

  return (
    <aside className="w-64 shrink-0 hidden lg:block">
      <nav className="sticky top-20 space-y-6 pr-4">
        {docsNav.map((section) => (
          <div key={section.href}>
            <h4 className="font-medium text-sm mb-2">{section.title}</h4>
            {section.items ? (
              <ul className="space-y-1">
                {section.items.map((item) => (
                  <li key={item.href}>
                    <Link
                      to={item.href}
                      className={cn(
                        "flex items-center gap-2 text-sm py-1.5 px-2 rounded-md transition-colors",
                        location.pathname === item.href
                          ? "bg-primary/10 text-primary font-medium"
                          : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
                      )}
                    >
                      <ChevronRight className="h-3 w-3" />
                      {item.title}
                    </Link>
                  </li>
                ))}
              </ul>
            ) : (
              <Link
                to={section.href}
                className={cn(
                  "flex items-center gap-2 text-sm py-1.5 px-2 rounded-md transition-colors",
                  location.pathname === section.href
                    ? "bg-primary/10 text-primary font-medium"
                    : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
                )}
              >
                <ChevronRight className="h-3 w-3" />
                {section.title}
              </Link>
            )}
          </div>
        ))}
      </nav>
    </aside>
  )
}

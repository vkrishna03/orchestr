import { cn } from "@/lib/utils"
import { AlertCircle, Info, AlertTriangle, CheckCircle } from "lucide-react"
import type { ReactNode } from "react"

type CalloutType = "info" | "warning" | "error" | "success"

interface CalloutProps {
  type?: CalloutType
  title?: string
  children: ReactNode
}

const icons = {
  info: Info,
  warning: AlertTriangle,
  error: AlertCircle,
  success: CheckCircle,
}

const styles = {
  info: "border-blue-500/30 bg-blue-500/10 text-blue-200",
  warning: "border-yellow-500/30 bg-yellow-500/10 text-yellow-200",
  error: "border-red-500/30 bg-red-500/10 text-red-200",
  success: "border-primary/30 bg-primary/10 text-primary",
}

export function Callout({ type = "info", title, children }: CalloutProps) {
  const Icon = icons[type]

  return (
    <div className={cn("my-4 rounded-lg border p-4", styles[type])}>
      <div className="flex items-start gap-3">
        <Icon className="h-5 w-5 mt-0.5 shrink-0" />
        <div className="flex-1">
          {title && <div className="font-medium mb-1">{title}</div>}
          <div className="text-sm opacity-90">{children}</div>
        </div>
      </div>
    </div>
  )
}

import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ArrowRight } from "lucide-react"

export function CTA() {
  return (
    <section className="py-20 border-t border-border/40">
      <div className="container max-w-6xl px-4">
        <div className="rounded-lg border border-border bg-card/50 p-8 md:p-12 text-center">
          <h2 className="text-2xl md:text-3xl font-bold tracking-tight mb-4">
            Ready to automate your testing?
          </h2>
          <p className="text-muted-foreground max-w-xl mx-auto mb-6">
            Get started with Orchestr in minutes. No complex setup, no configuration files.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link to="/docs">
              <Button size="lg">
                Read the Docs
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
          </div>
          <div className="mt-6 max-w-full overflow-x-auto">
            <code className="inline-block bg-zinc-950 text-zinc-300 text-sm px-4 py-2 rounded-lg whitespace-nowrap">
              <span className="text-emerald-400">$</span> curl -L https://github.com/vkrishna03/orchestr/releases/latest/download/orchestr-darwin -o orchestr
            </code>
          </div>
        </div>
      </div>
    </section>
  )
}

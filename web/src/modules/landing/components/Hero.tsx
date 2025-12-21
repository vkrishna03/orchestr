import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Terminal, ArrowRight, Github } from "lucide-react"

export function Hero() {
  return (
    <section className="py-20 md:py-32">
      <div className="container max-w-6xl px-4">
        <div className="flex flex-col items-center text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-border/60 bg-muted/30 px-4 py-1.5 text-sm text-muted-foreground mb-6">
            <Terminal className="h-4 w-4" />
            <span>Zero-config testing for AI-generated code</span>
          </div>

          <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6 max-w-3xl">
            Test your services{" "}
            <span className="text-primary">automatically</span>
          </h1>

          <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mb-8">
            Orchestr discovers your services, starts them in the right order,
            mocks external APIs, and runs functional tests with zero configuration.
          </p>

          <div className="flex flex-col sm:flex-row items-center gap-4">
            <Link to="/docs">
              <Button size="lg">
                Get Started
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
            <a
              href="https://github.com/vkrishna03/orchestr"
              target="_blank"
              rel="noopener noreferrer"
            >
              <Button size="lg" variant="outline">
                <Github className="mr-2 h-4 w-4" />
                View on GitHub
              </Button>
            </a>
          </div>
        </div>
      </div>
    </section>
  )
}

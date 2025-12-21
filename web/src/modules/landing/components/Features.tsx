import { Cpu, GitBranch, Zap, Shield, FileCode, BarChart3 } from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"

const features = [
  {
    icon: Cpu,
    title: "Auto-Discovery",
    description: "Detects Spring Boot, Vite, Express, and more. No config needed.",
  },
  {
    icon: GitBranch,
    title: "Dependency Graph",
    description: "Builds service dependencies and starts them in the right order.",
  },
  {
    icon: Zap,
    title: "Fast Execution",
    description: "Parallel service startup and test execution for quick feedback.",
  },
  {
    icon: Shield,
    title: "Auto-Mocking",
    description: "Automatically mocks Stripe, SendGrid, and other external APIs.",
  },
  {
    icon: FileCode,
    title: "Simple DSL",
    description: "Write tests in a simple, readable Orchestrfile format.",
  },
  {
    icon: BarChart3,
    title: "Detailed Reports",
    description: "Get clear pass/fail reports with distributed tracing.",
  },
]

export function Features() {
  return (
    <section className="py-20 border-t border-border/40">
      <div className="container max-w-6xl px-4">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold tracking-tight mb-4">
            Everything you need
          </h2>
          <p className="text-muted-foreground max-w-2xl mx-auto">
            From service discovery to test reporting, Orchestr handles the entire
            testing pipeline.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature) => (
            <Card key={feature.title} className="bg-card/50 border-border/60">
              <CardHeader>
                <feature.icon className="h-10 w-10 text-primary mb-2" />
                <CardTitle className="text-lg">{feature.title}</CardTitle>
                <CardDescription>{feature.description}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </div>
    </section>
  )
}

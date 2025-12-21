import { Terminal, AnimatedSpan } from "@/components/ui/terminal"

export function CodePreview() {
  return (
    <section className="py-20 border-t border-border/40">
      <div className="container max-w-6xl px-4">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold tracking-tight mb-4">
            Simple, readable tests
          </h2>
          <p className="text-muted-foreground max-w-2xl mx-auto">
            Write your test specs in a human-readable format. No complex setup required.
          </p>
        </div>

        <div className="flex justify-center">
          <Terminal className="max-w-3xl w-full max-h-none bg-zinc-950 text-zinc-100 border-zinc-800" sequence={false}>
            <AnimatedSpan className="text-zinc-500"># User registration flow</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">ENTITY:</span> User</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">ENDPOINT:</span> /api/register</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">METHOD:</span> POST</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">DATA:</span> {`{email: "test@example.com", password: "secret123"}`}</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">EXPECT:</span> 201 with {`{id: <uuid>, email: "test@example.com"}`}</AnimatedSpan>
            <AnimatedSpan className="text-zinc-600">---</AnimatedSpan>
            <AnimatedSpan className="text-zinc-500"># Login and save token</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">ENTITY:</span> User</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">ENDPOINT:</span> /api/login</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">METHOD:</span> POST</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">DATA:</span> {`{email: "test@example.com", password: "secret123"}`}</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">EXPECT:</span> 200 with {`{token: <jwt>}`}</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">SAVE:</span> token from body.token</AnimatedSpan>
            <AnimatedSpan className="text-zinc-600">---</AnimatedSpan>
            <AnimatedSpan className="text-zinc-500"># Authenticated request</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">ENTITY:</span> User</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">ENDPOINT:</span> /api/me</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">METHOD:</span> GET</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">AUTH:</span> {`Bearer {{token}}`}</AnimatedSpan>
            <AnimatedSpan><span className="text-emerald-400">EXPECT:</span> 200 with {`{email: "test@example.com"}`}</AnimatedSpan>
          </Terminal>
        </div>
      </div>
    </section>
  )
}

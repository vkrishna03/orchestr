import type { ComponentPropsWithoutRef } from "react"
import { CodeBlock } from "./CodeBlock"
import { Callout } from "./Callout"
import { cn } from "@/lib/utils"

export const mdxComponents = {
  h1: (props: ComponentPropsWithoutRef<"h1">) => (
    <h1 className="scroll-m-20 text-3xl font-bold tracking-tight mt-8 mb-4" {...props} />
  ),
  h2: (props: ComponentPropsWithoutRef<"h2">) => (
    <h2 className="scroll-m-20 text-2xl font-semibold tracking-tight mt-8 mb-4 border-b border-border pb-2" {...props} />
  ),
  h3: (props: ComponentPropsWithoutRef<"h3">) => (
    <h3 className="scroll-m-20 text-xl font-semibold tracking-tight mt-6 mb-3" {...props} />
  ),
  h4: (props: ComponentPropsWithoutRef<"h4">) => (
    <h4 className="scroll-m-20 text-lg font-semibold tracking-tight mt-4 mb-2" {...props} />
  ),
  p: (props: ComponentPropsWithoutRef<"p">) => (
    <p className="leading-7 [&:not(:first-child)]:mt-4" {...props} />
  ),
  ul: (props: ComponentPropsWithoutRef<"ul">) => (
    <ul className="my-4 ml-6 list-disc [&>li]:mt-2" {...props} />
  ),
  ol: (props: ComponentPropsWithoutRef<"ol">) => (
    <ol className="my-4 ml-6 list-decimal [&>li]:mt-2" {...props} />
  ),
  li: (props: ComponentPropsWithoutRef<"li">) => (
    <li className="leading-7" {...props} />
  ),
  blockquote: (props: ComponentPropsWithoutRef<"blockquote">) => (
    <blockquote className="mt-4 border-l-4 border-primary/50 pl-4 italic text-muted-foreground" {...props} />
  ),
  code: ({ className, children, ...props }: ComponentPropsWithoutRef<"code">) => {
    const isInline = !className?.includes("language-")
    if (isInline) {
      return (
        <code
          className="relative rounded bg-muted px-1.5 py-0.5 font-mono text-sm"
          {...props}
        >
          {children}
        </code>
      )
    }
    const language = className?.replace("language-", "")
    return <CodeBlock language={language}>{String(children)}</CodeBlock>
  },
  pre: ({ children }: ComponentPropsWithoutRef<"pre">) => <>{children}</>,
  a: ({ className, ...props }: ComponentPropsWithoutRef<"a">) => (
    <a
      className={cn("text-primary underline underline-offset-4 hover:text-primary/80", className)}
      {...props}
    />
  ),
  table: (props: ComponentPropsWithoutRef<"table">) => (
    <div className="my-6 w-full overflow-auto">
      <table className="w-full border-collapse text-sm" {...props} />
    </div>
  ),
  th: (props: ComponentPropsWithoutRef<"th">) => (
    <th className="border border-border px-4 py-2 text-left font-semibold bg-muted/50" {...props} />
  ),
  td: (props: ComponentPropsWithoutRef<"td">) => (
    <td className="border border-border px-4 py-2" {...props} />
  ),
  hr: () => <hr className="my-8 border-border" />,
  Callout,
  CodeBlock,
}

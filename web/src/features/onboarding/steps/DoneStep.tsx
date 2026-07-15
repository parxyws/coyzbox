import { motion } from "framer-motion"
import { IconCheck, IconBuildingSkyscraper, IconUsers, IconUser } from "@tabler/icons-react"
import { Button } from "@/components/ui/button"
import type { OnboardingFormData } from "../useOnboardingStore"

interface DoneStepProps {
  formData: OnboardingFormData
  isSubmitting: boolean
  onFinish: () => void
}

const containerVariants = {
  hidden: {},
  show: {
    transition: {
      staggerChildren: 0.12,
      delayChildren: 0.1,
    },
  },
}

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.35 } },
}

export function DoneStep({ formData, isSubmitting, onFinish }: DoneStepProps) {
  const { profile, workspace, invite } = formData

  const summaryItems = [
    {
      icon: <IconUser className="size-4 text-primary" />,
      label: "Profile",
      value: profile.displayName || profile.fullName || "—",
    },
    {
      icon: <IconBuildingSkyscraper className="size-4 text-primary" />,
      label: "Workspace",
      value: workspace.workspaceName || "—",
      sub: workspace.workspaceSlug ? `cozybox.app/${workspace.workspaceSlug}` : undefined,
    },
    {
      icon: <IconUsers className="size-4 text-primary" />,
      label: "Invites",
      value: invite.skipped
        ? "Skipped"
        : invite.emails.length > 0
          ? `${invite.emails.length} invite${invite.emails.length > 1 ? "s" : ""} sent`
          : "None",
    },
  ]

  return (
    <div className="flex flex-col items-center gap-6 py-2 text-center">
      {/* Success icon */}
      <motion.div
        initial={{ scale: 0, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ type: "spring", stiffness: 260, damping: 20 }}
        className="flex size-16 items-center justify-center rounded-full bg-primary/10"
      >
        <div className="flex size-12 items-center justify-center rounded-full bg-primary">
          <IconCheck className="size-6 text-primary-foreground" strokeWidth={3} />
        </div>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2, duration: 0.35 }}
        className="space-y-1"
      >
        <h2 className="text-lg font-semibold">You're all set, {profile.displayName || profile.fullName}! 🎉</h2>
        <p className="text-xs text-muted-foreground">
          Your workspace is ready. Here's a quick summary of what you've set up.
        </p>
      </motion.div>

      {/* Summary cards */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="show"
        className="w-full space-y-2"
      >
        {summaryItems.map((item) => (
          <motion.div
            key={item.label}
            variants={itemVariants}
            className="flex items-center gap-3 rounded-lg border border-border bg-muted/30 px-3 py-2.5 text-left"
          >
            <div className="flex size-7 shrink-0 items-center justify-center rounded-md border border-border bg-background">
              {item.icon}
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-xs font-medium leading-snug">{item.value}</p>
              {item.sub && (
                <p className="truncate text-[10px] text-muted-foreground">{item.sub}</p>
              )}
            </div>
            <span className="text-[10px] text-muted-foreground">{item.label}</span>
          </motion.div>
        ))}
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6, duration: 0.35 }}
        className="w-full"
      >
        <Button
          className="w-full"
          size="lg"
          onClick={onFinish}
          disabled={isSubmitting}
        >
          {isSubmitting ? "Setting things up…" : "Go to Dashboard →"}
        </Button>
      </motion.div>
    </div>
  )
}

import { useState } from "react"

export interface OnboardingProfileData {
  fullName: string
  displayName: string
  avatarPreview?: string
}

export interface OnboardingWorkspaceData {
  workspaceName: string
  workspaceSlug: string
  workspaceDescription?: string
}

export interface OnboardingInviteData {
  emails: string[]
  skipped: boolean
}

export interface OnboardingFormData {
  profile: OnboardingProfileData
  workspace: OnboardingWorkspaceData
  invite: OnboardingInviteData
}

const INITIAL_DATA: OnboardingFormData = {
  profile: {
    fullName: "",
    displayName: "",
    avatarPreview: undefined,
  },
  workspace: {
    workspaceName: "",
    workspaceSlug: "",
    workspaceDescription: "",
  },
  invite: {
    emails: [],
    skipped: false,
  },
}

export const STEPS = [
  { id: "profile", label: "Profile" },
  { id: "workspace", label: "Workspace" },
  { id: "invite", label: "Invite" },
  { id: "done", label: "Done" },
] as const

export type StepId = (typeof STEPS)[number]["id"]

export function useOnboardingStore() {
  const [currentStep, setCurrentStep] = useState(0)
  const [formData, setFormData] = useState<OnboardingFormData>(INITIAL_DATA)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const totalSteps = STEPS.length

  function goNext() {
    setCurrentStep((prev) => Math.min(prev + 1, totalSteps - 1))
  }

  function goPrev() {
    setCurrentStep((prev) => Math.max(prev - 1, 0))
  }

  function setProfileData(data: OnboardingProfileData) {
    setFormData((prev) => ({ ...prev, profile: data }))
  }

  function setWorkspaceData(data: OnboardingWorkspaceData) {
    setFormData((prev) => ({ ...prev, workspace: data }))
  }

  function setInviteData(data: OnboardingInviteData) {
    setFormData((prev) => ({ ...prev, invite: data }))
  }

  return {
    currentStep,
    totalSteps,
    formData,
    isSubmitting,
    setIsSubmitting,
    goNext,
    goPrev,
    setProfileData,
    setWorkspaceData,
    setInviteData,
  }
}

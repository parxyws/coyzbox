import { api } from "./client"
import type { APIResponse } from "@/types"

export interface CreateWorkspaceRequest {
  name: string
  slug: string
  description?: string
}

export interface InviteMembersRequest {
  emails: string[]
}

export async function createWorkspace(
  data: CreateWorkspaceRequest
): Promise<APIResponse<null>> {
  return api.post<APIResponse<null>>("/workspace", data)
}

export async function inviteMembers(
  data: InviteMembersRequest
): Promise<APIResponse<null>> {
  return api.post<APIResponse<null>>("/workspace/invite", data)
}

export async function completeOnboarding(): Promise<APIResponse<null>> {
  return api.post<APIResponse<null>>("/auth/onboarding/complete")
}

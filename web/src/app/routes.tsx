import { Routes, Route, Navigate } from "react-router-dom"
import RegisterPage from "@/pages/RegisterPage.tsx"
import LoginPage from "@/pages/LoginPage.tsx"
import OnboardingPage from "@/pages/OnboardingPage.tsx"

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/onboarding" element={<OnboardingPage />} />
      <Route path="/" element={<Navigate to="/login" replace />} />
    </Routes>
  )
}
